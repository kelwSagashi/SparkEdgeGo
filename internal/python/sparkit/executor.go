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

func (e *Executor) ReadmeFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.DefaultTimeout)
	defer cancel()

	output, err := commandOutput(ctx, pythonExecutable(venvPath), filepath.Join(scriptFolder, mainFile), "--readme")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
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
	exitCode := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return domain.ScriptResult{}, err
		}
		exitCode = exitErr.ExitCode()
	}
	result := domain.ScriptResult{
		Stdout: string(stdout),
		Stderr: string(stderr),
	}

	if parsed, ok := parseSparkitEnvelope(stdout); ok {
		if exitCode != 0 {
			result.ExitCode = exitCode
			result.ErrorData = parsed.envelope
			result.Stderr = structuredErrorMessage(parsed.stderr, result.Stderr)
			if result.Stderr == "" && parsed.stderr != nil {
				result.Stderr = fmt.Sprint(parsed.stderr)
			}
			return result, nil
		}

		result.ExitCode = 0
		if data, ok := parsed.stdout.(map[string]any); ok {
			result.Data = data
		}
		if result.Stderr == "" {
			result.Stderr = structuredErrorMessage(parsed.stderr, result.Stderr)
		}
		return result, nil
	}

	if data, ok := parseJSONMap(stdout); ok {
		result.Data = data
	}

	result.ExitCode = exitCode
	return result, nil
}

type sparkitParsedEnvelope struct {
	envelope map[string]any
	stdout   any
	stderr   any
}

func parseSparkitEnvelope(output []byte) (sparkitParsedEnvelope, bool) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return sparkitParsedEnvelope{}, false
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &raw); err != nil {
		return sparkitParsedEnvelope{}, false
	}
	if len(raw) != 2 {
		return sparkitParsedEnvelope{}, false
	}

	stdoutRaw, hasStdout := raw["stdout"]
	stderrRaw, hasStderr := raw["stderr"]
	if !hasStdout || !hasStderr {
		return sparkitParsedEnvelope{}, false
	}

	stdoutValue, err := decodeJSONValue(stdoutRaw)
	if err != nil {
		return sparkitParsedEnvelope{}, false
	}
	stderrValue, err := decodeJSONValue(stderrRaw)
	if err != nil {
		return sparkitParsedEnvelope{}, false
	}

	envelope := map[string]any{
		"stdout": stdoutValue,
		"stderr": stderrValue,
	}
	return sparkitParsedEnvelope{
		envelope: envelope,
		stdout:   stdoutValue,
		stderr:   stderrValue,
	}, true
}

func parseJSONMap(output []byte) (map[string]any, bool) {
	var data map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output), &data); err != nil {
		return nil, false
	}
	if data == nil {
		return map[string]any{}, true
	}
	return data, true
}

func decodeJSONValue(raw json.RawMessage) (any, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func structuredErrorMessage(stderr any, fallback string) string {
	switch value := stderr.(type) {
	case nil:
		return strings.TrimSpace(fallback)
	case string:
		if strings.TrimSpace(value) != "" {
			return value
		}
	case map[string]any:
		if message, ok := value["message"].(string); ok && strings.TrimSpace(message) != "" {
			if runtimeStderr, ok := value["runtime_stderr"].(string); ok && strings.TrimSpace(runtimeStderr) != "" {
				return message
			}
			return message
		}
		if runtimeStderr, ok := value["runtime_stderr"].(string); ok && strings.TrimSpace(runtimeStderr) != "" {
			return runtimeStderr
		}
	}
	return strings.TrimSpace(fallback)
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
