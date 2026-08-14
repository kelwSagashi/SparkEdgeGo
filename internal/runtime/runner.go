package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"sync"
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
	CircuitBreakers interface {
		GetByDestination(context.Context, string) (domain.CircuitBreakerState, error)
		Upsert(context.Context, domain.CircuitBreakerState) (domain.CircuitBreakerState, error)
		Delete(context.Context, string) error
	}
	Devices interface {
		FindByID(context.Context, string) (domain.Device, error)
	}
	EdgeConfig interface {
		GetEdgeConfig(context.Context) (domain.EdgeConfig, error)
	}
}

type Runner struct {
	deps      Dependencies
	breakerMu sync.Mutex
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
	ExecutionID        string
	Status             domain.ExecutionStatus
	InputPayload       map[string]any
	Output             map[string]any
	MappedPayloads     []MappedPayload
	DestinationDetails []domain.ExecutionDestinationDetail
	RawOutput          string
	Error              string
	Logs               []domain.ExecutionLog
	DurationMS         int
	DestinationSent    bool
	FallbackUsed       bool
	StartedAt          time.Time
	FinishedAt         time.Time
}

type MappedPayload struct {
	DestinationID       string         `json:"destination_id"`
	ResourceOperationID string         `json:"resource_operation_id"`
	Payload             map[string]any `json:"payload"`
}

type destinationMetadata struct {
	serverName    string
	resourceName  string
	operationName string
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
	input = r.resolveScriptInput(ctx, req, input)
	result.InputPayload = input

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
	result.DestinationDetails = delivery.details
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
	details  []domain.ExecutionDestinationDetail
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

	result := deliveryResult{payloads: []MappedPayload{}, details: []domain.ExecutionDestinationDetail{}}
	var failures []string
	destinations := append([]domain.InstanceDestinationWithMapping{}, req.Destinations...)
	sort.SliceStable(destinations, func(i, j int) bool {
		return destinations[i].Destination.Priority < destinations[j].Destination.Priority
	})

	process := func(item domain.InstanceDestinationWithMapping) (domain.ExecutionDestinationDetail, *MappedPayload, []domain.ExecutionLog, bool, bool, error) {
		destination := item.Destination
		now := time.Now().UTC()
		metadata := r.resolveDestinationMetadata(ctx, destination)
		if !destination.Enabled {
			return domain.ExecutionDestinationDetail{
				DestinationID:       destination.ID,
				ResourceOperationID: destination.ResourceOperationID,
				ServerName:          metadata.serverName,
				ResourceName:        metadata.resourceName,
				OperationName:       metadata.operationName,
				Status:              "skipped",
				Payload:             map[string]any{},
				Timestamp:           now,
			}, nil, []domain.ExecutionLog{newLog("info", "Destination "+destination.ID+" disabled, skipping", now)}, false, false, nil
		}
		payload, mapErr := r.applyMapping(ctx, item.Mapping, output, req)
		if mapErr != nil {
			return domain.ExecutionDestinationDetail{
				DestinationID:       destination.ID,
				ResourceOperationID: destination.ResourceOperationID,
				ServerName:          metadata.serverName,
				ResourceName:        metadata.resourceName,
				OperationName:       metadata.operationName,
				Status:              "failed",
				Payload:             map[string]any{},
				Error:               mapErr.Error(),
				Timestamp:           now,
			}, nil, []domain.ExecutionLog{newLog("error", "Failed to map destination "+destination.ID+": "+mapErr.Error(), now)}, false, false, mapErr
		}
		mapped := &MappedPayload{
			DestinationID:       destination.ID,
			ResourceOperationID: destination.ResourceOperationID,
			Payload:             payload,
		}

		if err := r.sendToDestinationWithPolicy(ctx, destination, payload); err != nil {
			if fallbackLogs, ok := r.enqueueFallback(ctx, req, destination.ID, payload, err.Error()); ok {
				return domain.ExecutionDestinationDetail{
					DestinationID:       destination.ID,
					ResourceOperationID: destination.ResourceOperationID,
					ServerName:          metadata.serverName,
					ResourceName:        metadata.resourceName,
					OperationName:       metadata.operationName,
					Status:              "fallback",
					Payload:             payload,
					Error:               err.Error(),
					UsedFallback:        true,
					Timestamp:           time.Now().UTC(),
				}, mapped, fallbackLogs, false, true, nil
			}
			return domain.ExecutionDestinationDetail{
				DestinationID:       destination.ID,
				ResourceOperationID: destination.ResourceOperationID,
				ServerName:          metadata.serverName,
				ResourceName:        metadata.resourceName,
				OperationName:       metadata.operationName,
				Status:              "failed",
				Payload:             payload,
				Error:               err.Error(),
				Timestamp:           time.Now().UTC(),
			}, mapped, nil, false, false, err
		}
		successAt := time.Now().UTC()
		return domain.ExecutionDestinationDetail{
			DestinationID:       destination.ID,
			ResourceOperationID: destination.ResourceOperationID,
			ServerName:          metadata.serverName,
			ResourceName:        metadata.resourceName,
			OperationName:       metadata.operationName,
			Status:              "success",
			Payload:             payload,
			Timestamp:           successAt,
		}, mapped, []domain.ExecutionLog{newLog("info", "Successfully dispatched to destination "+destination.ID, successAt)}, true, false, nil
	}

	if req.Instance.ExecutionMode == domain.ExecutionModeParallel {
		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, item := range destinations {
			wg.Add(1)
			go func(item domain.InstanceDestinationWithMapping) {
				defer wg.Done()
				detail, mapped, logs, sent, fallbackUsed, err := process(item)
				mu.Lock()
				defer mu.Unlock()
				if mapped != nil {
					result.payloads = append(result.payloads, *mapped)
				}
				result.details = append(result.details, detail)
				result.logs = append(result.logs, logs...)
				if sent {
					result.sent = true
				}
				if fallbackUsed {
					result.fallback = true
				}
				if err != nil {
					failures = append(failures, item.Destination.ID+": "+err.Error())
				}
			}(item)
		}
		wg.Wait()
	} else {
		for _, item := range destinations {
			detail, mapped, logs, sent, fallbackUsed, err := process(item)
			if mapped != nil {
				result.payloads = append(result.payloads, *mapped)
			}
			result.details = append(result.details, detail)
			result.logs = append(result.logs, logs...)
			if sent {
				result.sent = true
			}
			if fallbackUsed {
				result.fallback = true
			}
			if err != nil {
				failures = append(failures, item.Destination.ID+": "+err.Error())
				if !item.Destination.RetryPolicy.ContinueOnError && item.Destination.RetryPolicy.IsolationMode != "continue" {
					break
				}
			}
		}
	}
	if len(failures) > 0 && !result.sent {
		return result, errors.New(strings.Join(failures, "; "))
	}
	if result.sent {
		result.logs = append(result.logs, newLog("info", "Destination dispatch completed", time.Now().UTC()))
	}
	return result, nil
}

