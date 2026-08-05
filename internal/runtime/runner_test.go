package runtime

import (
	"context"
	"testing"

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
