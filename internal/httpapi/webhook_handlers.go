package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleWebhookReceive(r *http.Request) (any, error) {
	instanceID := r.PathValue("instanceId")
	instanceWithDestinations, err := s.deps.Instances.GetWithDestinations(r.Context(), instanceID)
	if err != nil {
		return instanceError(err)
	}
	instance := instanceWithDestinations.Instance
	if instance.TriggerType != domain.TriggerWebhook && instance.TriggerType != domain.TriggerIntervalAndWebhook {
		return map[string]any{"error": "Instance does not accept webhooks", "data": nil}, nil
	}
	if expected, ok := stringFromMap(instance.TriggerConfig, "webhook_secret", "webhookSecret"); ok && expected != "" {
		received := r.Header.Get("x-webhook-secret")
		if received == "" {
			received = r.URL.Query().Get("secret")
		}
		if received != expected {
			return map[string]any{"error": "Invalid webhook secret", "data": nil}, nil
		}
	}
	input, err := webhookInput(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	script, err := s.deps.Scripts.FindByID(r.Context(), instance.ScriptID)
	if err != nil {
		return scriptError(err)
	}
	startedAt := time.Now().UTC()
	execution, err := s.deps.Executions.Create(r.Context(), sqlite.CreateInstanceExecutionParams{
		InstanceID:  instance.ID,
		Status:      domain.ExecutionRunning,
		TriggerType: domain.TriggerWebhook,
		StartedAt:   &startedAt,
		Logs:        []domain.ExecutionLog{{Level: "info", Message: "Queued webhook trigger", Timestamp: startedAt}},
	})
	if err != nil {
		return executionError(err)
	}
	_, _ = s.deps.Instances.UpdateStatus(r.Context(), instance.ID, domain.InstanceStatusRunning)
	runtimeReq := runtimeTriggerRequest(instance, script, instanceWithDestinations.Destinations, domain.TriggerWebhook, input)
	runtimeReq.ExecutionID = execution.ID
	result, runErr := s.deps.Runtime.Trigger(r.Context(), runtimeReq)

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
	destinationSent := result.DestinationSent
	fallbackUsed := result.FallbackUsed
	logs := result.Logs
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
	response := map[string]any{"data": map[string]any{"triggered": true, "execution": publicExecution(updatedExecution)}, "error": nil}
	if runErr != nil || result.Status != domain.ExecutionSuccess {
		response["error"] = errorMessage
	}
	return response, nil
}

func webhookInput(r *http.Request) (map[string]any, error) {
	if r.Body == nil {
		return map[string]any{}, nil
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if nested, ok := body["input"].(map[string]any); ok {
		return nested, nil
	}
	return body, nil
}

func stringFromMap(values map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			return value, true
		}
	}
	return "", false
}
