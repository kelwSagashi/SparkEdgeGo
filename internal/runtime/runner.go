package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

type SparkitExecutor interface {
	Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error)
	RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error)
	Schema(ctx context.Context, scriptPath string) (map[string]any, error)
}

type Dependencies struct {
	Sparkit            SparkitExecutor
	Providers          *providers.Registry
	ResourceOperations interface {
		ResolveTarget(context.Context, string) (sqlite.OperationTarget, error)
	}
	Fallback interface {
		Create(context.Context, sqlite.CreateLocalFallbackParams) (domain.LocalFallbackItem, error)
		ListPending(context.Context) ([]domain.LocalFallbackItem, error)
		MarkAsSending(context.Context, string) (domain.LocalFallbackItem, error)
		MarkAsSent(context.Context, string) (domain.LocalFallbackItem, error)
		IncrementRetry(context.Context, string, string) (domain.LocalFallbackItem, error)
		MarkAsFailed(context.Context, string, string) (domain.LocalFallbackItem, error)
		FindByID(context.Context, string) (domain.LocalFallbackItem, error)
	}
	Destinations interface {
		FindByID(context.Context, string) (domain.InstanceDestination, error)
	}
}

type Runner struct {
	deps Dependencies
}

func NewRunner(deps Dependencies) *Runner {
	return &Runner{deps: deps}
}

type TriggerRequest struct {
	ExecutionID  string
	Instance     domain.Instance
	Script       domain.DownloadedScript
	Destinations []domain.InstanceDestinationWithMapping
	Trigger      domain.TriggerType
	Input        map[string]any
}

type TriggerResult struct {
	ExecutionID     string
	Status          domain.ExecutionStatus
	Output          map[string]any
	MappedPayloads  []MappedPayload
	RawOutput       string
	Error           string
	Logs            []domain.ExecutionLog
	DurationMS      int
	DestinationSent bool
	FallbackUsed    bool
	StartedAt       time.Time
	FinishedAt      time.Time
}

type MappedPayload struct {
	DestinationID       string         `json:"destination_id"`
	ResourceOperationID string         `json:"resource_operation_id"`
	Payload             map[string]any `json:"payload"`
}

func (r *Runner) Trigger(ctx context.Context, req TriggerRequest) (TriggerResult, error) {
	startedAt := time.Now().UTC()
	logs := []domain.ExecutionLog{newLog("info", "Starting instance execution", startedAt)}
	result := TriggerResult{
		Status:    domain.ExecutionRunning,
		Logs:      logs,
		StartedAt: startedAt,
	}

	if r == nil || r.deps.Sparkit == nil {
		return finish(result, domain.ExecutionFailed, "sparkit runtime unavailable"), errors.New("sparkit runtime unavailable")
	}
	if strings.TrimSpace(req.Script.LocalPath) == "" || strings.TrimSpace(req.Script.MainFile) == "" {
		return finish(result, domain.ExecutionFailed, "script path not configured"), errors.New("script path not configured")
	}

	input := map[string]any{}
	for key, value := range req.Instance.ScriptParameters {
		input[key] = value
	}
	for key, value := range req.Input {
		input[key] = value
	}

	logs = append(logs, newLog("info", "Running Python script with Sparkit", time.Now().UTC()))
	scriptResult, err := r.deps.Sparkit.RunFile(ctx, resolvePath(req.Script.LocalPath), req.Script.MainFile, resolvePath(req.Script.VenvPath), input)
	result.Logs = logs
	result.Output = scriptResult.Data
	result.RawOutput = scriptResult.Stdout

	if err != nil {
		return finish(result, domain.ExecutionFailed, err.Error()), err
	}
	if scriptResult.ExitCode != 0 {
		message := strings.TrimSpace(scriptResult.Stderr)
		if message == "" {
			message = "script exited with non-zero status"
		}
		return finish(result, domain.ExecutionFailed, message), nil
	}

	result.Logs = append(result.Logs, newLog("info", "Script finished successfully", time.Now().UTC()))
	delivery, deliveryErr := r.dispatchDestinations(ctx, req, result.Output)
	result.MappedPayloads = delivery.payloads
	result.DestinationSent = delivery.sent
	result.FallbackUsed = delivery.fallback
	result.Logs = append(result.Logs, delivery.logs...)
	if deliveryErr != nil {
		if delivery.fallback {
			return finish(result, domain.ExecutionSuccess, ""), nil
		}
		return finish(result, domain.ExecutionFailed, deliveryErr.Error()), nil
	}
	return finish(result, domain.ExecutionSuccess, ""), nil
}

type deliveryResult struct {
	sent     bool
	fallback bool
	payloads []MappedPayload
	logs     []domain.ExecutionLog
}

