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
			ScriptParameters: map[string]any{"host": "127.0.0.1", "port": float64(502)},
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
		Input: map[string]any{"port": float64(503)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionSuccess || result.Output["ok"] != true {
		t.Fatalf("unexpected result %#v", result)
	}
	if executor.input["host"] != "127.0.0.1" || executor.input["port"] != float64(503) {
		t.Fatalf("unexpected merged input %#v", executor.input)
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
		Data:     map[string]any{"ok": true},
	}, nil
}

func (e *fakeSparkitExecutor) Schema(ctx context.Context, scriptPath string) (map[string]any, error) {
	return map[string]any{}, nil
}