func (r *Runner) sendToDestinationWithPolicy(ctx context.Context, destination domain.InstanceDestination, payload map[string]any) error {
	policy := destination.RetryPolicy
	if policy.MaxRetries <= 0 {
		policy.MaxRetries = 1
	}
	if policy.RetryInterval <= 0 {
		policy.RetryInterval = 1
	}
	if policy.TimeoutSeconds <= 0 {
		policy.TimeoutSeconds = 30
	}
	if err := r.beforeSend(destination.ID, policy); err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= policy.MaxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(policy.TimeoutSeconds)*time.Second)
		err := r.sendToDestination(attemptCtx, destination, payload)
		cancel()
		if err == nil {
			r.afterSendSuccess(destination.ID)
			return nil
		}
		lastErr = err
		if attempt < policy.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(policy.RetryInterval) * time.Second):
			}
		}
	}
	r.afterSendFailure(destination.ID, policy)
	return lastErr
}

func (r *Runner) beforeSend(destinationID string, policy domain.RetryPolicy) error {
	if policy.CircuitBreakerThreshold <= 0 || policy.CircuitBreakerCooldownSeconds <= 0 {
		return nil
	}
	r.breakerMu.Lock()
	defer r.breakerMu.Unlock()
	state, err := r.loadBreakerState(destinationID)
	if err != nil {
		return err
	}
	if state.OpenedUntil != nil && time.Now().UTC().Before(*state.OpenedUntil) {
		return errors.New("circuit breaker open for destination")
	}
	if state.OpenedUntil != nil && time.Now().UTC().After(*state.OpenedUntil) {
		state.OpenedUntil = nil
		_ = r.saveBreakerState(state)
	}
	return nil
}

