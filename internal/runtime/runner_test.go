package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

func TestRunnerTriggerRunsSparkitAndMergesInput(t *testing.T) {
	executor := &fakeSparkitExecutor{}
	runner := NewRunner(Dependencies{Sparkit: executor})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:               "instance-1",
			Name:             "Temperature Monitor",
			ScriptParameters: map[string]any{"host": "127.0.0.1", "port": float64(502)},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
		Destinations: []domain.InstanceDestinationWithMapping{
			{
				Destination: domain.InstanceDestination{
					ID:                  "destination-1",
					ResourceOperationID: "operation-1",
					Enabled:             true,
				},
				Mapping: &domain.DataMapping{
					Mapping:         map[string]any{"temperature": "$.temperature"},
					PayloadTemplate: map[string]any{"instance_name": "{{instance.name}}"},
				},
			},
		},
		Input: map[string]any{"port": float64(503)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionFailed {
		t.Fatalf("expected failed result without provider, got %#v", result)
	}
	if result.Error == "" {
		t.Fatalf("expected provider error message in result")
	}
	if len(result.MappedPayloads) != 1 {
		t.Fatalf("expected one mapped payload, got %#v", result.MappedPayloads)
	}
	payload := result.MappedPayloads[0].Payload
	if payload["temperature"] != float64(42) || payload["instance_name"] != "Temperature Monitor" {
		t.Fatalf("unexpected mapped payload %#v", payload)
	}
	if executor.input["host"] != "127.0.0.1" || executor.input["port"] != float64(503) {
		t.Fatalf("unexpected merged input %#v", executor.input)
	}
}

func TestRunnerTriggerSucceedsWithoutDestinations(t *testing.T) {
	executor := &fakeSparkitExecutor{}
	runner := NewRunner(Dependencies{Sparkit: executor})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:               "instance-1",
			ScriptParameters: map[string]any{"host": "127.0.0.1"},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionSuccess || result.Output["ok"] != true {
		t.Fatalf("unexpected result %#v", result)
	}
	if !result.DestinationSent {
		t.Fatalf("expected destination sent to be true when no destinations are configured")
	}
}

func TestRunnerTriggerRetriesScriptAndEventuallySucceeds(t *testing.T) {
	executor := &fakeSparkitExecutor{
		results: []domain.ScriptResult{
			{ExitCode: 1, Stderr: "modbus timeout"},
			{ExitCode: 0, Stdout: `{"ok":true}`, Data: map[string]any{"ok": true}},
		},
	}
	runner := NewRunner(Dependencies{Sparkit: executor})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:            "instance-1",
			OnErrorAction: domain.OnErrorRetry,
			OnErrorConfig: map[string]any{
				"max_retries":            2,
				"retry_interval_seconds": 1,
			},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionSuccess {
		t.Fatalf("expected success after retry, got %#v", result)
	}
	if executor.calls != 2 {
		t.Fatalf("expected 2 attempts, got %d", executor.calls)
	}
	if len(result.Logs) < 4 {
		t.Fatalf("expected retry logs to be recorded, got %#v", result.Logs)
	}
	if !hasLogCode(result.Logs, "script_attempt_result") || !hasLogCode(result.Logs, "script_attempt_retry") {
		t.Fatalf("expected structured retry logs, got %#v", result.Logs)
	}
}

func TestRunnerTriggerStopsAfterConfiguredScriptRetries(t *testing.T) {
	executor := &fakeSparkitExecutor{
		results: []domain.ScriptResult{
			{ExitCode: 1, Stderr: "connection refused"},
			{ExitCode: 1, Stderr: "connection refused"},
			{ExitCode: 1, Stderr: "connection refused"},
		},
	}
	runner := NewRunner(Dependencies{Sparkit: executor})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:            "instance-1",
			OnErrorAction: domain.OnErrorRetry,
			OnErrorConfig: map[string]any{
				"max_retries":            2,
				"retry_interval_seconds": 1,
			},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionFailed {
		t.Fatalf("expected failure after exhausting retries, got %#v", result)
	}
	if result.Error != "connection refused" {
		t.Fatalf("expected final stderr in error, got %#v", result.Error)
	}
	if executor.calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", executor.calls)
	}
}