func (r *Runner) dispatchDestinations(ctx context.Context, req TriggerRequest, output map[string]any) (deliveryResult, error) {
	if err := ctx.Err(); err != nil {
		return deliveryResult{}, err
	}
	if len(req.Destinations) == 0 {
		return deliveryResult{
			sent: true,
			logs: []domain.ExecutionLog{newLog("info", "No active destinations configured for this instance", time.Now().UTC())},
		}, nil
	}

	result := deliveryResult{payloads: []MappedPayload{}}
	var failures []string
	for _, item := range req.Destinations {
		destination := item.Destination
		if !destination.Enabled {
			result.logs = append(result.logs, newLog("info", "Destination "+destination.ID+" disabled, skipping", time.Now().UTC()))
			continue
		}
		payload := applyMapping(item.Mapping, output, req)
		result.payloads = append(result.payloads, MappedPayload{
			DestinationID:       destination.ID,
			ResourceOperationID: destination.ResourceOperationID,
			Payload:             payload,
		})

		if err := r.sendToDestination(ctx, destination, payload); err != nil {
			if r.enqueueFallback(ctx, req, destination.ID, payload, err.Error(), &result) {
				continue
			}
			failures = append(failures, destination.ID+": "+err.Error())
			continue
		}
		result.sent = true
		result.logs = append(result.logs, newLog("info", "Successfully dispatched to destination "+destination.ID, time.Now().UTC()))
	}
	if len(failures) > 0 && !result.sent {
		return result, errors.New(strings.Join(failures, "; "))
	}
	if result.sent {
		result.logs = append(result.logs, newLog("info", "Destination dispatch completed", time.Now().UTC()))
	}
	return result, nil
}

func (r *Runner) FlushFallback(ctx context.Context, maxRetries int) (int, error) {
	if r == nil || r.deps.Fallback == nil {
		return 0, nil
	}
	items, err := r.deps.Fallback.ListPending(ctx)
	if err != nil {
		return 0, err
	}
	sent := 0
	var failures []string
	for _, item := range items {
		if maxRetries > 0 && item.RetryCount >= maxRetries {
			_, _ = r.deps.Fallback.MarkAsFailed(ctx, item.ID, "max retries exceeded")
			continue
		}
		if item.DestinationID == "" || r.deps.Destinations == nil {
			_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, "destination not configured")
			failures = append(failures, item.ID+": destination not configured")
			continue
		}
		destination, err := r.deps.Destinations.FindByID(ctx, item.DestinationID)
		if err != nil {
			_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, err.Error())
			failures = append(failures, item.ID+": "+err.Error())
			continue
		}
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(item.Payload), &payload); err != nil {
			_, _ = r.deps.Fallback.MarkAsFailed(ctx, item.ID, err.Error())
			failures = append(failures, item.ID+": "+err.Error())
			continue
		}
		_, _ = r.deps.Fallback.MarkAsSending(ctx, item.ID)
		if err := r.sendToDestination(ctx, destination, payload); err != nil {
			if maxRetries > 0 && item.RetryCount+1 >= maxRetries {
				_, _ = r.deps.Fallback.MarkAsFailed(ctx, item.ID, err.Error())
			} else {
				_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, err.Error())
			}
			failures = append(failures, item.ID+": "+err.Error())
			continue
		}
		_, _ = r.deps.Fallback.MarkAsSent(ctx, item.ID)
		sent++
	}
	if len(failures) > 0 {
		return sent, errors.New(strings.Join(failures, "; "))
	}
	return sent, nil
}

func (r *Runner) RetryFallbackItem(ctx context.Context, id string) (bool, error) {
	if r == nil || r.deps.Fallback == nil {
		return false, errors.New("fallback store not configured")
	}
	item, err := r.deps.Fallback.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	if item.DestinationID == "" || r.deps.Destinations == nil {
		_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, "destination not configured")
		return false, errors.New("destination not configured")
	}
	destination, err := r.deps.Destinations.FindByID(ctx, item.DestinationID)
	if err != nil {
		_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, err.Error())
		return false, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(item.Payload), &payload); err != nil {
		_, _ = r.deps.Fallback.MarkAsFailed(ctx, item.ID, err.Error())
		return false, err
	}
	_, _ = r.deps.Fallback.MarkAsSending(ctx, item.ID)
	if err := r.sendToDestination(ctx, destination, payload); err != nil {
		_, _ = r.deps.Fallback.IncrementRetry(ctx, item.ID, err.Error())
		return false, err
	}
	_, _ = r.deps.Fallback.MarkAsSent(ctx, item.ID)
	return true, nil
}

