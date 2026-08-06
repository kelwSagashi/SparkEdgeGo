package runtime

import (
	"context"
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

	payload := runner.applyMapping(context.Background(), &domain.DataMapping{
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
	}, map[string]any{
		"device_name":   "{{device.name}}",
		"device_sector": "{{device.sector}}",
		"edge_name":     "{{system.edge_name}}",
		"instance_name": "{{system.instance_name}}",
		"manual":        "value",
	})

	if input["device_name"] != "Pump A" || input["device_sector"] != "line-7" {
		t.Fatalf("expected resolved device template input, got %#v", input)
	}
	if input["edge_name"] != "Factory Edge" || input["instance_name"] != "Collector 1" {
		t.Fatalf("expected resolved system aliases, got %#v", input)
	}
	if input["manual"] != "value" {
		t.Fatalf("expected plain input to remain unchanged, got %#v", input)
	}
}

type fakeSparkitExecutor struct {
	input map[string]any
}

func (e *fakeSparkitExecutor) Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error) {
	return domain.ScriptResult{}, nil
}

func (e *fakeSparkitExecutor) RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error) {
	e.input = input
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
