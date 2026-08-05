package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleInstancesList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstancesActiveList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListActive(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstancesByProjectList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListByProject(r.Context(), r.PathValue("project_id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstanceGet(r *http.Request) (any, error) {
	result, err := s.deps.Instances.GetWithDestinations(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{
		"data": map[string]any{
			"instance":     publicInstance(result.Instance),
			"destinations": publicDestinationsWithMappings(result.Destinations),
		},
		"error": nil,
	}, nil
}

func (s *Server) handleInstanceCreate(r *http.Request) (any, error) {
	var req instances.Payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified {
		req.CreatedBy = identity.UserID
	}

	instance, err := s.deps.Instances.Create(r.Context(), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceUpdate(r *http.Request) (any, error) {
	var req instances.Payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceDelete(r *http.Request) (any, error) {
	if err := s.deps.Instances.Delete(r.Context(), r.PathValue("id")); err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleInstanceTrigger(r *http.Request) (any, error) {
	var req instanceTriggerRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
	}

	instance, err := s.deps.Instances.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	script, err := s.deps.Scripts.FindByID(r.Context(), instance.ScriptID)
	if err != nil {
		return scriptError(err)
	}

	triggerType := req.TriggerType
	if triggerType == "" {
		triggerType = domain.TriggerManual
	}
	startedAt := time.Now().UTC()
	execution, err := s.deps.Executions.Create(r.Context(), sqlite.CreateInstanceExecutionParams{
		InstanceID:  instance.ID,
		Status:      domain.ExecutionRunning,
		TriggerType: triggerType,
		StartedAt:   &startedAt,
		Logs: []domain.ExecutionLog{
			{Level: "info", Message: "Queued manual trigger", Timestamp: startedAt},
		},
	})
	if err != nil {
		return executionError(err)
	}

	_, _ = s.deps.Instances.UpdateStatus(r.Context(), instance.ID, domain.InstanceStatusRunning)
	result, runErr := s.deps.Runtime.Trigger(r.Context(), runtimeTriggerRequest(instance, script, triggerType, req.Input))

	finishedAt := result.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	duration := result.DurationMS
	if duration == 0 {
		duration = int(finishedAt.Sub(startedAt).Milliseconds())
	}
	output := result.RawOutput
	if output == "" && result.Output != nil {
		if data, err := json.Marshal(result.Output); err == nil {
			output = string(data)
		}
	}
	errorMessage := result.Error
	destinationSent := false
	fallbackUsed := false
	logs := result.Logs
	if len(logs) == 0 {
		logs = []domain.ExecutionLog{{Level: "info", Message: "Execution finished", Timestamp: finishedAt}}
	}

	updatedExecution, updateErr := s.deps.Executions.UpdateStatus(r.Context(), execution.ID, sqlite.UpdateInstanceExecutionStatusParams{
		Status:          result.Status,
		FinishedAt:      &finishedAt,
		DurationMS:      &duration,
		ErrorMessage:    &errorMessage,
		Output:          &output,
		DestinationSent: &destinationSent,
		FallbackUsed:    &fallbackUsed,
		Logs:            &logs,
	})
	if updateErr != nil {
		return executionError(updateErr)
	}
	if result.Status == domain.ExecutionSuccess {
		_, _ = s.deps.Instances.UpdateStatus(r.Context(), instance.ID, domain.InstanceStatusIdle)
	} else {
		_, _ = s.deps.Instances.UpdateStatus(r.Context(), instance.ID, domain.InstanceStatusError)
	}

	body := map[string]any{
		"data": map[string]any{
			"execution": publicExecution(updatedExecution),
			"result": map[string]any{
				"status": result.Status,
				"output": result.Output,
			},
		},
		"error": nil,
	}
	if runErr != nil || result.Status != domain.ExecutionSuccess {
		body["error"] = errorMessage
	}
	return body, nil
}

type instanceTriggerRequest struct {
	Input       map[string]any     `json:"input"`
	TriggerType domain.TriggerType `json:"trigger_type"`
}

func runtimeTriggerRequest(instance domain.Instance, script domain.DownloadedScript, triggerType domain.TriggerType, input map[string]any) runtime.TriggerRequest {
	if input == nil {
		input = map[string]any{}
	}
	return runtime.TriggerRequest{
		Instance: instance,
		Script:   script,
		Trigger:  triggerType,
		Input:    input,
	}
}

func (s *Server) handleInstanceTriggerPlaceholder(r *http.Request) (any, error) {
	return map[string]any{
		"data": map[string]any{
			"status": "queued",
			"note":   "runtime trigger will be connected after destinations and executions are migrated",
		},
	}, nil
}

func (s *Server) handleInstanceExecutionsList(r *http.Request) (any, error) {
	items, err := s.deps.Executions.ListByInstance(r.Context(), r.PathValue("id"), limitFromQuery(r, 50))
	if err != nil {
		return executionError(err)
	}
	return map[string]any{"data": publicExecutions(items), "error": nil}, nil
}

func instanceError(err error) (any, error) {
	if errors.Is(err, instances.ErrInvalidInstance) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid instance")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Instance not found"}, nil
	}
	return nil, err
}

func publicInstances(items []domain.Instance) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicInstance(item))
	}
	return result
}

func publicInstance(instance domain.Instance) map[string]any {
	return map[string]any{
		"id":                              instance.ID,
		"name":                            instance.Name,
		"description":                     instance.Description,
		"tags":                            instance.Tags,
		"status":                          instance.Status,
		"active":                          instance.Active,
		"project_id":                      instance.ProjectID,
		"device_id":                       instance.DeviceID,
		"script_id":                       instance.ScriptID,
		"include_device_data":             instance.IncludeDeviceData,
		"script_parameters":               instance.ScriptParameters,
		"trigger_type":                    instance.TriggerType,
		"trigger_config":                  instance.TriggerConfig,
		"fallback_enabled":                instance.FallbackEnabled,
		"fallback_strategy":               instance.FallbackStrategy,
		"fallback_retry_interval_seconds": instance.FallbackRetryIntervalSeconds,
		"on_error_action":                 instance.OnErrorAction,
		"on_error_config":                 instance.OnErrorConfig,
		"created_by":                      instance.CreatedBy,
		"created_at":                      instance.CreatedAt,
		"updated_at":                      instance.UpdatedAt,
	}
}

func publicDestinationsWithMappings(items []domain.InstanceDestinationWithMapping) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{
			"destination": publicInstanceDestination(item.Destination),
			"mapping":     nil,
		}
		if item.Mapping != nil {
			entry["mapping"] = publicDataMapping(*item.Mapping)
		}
		result = append(result, entry)
	}
	return result
}

func publicInstanceDestination(destination domain.InstanceDestination) map[string]any {
	return map[string]any{
		"id":                    destination.ID,
		"instance_id":           destination.InstanceID,
		"resource_operation_id": destination.ResourceOperationID,
		"enabled":               destination.Enabled,
		"priority":              destination.Priority,
		"retry_policy": map[string]any{
			"max_retries":    destination.RetryPolicy.MaxRetries,
			"retry_interval": destination.RetryPolicy.RetryInterval,
		},
		"created_at": destination.CreatedAt,
	}
}

func publicDataMapping(mapping domain.DataMapping) map[string]any {
	return map[string]any{
		"id":                      mapping.ID,
		"instance_destination_id": mapping.InstanceDestinationID,
		"mapping":                 mapping.Mapping,
		"payload_template":        mapping.PayloadTemplate,
		"custom_fields":           mapping.CustomFields,
		"transform_script":        mapping.TransformScript,
		"created_at":              mapping.CreatedAt,
	}
}