func (r *Runner) sendToDestination(ctx context.Context, destination domain.InstanceDestination, payload map[string]any) error {
	providerKey := strings.TrimSpace(destination.ResourceOperationID)
	providerConfig := providers.Config{Operation: map[string]any{"resource_operation_id": destination.ResourceOperationID}}
	if r.deps.ResourceOperations != nil {
		target, err := r.deps.ResourceOperations.ResolveTarget(ctx, destination.ResourceOperationID)
		if err != nil {
			return err
		}
		providerKey = strings.TrimSpace(target.Server.DriverKey)
		providerConfig.Server = map[string]any{"id": target.Server.ID, "name": target.Server.Name, "type": target.Server.Type, "server_type_id": target.Server.ServerTypeID, "credential_id": target.Server.CredentialID, "headers": target.Server.Headers, "project_id": target.Server.ProjectID}
		providerConfig.Resource = map[string]any{"id": target.Resource.ID, "server_id": target.Resource.ServerID, "name": target.Resource.Name, "type": target.Resource.Type, "config": target.Resource.Config}
		providerConfig.Operation = map[string]any{"id": target.Operation.ID, "resource_id": target.Operation.ResourceID, "name": target.Operation.Name, "type": target.Operation.Type, "config": target.Operation.Config, "input_schema": target.Operation.InputSchema, "output_schema": target.Operation.OutputSchema}
		if target.Credential != nil {
			providerConfig.Credentials = map[string]any{"id": target.Credential.ID, "name": target.Credential.Name, "auth_type_id": target.Credential.AuthTypeID, "data": target.Credential.Data}
			if strings.TrimSpace(target.Credential.AuthTypeID) != "" {
				providerKey = strings.TrimSpace(target.Credential.AuthTypeID)
			}
		}
		if providerKey == "http" && target.Credential == nil {
			providerKey = "no_auth"
		}
	}
	if r.deps.Providers == nil || providerKey == "" {
		return errors.New("provider not configured")
	}
	adapter, ok, err := r.deps.Providers.Create(providerKey, providerConfig)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("provider " + providerKey + " not registered")
	}
	return adapter.Send(ctx, payload)
}

func (r *Runner) enqueueFallback(ctx context.Context, req TriggerRequest, destinationID string, payload map[string]any, reason string, result *deliveryResult) bool {
	if !req.Instance.FallbackEnabled || r.deps.Fallback == nil {
		return false
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if _, err := r.deps.Fallback.Create(ctx, sqlite.CreateLocalFallbackParams{
		InstanceID:    req.Instance.ID,
		DestinationID: destinationID,
		ExecutionID:   req.ExecutionID,
		Payload:       string(encoded),
	}); err != nil {
		return false
	}
	result.fallback = true
	result.logs = append(result.logs, newLog("info", "Enqueued destination "+destinationID+" into fallback storage: "+reason, time.Now().UTC()))
	return true
}

func applyMapping(mapping *domain.DataMapping, output map[string]any, req TriggerRequest) map[string]any {
	context := map[string]any{
		"script":            output,
		"output":            output,
		"instance":          instanceContext(req.Instance),
		"script_metadata":   scriptContext(req.Script),
		"script_parameters": req.Instance.ScriptParameters,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range output {
		context[key] = value
	}

	resolver := TemplateResolver{}
	payload := map[string]any{}
	if mapping != nil {
		if resolvedTemplate, ok := resolver.Resolve(mapping.PayloadTemplate, context).(map[string]any); ok {
			for key, value := range resolvedTemplate {
				payload[key] = value
			}
		}
		for target, source := range mapping.Mapping {
			if sourceText, ok := source.(string); ok {
				if strings.HasPrefix(strings.TrimSpace(sourceText), "$") {
					if value := resolver.ResolvePath(sourceText, context); value != nil {
						payload[target] = value
					}
					continue
				}
				if value := resolver.Resolve(sourceText, context); value != nil {
					payload[target] = value
				}
			} else {
				payload[target] = source
			}
		}
		for _, field := range mapping.CustomFields {
			if field.Key != "" {
				payload[field.Key] = resolver.Resolve(field.Value, context)
			}
		}
	}
	if len(payload) == 0 {
		for key, value := range output {
			payload[key] = value
		}
	}
	return payload
}

func instanceContext(instance domain.Instance) map[string]any {
	return map[string]any{
		"id":                instance.ID,
		"name":              instance.Name,
		"project_id":        instance.ProjectID,
		"device_id":         instance.DeviceID,
		"script_id":         instance.ScriptID,
		"script_parameters": instance.ScriptParameters,
	}
}

func scriptContext(script domain.DownloadedScript) map[string]any {
	return map[string]any{
		"id":         script.ID,
		"name":       script.Name,
		"local_path": script.LocalPath,
		"main_file":  script.MainFile,
	}
}

func finish(result TriggerResult, status domain.ExecutionStatus, message string) TriggerResult {
	finishedAt := time.Now().UTC()
	result.Status = status
	result.FinishedAt = finishedAt
	result.DurationMS = int(finishedAt.Sub(result.StartedAt).Milliseconds())
	if message != "" {
		result.Error = message
		result.Logs = append(result.Logs, newLog("error", message, finishedAt))
	}
	return result
}

func newLog(level string, message string, timestamp time.Time) domain.ExecutionLog {
	return domain.ExecutionLog{Level: level, Message: message, Timestamp: timestamp}
}

func resolvePath(pathValue string) string {
	if strings.HasPrefix(pathValue, ".spark_edge") {
		home, err := os.UserHomeDir()
		if err != nil {
			return pathValue
		}
		return filepath.Join(home, pathValue)
	}
	if strings.HasPrefix(pathValue, "~/") || pathValue == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return pathValue
		}
		if pathValue == "~" {
			return home
		}
		return filepath.Join(home, strings.TrimPrefix(pathValue, "~/"))
	}
	return pathValue
}