func TestRunnerTriggerDoesNotRetryNonTransientScriptFailures(t *testing.T) {
	executor := &fakeSparkitExecutor{
		results: []domain.ScriptResult{
			{ExitCode: 1, Stderr: "TypeError: invalid payload"},
			{ExitCode: 0, Stdout: `{"ok":true}`, Data: map[string]any{"ok": true}},
		},
	}
	runner := NewRunner(Dependencies{Sparkit: executor})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:            "instance-1",
			OnErrorAction: domain.OnErrorRetry,
			OnErrorConfig: map[string]any{
				"max_retries":            3,
				"retry_interval_seconds": 1,
			},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionFailed {
		t.Fatalf("expected immediate failure for non-transient error, got %#v", result)
	}
	if executor.calls != 1 {
		t.Fatalf("expected a single attempt, got %d", executor.calls)
	}
	if !hasLogCode(result.Logs, "script_retry_skipped") {
		t.Fatalf("expected retry skipped log, got %#v", result.Logs)
	}
}

func TestRetryDelaySecondsUsesExponentialBackoffWithCap(t *testing.T) {
	policy := scriptRetryPolicy{
		attempts:           4,
		intervalSeconds:    2,
		maxIntervalSeconds: 5,
		retryScope:         "transient",
	}
	if retryDelaySeconds(policy, 1) != 2 {
		t.Fatalf("expected first delay 2s")
	}
	if retryDelaySeconds(policy, 2) != 4 {
		t.Fatalf("expected second delay 4s")
	}
	if retryDelaySeconds(policy, 3) != 5 {
		t.Fatalf("expected capped third delay 5s")
	}
}