func (r *Runner) afterSendSuccess(destinationID string) {
	r.breakerMu.Lock()
	defer r.breakerMu.Unlock()
	if r.deps.CircuitBreakers != nil {
		_ = r.deps.CircuitBreakers.Delete(context.Background(), destinationID)
	}
}

func (r *Runner) afterSendFailure(destinationID string, policy domain.RetryPolicy) {
	if policy.CircuitBreakerThreshold <= 0 || policy.CircuitBreakerCooldownSeconds <= 0 {
		return
	}
	r.breakerMu.Lock()
	defer r.breakerMu.Unlock()
	state, err := r.loadBreakerState(destinationID)
	if err != nil {
		return
	}
	state.ConsecutiveFailures++
	if state.ConsecutiveFailures >= policy.CircuitBreakerThreshold {
		openedUntil := time.Now().UTC().Add(time.Duration(policy.CircuitBreakerCooldownSeconds) * time.Second)
		state.OpenedUntil = &openedUntil
		state.ConsecutiveFailures = 0
	}
	_ = r.saveBreakerState(state)
}

func (r *Runner) loadBreakerState(destinationID string) (domain.CircuitBreakerState, error) {
	if r.deps.CircuitBreakers == nil {
		return domain.CircuitBreakerState{DestinationID: destinationID}, nil
	}
	state, err := r.deps.CircuitBreakers.GetByDestination(context.Background(), destinationID)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return domain.CircuitBreakerState{DestinationID: destinationID}, nil
		}
		return domain.CircuitBreakerState{}, err
	}
	return state, nil
}

