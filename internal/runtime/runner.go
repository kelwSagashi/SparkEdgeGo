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
)

type SparkitExecutor interface {
	Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error)
	RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error)
	Schema(ctx context.Context, scriptPath string) (map[string]any, error)
}

type Dependencies struct {
	Sparkit   SparkitExecutor
	Providers *providers.Registry
}

type Runner struct {
	deps Dependencies
}

func NewRunner(deps Dependencies) *Runner {
	return &Runner{deps: deps}
}

type TriggerRequest struct {
	Instance domain.Instance
	Script   domain.DownloadedScript
	Trigger  domain.TriggerType
	Input    map[string]any
}

type TriggerResult struct {
	ExecutionID string
	Status      domain.ExecutionStatus
	Output      map[string]any
	RawOutput   string
	Error       string
	Logs        []domain.ExecutionLog
	DurationMS  int
	StartedAt   time.Time
	FinishedAt  time.Time
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
	return finish(result, domain.ExecutionSuccess, ""), nil
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