func TestRunnerApplyMappingIncludesDeviceAndSystemContext(t *testing.T) {
	runner := NewRunner(Dependencies{
		Devices: fakeDevicesRepo{
			items: map[string]domain.Device{
				"device-1": {
					ID:               "device-1",
					DeviceID:         "edge-device-001",
					Name:             "Pump A",
					Brand:            "Spark",
					ConnectionMethod: domain.DeviceConnectionMQTT,
					Others: []domain.DeviceOtherField{
						{Key: "sector", Value: "line-7"},
					},
					CreatedAt: time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
				},
			},
		},
		EdgeConfig: fakeEdgeConfigRepo{
			config: domain.EdgeConfig{
				ID:          "edge-1",
				EdgeName:    "Factory Edge",
				Environment: "staging",
				OS:          "linux",
			},
		},
	})

	payload, err := runner.applyMapping(context.Background(), &domain.DataMapping{
		PayloadTemplate: map[string]any{
			"deviceName": "{{device.name}}",
			"deviceZone": "{{device_data.sector}}",
			"edgeName":   "{{system.edge_name}}",
			"edgeEnv":    "{{system_data.environment}}",
		},
		Mapping: map[string]any{
			"temperature": "$.temperature",
		},
	}, map[string]any{
		"temperature": float64(42),
	}, TriggerRequest{
		Instance: domain.Instance{
			DeviceID: "device-1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if payload["deviceName"] != "Pump A" {
		t.Fatalf("expected device context, got %#v", payload)
	}
	if payload["deviceZone"] != "line-7" {
		t.Fatalf("expected device other fields to be flattened, got %#v", payload)
	}
	if payload["edgeName"] != "Factory Edge" || payload["edgeEnv"] != "staging" {
		t.Fatalf("expected system context, got %#v", payload)
	}
	if payload["temperature"] != float64(42) {
		t.Fatalf("expected output mapping to remain intact, got %#v", payload)
	}
}

func TestRunnerApplyMappingSupportsTransformScript(t *testing.T) {
	runner := NewRunner(Dependencies{})

	payload, err := runner.applyMapping(context.Background(), &domain.DataMapping{
		PayloadTemplate: map[string]any{
			"temperature": "{{script.temperature}}",
			"status":      "{{script.status}}",
		},
		TransformScript: `
set meta.collected_at = now()
set summary = concat('status:', data.status)
delete status
return { ...data, status_label: summary }
`,
	}, map[string]any{
		"temperature": float64(42),
		"status":      "warning",
	}, TriggerRequest{
		Instance: domain.Instance{
			ID:   "instance-1",
			Name: "Collector 1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if payload["temperature"] != float64(42) {
		t.Fatalf("expected original payload to remain, got %#v", payload)
	}
	if _, ok := payload["status"]; ok {
		t.Fatalf("expected status to be removed by transform, got %#v", payload)
	}
	if payload["status_label"] != "status:warning" {
		t.Fatalf("expected derived label, got %#v", payload)
	}
	meta, ok := payload["meta"].(map[string]any)
	if !ok || meta["collected_at"] == "" {
		t.Fatalf("expected nested meta timestamp, got %#v", payload)
	}
}

func TestRunnerResolveScriptInputSupportsTemplates(t *testing.T) {
	runner := NewRunner(Dependencies{
		Devices: fakeDevicesRepo{
			items: map[string]domain.Device{
				"device-1": {
					ID:       "device-1",
					DeviceID: "edge-device-001",
					Name:     "Pump A",
					Others: []domain.DeviceOtherField{
						{Key: "sector", Value: "line-7"},
					},
				},
			},
		},
		EdgeConfig: fakeEdgeConfigRepo{
			config: domain.EdgeConfig{
				EdgeName:    "Factory Edge",
				Environment: "staging",
			},
		},
	})

	input := runner.resolveScriptInput(context.Background(), TriggerRequest{
		Instance: domain.Instance{
			ID:       "instance-1",
			Name:     "Collector 1",
			DeviceID: "device-1",
		},
		Trigger: domain.TriggerWorkflow,
		Input: map[string]any{
			"upstream": map[string]any{
				"output": map[string]any{
					"temperature": float64(42),
				},
			},
		},
	}, map[string]any{
		"device_name":   "{{device.name}}",
		"device_sector": "{{device.sector}}",
		"edge_name":     "{{system.edge_name}}",
		"instance_name": "{{system.instance_name}}",
		"upstream_temp": "{{input.upstream.output.temperature}}",
		"trigger_type":  "{{trigger.type}}",
		"manual":        "value",
	})

	if input["device_name"] != "Pump A" || input["device_sector"] != "line-7" {
		t.Fatalf("expected resolved device template input, got %#v", input)
	}
	if input["edge_name"] != "Factory Edge" || input["instance_name"] != "Collector 1" {
		t.Fatalf("expected resolved system aliases, got %#v", input)
	}
	if numericValue(input["upstream_temp"]) != 42 || fmt.Sprint(input["trigger_type"]) != string(domain.TriggerWorkflow) {
		t.Fatalf("expected trigger input context to be available, got %#v", input)
	}
	if input["manual"] != "value" {
		t.Fatalf("expected plain input to remain unchanged, got %#v", input)
	}
}

type fakeSparkitExecutor struct {
	input   map[string]any
	results []domain.ScriptResult
	errors  []error
	calls   int
}

func (e *fakeSparkitExecutor) Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error) {
	return domain.ScriptResult{}, nil
}

func (e *fakeSparkitExecutor) RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error) {
	e.input = input
	e.calls++
	if len(e.errors) >= e.calls && e.errors[e.calls-1] != nil {
		return domain.ScriptResult{}, e.errors[e.calls-1]
	}
	if len(e.results) >= e.calls {
		return e.results[e.calls-1], nil
	}
	return domain.ScriptResult{
		Stdout:   `{"ok":true}`,
		ExitCode: 0,
		Data:     map[string]any{"ok": true, "temperature": float64(42)},
	}, nil
}

func (e *fakeSparkitExecutor) Schema(ctx context.Context, scriptPath string) (map[string]any, error) {
	return map[string]any{}, nil
}

type fakeDevicesRepo struct {
	items map[string]domain.Device
}

func (f fakeDevicesRepo) FindByID(_ context.Context, id string) (domain.Device, error) {
	return f.items[id], nil
}

type fakeEdgeConfigRepo struct {
	config domain.EdgeConfig
}

func (f fakeEdgeConfigRepo) GetEdgeConfig(_ context.Context) (domain.EdgeConfig, error) {
	return f.config, nil
}

func numericValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func hasLogCode(logs []domain.ExecutionLog, code string) bool {
	for _, log := range logs {
		if log.Code == code {
			return true
		}
	}
	return false
}