func (r *Runner) saveBreakerState(state domain.CircuitBreakerState) error {
	if r.deps.CircuitBreakers == nil {
		return nil
	}
	state.DestinationID = strings.TrimSpace(state.DestinationID)
	if state.DestinationID == "" {
		return nil
	}
	_, err := r.deps.CircuitBreakers.Upsert(context.Background(), state)
	return err
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

func (r *Runner) resolveDestinationMetadata(ctx context.Context, destination domain.InstanceDestination) destinationMetadata {
	if r == nil || r.deps.ResourceOperations == nil {
		return destinationMetadata{}
	}
	target, err := r.deps.ResourceOperations.ResolveTarget(ctx, destination.ResourceOperationID)
	if err != nil {
		return destinationMetadata{}
	}
	return destinationMetadata{
		serverName:    target.Server.Name,
		resourceName:  target.Resource.Name,
		operationName: target.Operation.Name,
	}
}

func (r *Runner) enqueueFallback(ctx context.Context, req TriggerRequest, destinationID string, payload map[string]any, reason string) ([]domain.ExecutionLog, bool) {
	if !req.Instance.FallbackEnabled || r.deps.Fallback == nil {
		return nil, false
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}
	if _, err := r.deps.Fallback.Create(ctx, sqlite.CreateLocalFallbackParams{
		InstanceID:    req.Instance.ID,
		DestinationID: destinationID,
		ExecutionID:   req.ExecutionID,
		Payload:       string(encoded),
	}); err != nil {
		return nil, false
	}
	return []domain.ExecutionLog{
		newLog("info", "Enqueued destination "+destinationID+" into fallback storage: "+reason, time.Now().UTC()),
	}, true
}

func (r *Runner) applyMapping(ctx context.Context, mapping *domain.DataMapping, output map[string]any, req TriggerRequest) (map[string]any, error) {
	context := r.buildTemplateContext(ctx, req, output)
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
	if mapping != nil && strings.TrimSpace(mapping.TransformScript) != "" {
		transformed, err := applyTransformScript(mapping.TransformScript, payload, context)
		if err != nil {
			return nil, err
		}
		payload = transformed
	}
	return payload, nil
}

func (r *Runner) resolveScriptInput(ctx context.Context, req TriggerRequest, input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	context := r.buildTemplateContext(ctx, req, nil)
	resolver := TemplateResolver{}

	resolved, ok := resolver.Resolve(input, context).(map[string]any)
	if ok {
		return resolved
	}
	return input
}

func (r *Runner) buildTemplateContext(ctx context.Context, req TriggerRequest, output map[string]any) map[string]any {
	device := r.resolveDeviceContext(ctx, req.Instance.DeviceID)
	instance := instanceContext(req.Instance)
	system := r.resolveSystemContext(ctx, req.Instance)
	triggerInput := cloneMap(req.Input)
	context := map[string]any{
		"script":            output,
		"output":            output,
		"input":             triggerInput,
		"trigger_input":     triggerInput,
		"instance":          instance,
		"device":            device,
		"device_data":       device,
		"system":            system,
		"system_data":       system,
		"script_metadata":   scriptContext(req.Script),
		"script_parameters": req.Instance.ScriptParameters,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
		"trigger": map[string]any{
			"type":  req.Trigger,
			"input": triggerInput,
		},
	}
	return context
}

func (r *Runner) resolveDeviceContext(ctx context.Context, deviceID string) map[string]any {
	context := map[string]any{}
	if strings.TrimSpace(deviceID) == "" || r == nil || r.deps.Devices == nil {
		return context
	}

	device, err := r.deps.Devices.FindByID(ctx, deviceID)
	if err != nil {
		return context
	}
	return deviceContext(device)
}

func (r *Runner) resolveSystemContext(ctx context.Context, instance domain.Instance) map[string]any {
	config := domain.EdgeConfig{}
	if r != nil && r.deps.EdgeConfig != nil {
		if loaded, err := r.deps.EdgeConfig.GetEdgeConfig(ctx); err == nil {
			config = loaded
		}
	}
	return systemContext(config, instance)
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

func deviceContext(device domain.Device) map[string]any {
	others := map[string]any{}
	for _, field := range device.Others {
		if strings.TrimSpace(field.Key) == "" {
			continue
		}
		others[field.Key] = field.Value
	}

	context := map[string]any{
		"id":                    device.ID,
		"device_id":             device.DeviceID,
		"name":                  device.Name,
		"brand":                 device.Brand,
		"serial_number":         device.SerialNumber,
		"connection_method":     string(device.ConnectionMethod),
		"ip_address":            device.IPAddress,
		"location":              device.Location,
		"description":           device.Description,
		"resource_operation_id": device.ResourceOperationID,
		"others":                others,
	}
	if !device.CreatedAt.IsZero() {
		context["created_at"] = device.CreatedAt.UTC().Format(time.RFC3339)
	}
	if !device.UpdatedAt.IsZero() {
		context["updated_at"] = device.UpdatedAt.UTC().Format(time.RFC3339)
	}
	for key, value := range others {
		if _, exists := context[key]; !exists {
			context[key] = value
		}
	}
	return context
}

func systemContext(config domain.EdgeConfig, instance domain.Instance) map[string]any {
	return map[string]any{
		"id":              config.ID,
		"edge_name":       firstNonBlank(config.EdgeName, os.Getenv("SPARKEDGE_NAME")),
		"lat":             config.Lat,
		"lng":             config.Lng,
		"location_source": firstNonBlank(config.LocationSource, "manual"),
		"tags":            config.Tags,
		"os":              firstNonBlank(config.OS, os.Getenv("SPARKEDGE_OS"), goruntime.GOOS),
		"os_version":      firstNonBlank(config.OSVersion, os.Getenv("SPARKEDGE_OS_VERSION"), goruntime.GOOS+"/"+goruntime.GOARCH),
		"edge_version":    firstNonBlank(config.EdgeVersion, os.Getenv("SPARKEDGE_VERSION"), "go-dev"),
		"hardware":        firstNonBlank(config.Hardware, os.Getenv("SPARKEDGE_HARDWARE"), goruntime.GOARCH),
		"environment":     firstNonBlank(config.Environment, os.Getenv("SPARKEDGE_ENV"), "production"),
		"description":     config.Description,
		"now":             time.Now().UTC().Format(time.RFC3339),
		"instance_name":   instance.Name,
		"instance_id":     instance.ID,
		"device_id":       instance.DeviceID,
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
