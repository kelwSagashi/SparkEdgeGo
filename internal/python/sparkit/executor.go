package sparkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type Executor struct {
	DefaultTimeout time.Duration
}

func NewExecutor() *Executor {
	return &Executor{DefaultTimeout: 60 * time.Second}
}

func (e *Executor) Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error) {
	return e.runWithPython(ctx, "python", scriptPath, input)
}

func (e *Executor) Schema(ctx context.Context, scriptPath string) (map[string]any, error) {
	return e.schemaWithPython(ctx, "python", scriptPath)
}

func (e *Executor) CreateVenv(ctx context.Context, venvPath string) error {
	if err := runCommand(ctx, "python", "-m", "venv", venvPath); err != nil {
		return runCommand(ctx, "python3", "-m", "venv", venvPath)
	}
	return nil
}

func (e *Executor) InstallRequirements(ctx context.Context, venvPath string, requirementsPath string) error {
	return runCommand(ctx, pipPath(venvPath), "install", "-r", requirementsPath)
}

func (e *Executor) SchemaFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string) (map[string]any, error) {
	return e.schemaWithPython(ctx, pythonExecutable(venvPath), filepath.Join(scriptFolder, mainFile))
}

func (e *Executor) RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error) {
	return e.runWithPython(ctx, pythonExecutable(venvPath), filepath.Join(scriptFolder, mainFile), input)
}

func (e *Executor) schemaWithPython(ctx context.Context, pythonExecutable string, scriptPath string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, e.DefaultTimeout)
	defer cancel()

	output, err := commandOutput(ctx, pythonExecutable, scriptPath, "--schema")
	if err != nil {
		return nil, err
	}

	var result struct {
		Schema map[string]any `json:"schema"`
	}
	trimmed := bytes.TrimSpace(output)
	if err := json.Unmarshal(trimmed, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema from script stdout: %w. stdout: %s", err, strings.TrimSpace(string(trimmed)))
	}
	if result.Schema == nil {
		return map[string]any{}, nil
	}
	return result.Schema, nil
}

func (e *Executor) runWithPython(ctx context.Context, pythonExecutable string, scriptPath string, input map[string]any) (domain.ScriptResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.DefaultTimeout)
	defer cancel()

	tempFile, err := os.CreateTemp("", "spark-edge_input_*.json")
	if err != nil {
		return domain.ScriptResult{}, err
	}
	defer os.Remove(tempFile.Name())

	if err := json.NewEncoder(tempFile).Encode(input); err != nil {
		_ = tempFile.Close()
		return domain.ScriptResult{}, err
	}
	if err := tempFile.Close(); err != nil {
		return domain.ScriptResult{}, err
	}

	cmd := exec.CommandContext(ctx, pythonExecutable, scriptPath, "--input-file", tempFile.Name())
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err = cmd.Run()

	stdout := stdoutBuffer.Bytes()
	stderr := stderrBuffer.Bytes()
	var data map[string]any
	_ = json.Unmarshal(bytes.TrimSpace(stdout), &data)

	result := domain.ScriptResult{
		Stdout: string(stdout),
		Stderr: string(stderr),
		Data:   data,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return domain.ScriptResult{}, err
	}

	result.ExitCode = 0
	return result, nil
}

func runCommand(ctx context.Context, name string, args ...string) error {
	output, err := commandOutput(ctx, name, args...)
	if err != nil {
		return err
	}
	_ = output
	return nil
}

func commandOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func pythonExecutable(venvPath string) string {
	if venvPath == "" {
		return "python"
	}
	executable := pythonPath(venvPath)
	if _, err := os.Stat(executable); err != nil {
		return "python"
	}
	return executable
}

func pythonPath(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

func pipPath(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	return filepath.Join(venvPath, "bin", "pip")
}
