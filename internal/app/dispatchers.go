package app

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
)

type DispatchResult struct {
	InstanceID  string                 `json:"instance_id"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	Status      domain.ExecutionStatus `json:"status"`
	Error       string                 `json:"error,omitempty"`
	TriggerType domain.TriggerType     `json:"trigger_type"`
}

func (a *App) DispatchEvent(ctx context.Context, eventName string, payload map[string]any) ([]DispatchResult, error) {
	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]DispatchResult, 0)
	for _, instance := range instances {
		if instance.TriggerType != domain.TriggerEvent {
			continue
		}
		configuredEvent, _ := stringFromAnyMap(instance.TriggerConfig, "event_name", "eventName")
		if strings.TrimSpace(configuredEvent) == "" || configuredEvent != strings.TrimSpace(eventName) {
			continue
		}
		input := cloneMap(payload)
		input["event_name"] = eventName
		execution, result, runErr := a.triggerInstance(ctx, instance.ID, input, domain.TriggerEvent)
		entry := DispatchResult{
			InstanceID:  instance.ID,
			TriggerType: domain.TriggerEvent,
			Status:      result.Status,
			Error:       result.Error,
		}
		if execution.ID != "" {
			entry.ExecutionID = execution.ID
		}
		if runErr != nil && entry.Error == "" {
			entry.Error = runErr.Error()
		}
		results = append(results, entry)
	}
	return results, nil
}

func (a *App) DispatchStateChange(ctx context.Context, payload map[string]any) ([]DispatchResult, error) {
	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]DispatchResult, 0)
	resolver := runtime.TemplateResolver{}
	stateContext := map[string]any{
		"state": payload,
		"input": payload,
	}
	for _, instance := range instances {
		if instance.TriggerType != domain.TriggerStateChange {
			continue
		}
		fieldPath, _ := stringFromAnyMap(instance.TriggerConfig, "state_field", "stateField")
		expectedValue, _ := stringFromAnyMap(instance.TriggerConfig, "state_equals", "stateEquals")
		if strings.TrimSpace(fieldPath) == "" {
			continue
		}
		actualValue := resolver.ResolvePath(fieldPath, stateContext)
		if expectedValue != "" && strings.TrimSpace(runtimeValueString(actualValue)) != strings.TrimSpace(expectedValue) {
			continue
		}
		input := cloneMap(payload)
		execution, result, runErr := a.triggerInstance(ctx, instance.ID, input, domain.TriggerStateChange)
		entry := DispatchResult{
			InstanceID:  instance.ID,
			TriggerType: domain.TriggerStateChange,
			Status:      result.Status,
			Error:       result.Error,
		}
		if execution.ID != "" {
			entry.ExecutionID = execution.ID
		}
		if runErr != nil && entry.Error == "" {
			entry.Error = runErr.Error()
		}
		results = append(results, entry)
	}
	return results, nil
}

func (a *App) SyncMQTTTriggerSubscriptions(ctx context.Context) error {
	if a == nil || a.MQTT == nil || !a.MQTT.IsConnected() {
		return nil
	}
	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		return err
	}

	grouped := make(map[string][]string)
	for _, instance := range instances {
		if instance.TriggerType != domain.TriggerMQTT {
			continue
		}
		topic, _ := stringFromAnyMap(instance.TriggerConfig, "mqtt_topic", "mqttTopic")
		if strings.TrimSpace(topic) == "" {
			continue
		}
		grouped[topic] = append(grouped[topic], instance.ID)
	}

	handlers := make(map[string]mqtt.TopicHandler)
	for topic, ids := range grouped {
		instanceIDs := append([]string{}, ids...)
		handlers[topic] = func(handlerCtx context.Context, receivedTopic string, raw []byte) {
			input := parseMQTTPayload(raw)
			input["mqtt_topic"] = receivedTopic
			var wg sync.WaitGroup
			for _, instanceID := range instanceIDs {
				wg.Add(1)
				go func(instanceID string) {
					defer wg.Done()
					_, _, _ = a.triggerInstance(handlerCtx, instanceID, cloneMap(input), domain.TriggerMQTT)
				}(instanceID)
			}
			wg.Wait()
		}
	}

	return a.MQTT.SyncTopicHandlers(ctx, handlers)
}

func parseMQTTPayload(raw []byte) map[string]any {
	result := map[string]any{
		"raw": string(raw),
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err == nil {
		for key, value := range parsed {
			result[key] = value
		}
	}
	return result
}

func stringFromAnyMap(values map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := values[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), true
		}
	}
	return "", false
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func runtimeValueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return strings.Trim(string(encoded), `"`)
	}
}
