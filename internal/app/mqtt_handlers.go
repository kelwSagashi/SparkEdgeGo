package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	apperuntime "github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/system"
)

func (a *App) registerMqttCommandHandlers() {
	if a == nil || a.MQTT == nil {
		return
	}
	a.MQTT.RegisterHandler("get_stats", func(context.Context, map[string]any) (map[string]any, error) {
		return system.CollectStats(), nil
	})
	a.MQTT.RegisterHandler("run_script", a.handleMqttRunScript)
	a.MQTT.RegisterHandler("CONFIG", a.handleMqttConfig)
	a.MQTT.RegisterHandler("restart", a.handleMqttRestart)
	a.MQTT.RegisterHandler("REBOOT", a.handleMqttRestart)
}

func (a *App) handleMqttRunScript(ctx context.Context, payload map[string]any) (map[string]any, error) {
	instanceID := firstString(payload, "script_name", "instance_id", "instanceId")
	if instanceID == "" {
		return nil, errors.New("missing script_name or instance_id in payload")
	}
	_ = a.MQTT.PublishLog(ctx, "info", "Executing script: "+instanceID)
	execution, result, err := a.triggerInstance(ctx, instanceID, mapValue(payload, "input", "inputs", "data"), domain.TriggerManual)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}
	if result.Status != domain.ExecutionSuccess {
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		return nil, errors.New("script execution failed")
	}
	return map[string]any{"execution_id": execution.ID, "status": result.Status, "message": "Script triggered successfully"}, nil
}

func (a *App) handleMqttConfig(ctx context.Context, payload map[string]any) (map[string]any, error) {
	config := payload
	if nested, ok := payload["data"].(map[string]any); ok {
		config = nested
	}
	if len(config) == 0 {
		return nil, errors.New("missing configuration data")
	}
	if _, err := a.Edge.UpsertConfigMap(ctx, config); err != nil {
		return nil, err
	}
	_ = a.MQTT.PublishLog(ctx, "info", "Local configuration updated via MQTT")
	return map[string]any{"message": "Configuration synchronized"}, nil
}

func (a *App) handleMqttRestart(ctx context.Context, _ map[string]any) (map[string]any, error) {
	_ = a.MQTT.PublishLog(ctx, "warn", "Edge rebooting via remote command...")
	if strings.EqualFold(os.Getenv("SPARKEDGE_ALLOW_RESTART"), "true") {
		go func() {
			time.Sleep(2 * time.Second)
			command, args := restartCommand()
			_ = exec.Command(command, args...).Start()
		}()
		return map[string]any{"message": "Reboot initiated"}, nil
	}
	return map[string]any{"message": "Reboot command acknowledged but not executed; set SPARKEDGE_ALLOW_RESTART=true to enable system restart"}, nil
}

func (a *App) triggerInstance(ctx context.Context, instanceID string, input map[string]any, triggerType domain.TriggerType) (domain.InstanceExecution, apperuntime.TriggerResult, error) {
	instanceWithDestinations, err := a.Instances.GetWithDestinations(ctx, instanceID)
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}
	instance := instanceWithDestinations.Instance
	script, err := a.Scripts.FindByID(ctx, instance.ScriptID)
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}

	startedAt := time.Now().UTC()
	execution, err := a.Executions.Create(ctx, sqlite.CreateInstanceExecutionParams{
		InstanceID:  instance.ID,
		Status:      domain.ExecutionRunning,
		TriggerType: triggerType,
		StartedAt:   &startedAt,
		Logs:        []domain.ExecutionLog{{Level: "info", Message: "Queued MQTT trigger", Timestamp: startedAt}},
	})
	if err != nil {
		return domain.InstanceExecution{}, apperuntime.TriggerResult{}, err
	}

	_, _ = a.Instances.UpdateStatus(ctx, instance.ID, domain.InstanceStatusRunning)
	result, runErr := a.Runtime.Trigger(ctx, apperuntime.TriggerRequest{
		ExecutionID:  execution.ID,
		Instance:     instance,
		Script:       script,
		Destinations: instanceWithDestinations.Destinations,
		Trigger:      triggerType,
		Input:        input,
	})
	updated, updateErr := a.finishExecution(ctx, execution.ID, instance.ID, startedAt, result)
	if updateErr != nil {
		return updated, result, updateErr
	}
	if runErr != nil {
		return updated, result, runErr
	}
	return updated, result, nil
}

func (a *App) finishExecution(ctx context.Context, executionID string, instanceID string, startedAt time.Time, result apperuntime.TriggerResult) (domain.InstanceExecution, error) {
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
	if len(logs) == 0 {
		logs = []domain.ExecutionLog{{Level: "info", Message: "Execution finished", Timestamp: finishedAt}}
	}
	updated, err := a.Executions.UpdateStatus(ctx, executionID, sqlite.UpdateInstanceExecutionStatusParams{
		Status:          result.Status,
		FinishedAt:      &finishedAt,
		DurationMS:      &duration,
		ErrorMessage:    &errorMessage,
		Output:          &output,
		DestinationSent: &destinationSent,
		FallbackUsed:    &fallbackUsed,
		Logs:            &logs,
	})
	if result.Status == domain.ExecutionSuccess {
		_, _ = a.Instances.UpdateStatus(ctx, instanceID, domain.InstanceStatusIdle)
	} else {
		_, _ = a.Instances.UpdateStatus(ctx, instanceID, domain.InstanceStatusError)
	}
	return updated, err
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mapValue(payload map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if value, ok := payload[key].(map[string]any); ok {
			return value
		}
	}
	return map[string]any{}
}

func restartCommand() (string, []string) {
	if goruntime.GOOS == "windows" {
		return "shutdown", []string{"/r", "/t", "5"}
	}
	return "sudo", []string{"reboot"}
}
