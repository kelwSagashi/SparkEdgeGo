package runtime

import (
	"context"
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
}

type Runner struct {
	deps Dependencies
}

func NewRunner(deps Dependencies) *Runner {
	return &Runner{deps: deps}
}

type TriggerRequest struct {
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

		providerKey := strings.TrimSpace(destination.ResourceOperationID)
		providerConfig := providers.Config{Operation: map[string]any{"resource_operation_id": destination.ResourceOperationID}}
		if r.deps.ResourceOperations != nil {
			target, err := r.deps.ResourceOperations.ResolveTarget(ctx, destination.ResourceOperationID)
			if err != nil {
				failures = append(failures, destination.ID+": "+err.Error())
				continue
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
			failures = append(failures, destination.ID+": provider not configured")
			continue
		}
		adapter, ok, err := r.deps.Providers.Create(providerKey, providerConfig)
		if err != nil {
			failures = append(failures, destination.ID+": "+err.Error())
			continue
		}
		if !ok {
			failures = append(failures, destination.ID+": provider "+providerKey+" not registered")
			continue
		}
		if err := adapter.Send(ctx, payload); err != nil {
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

	payload := map[string]any{}
	if mapping != nil {
		for key, value := range mapping.PayloadTemplate {
			payload[key] = resolveValue(value, context)
		}
		for target, source := range mapping.Mapping {
			if sourceText, ok := source.(string); ok {
				payload[target] = resolveString(sourceText, context)
			} else {
				payload[target] = source
			}
		}
		for _, field := range mapping.CustomFields {
			if field.Key != "" {
				payload[field.Key] = resolveString(field.Value, context)
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

func resolveValue(value any, context map[string]any) any {
	if text, ok := value.(string); ok {
		return resolveString(text, context)
	}
	return value
}

func resolveString(value string, context map[string]any) any {
	if strings.HasPrefix(value, "$.") {
		return lookupPath(context, strings.TrimPrefix(value, "$."))
	}
	if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "{{"), "}}"))
		return lookupPath(context, key)
	}
	return value
}

func lookupPath(context map[string]any, pathValue string) any {
	var current any = context
	for _, part := range strings.Split(pathValue, ".") {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[part]
		case domain.Instance:
			current = instanceContext(typed)[part]
		case domain.DownloadedScript:
			current = scriptContext(typed)[part]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}
	return current
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
