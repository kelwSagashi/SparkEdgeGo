package updater

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
)

type ExecuteResult struct {
	Version        string    `json:"version"`
	Target         string    `json:"target"`
	DownloadedPath string    `json:"downloaded_path"`
	StagingPath    string    `json:"staging_path"`
	BackupPath     string    `json:"backup_path"`
	LauncherPath   string    `json:"launcher_path"`
	WorkerBinary   string    `json:"worker_binary"`
	HealthURL      string    `json:"health_url,omitempty"`
	ScheduledExit  bool      `json:"scheduled_exit"`
	Message        string    `json:"message"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *Service) ExecuteDownloaded(ctx context.Context, downloadedPath string) (ExecuteResult, error) {
	applyResult, err := s.ApplyDownloaded(ctx, downloadedPath)
	if err != nil {
		return ExecuteResult{}, err
	}

	workerBinary := filepath.Join(applyResult.StagingPath, executableNameForTarget(applyResult.Target))
	launcherPath, err := writeWorkerLauncher(filepath.Dir(applyResult.StagingPath), workerLaunchConfig{
		ParentPID:   s.pid(),
		Executable:  workerBinary,
		AppRoot:     appfs.AppRoot(),
		PackageRoot: applyResult.StagingPath,
		BackupPath:  applyResult.BackupPath,
		Target:      applyResult.Target,
		Version:     applyResult.Version,
		HealthURL:   strings.TrimSpace(s.config.HealthURL),
	})
	if err != nil {
		return ExecuteResult{}, err
	}

	if err := s.start(launcherCommandForPath(launcherPath)); err != nil {
		return ExecuteResult{}, err
	}

	result := ExecuteResult{
		Version:        applyResult.Version,
		Target:         applyResult.Target,
		DownloadedPath: applyResult.DownloadedPath,
		StagingPath:    applyResult.StagingPath,
		BackupPath:     applyResult.BackupPath,
		LauncherPath:   launcherPath,
		WorkerBinary:   workerBinary,
		HealthURL:      strings.TrimSpace(s.config.HealthURL),
		ScheduledExit:  true,
		UpdatedAt:      time.Now().UTC(),
		Message:        "Atualizacao automatica agendada. O SparkEdge atual vai encerrar, aplicar a nova versao e iniciar novamente sozinho.",
	}

	previous, _ := s.LoadState()
	_ = s.saveStateWithHistory(previous, UpdateState{
		LastDownloadedPackage: previous.LastDownloadedPackage,
		LastPreparedVersion:   applyResult.Version,
		LastPreparedTarget:    applyResult.Target,
		LastApplyResult:       previous.LastApplyResult,
		LastDownloadResult:    previous.LastDownloadResult,
		LastExecuteResult:     &result,
		LastRollbackResult:    previous.LastRollbackResult,
		LastRestartResult:     previous.LastRestartResult,
	}, HistoryEntry{
		Type:      "execute",
		Status:    "scheduled",
		Version:   applyResult.Version,
		Target:    applyResult.Target,
		Message:   result.Message,
		Artifact:  launcherPath,
		CreatedAt: result.UpdatedAt,
	})

	s.exit(1500 * time.Millisecond)
	return result, nil
}

type workerLaunchConfig struct {
	ParentPID   int
	Executable  string
	AppRoot     string
	PackageRoot string
	BackupPath  string
	Target      string
	Version     string
	HealthURL   string
}

func writeWorkerLauncher(scriptDir string, cfg workerLaunchConfig) (string, error) {
	if runtime.GOOS == "windows" {
		return writeWindowsWorkerLauncher(scriptDir, cfg)
	}
	return writeUnixWorkerLauncher(scriptDir, cfg)
}

func writeWindowsWorkerLauncher(scriptDir string, cfg workerLaunchConfig) (string, error) {
	scriptPath := filepath.Join(scriptDir, "launch-worker.ps1")
	lines := []string{
		"$ErrorActionPreference = 'Stop'",
		fmt.Sprintf("$parentPid = %d", cfg.ParentPID),
		"$exePath = '" + strings.ReplaceAll(cfg.Executable, "'", "''") + "'",
		"$args = @(",
		"  '--internal-update-worker',",
		"  '--app-root', '" + strings.ReplaceAll(cfg.AppRoot, "'", "''") + "',",
		"  '--package-root', '" + strings.ReplaceAll(cfg.PackageRoot, "'", "''") + "',",
		"  '--backup-root', '" + strings.ReplaceAll(cfg.BackupPath, "'", "''") + "',",
		"  '--target', '" + strings.ReplaceAll(cfg.Target, "'", "''") + "',",
		"  '--version', '" + strings.ReplaceAll(cfg.Version, "'", "''") + "',",
	}
	if strings.TrimSpace(cfg.HealthURL) != "" {
		lines = append(lines, "  '--health-url', '"+strings.ReplaceAll(cfg.HealthURL, "'", "''")+"',")
	}
	lines = append(lines,
		")",
		"if ($parentPid -gt 0) {",
		"  try { Wait-Process -Id $parentPid -Timeout 180 -ErrorAction SilentlyContinue } catch {}",
		"  Start-Sleep -Seconds 2",
		"}",
		"Start-Process -FilePath $exePath -WorkingDirectory (Split-Path -Parent $exePath) -ArgumentList $args -WindowStyle Hidden",
	)
	content := strings.Join(lines, "\r\n") + "\r\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return scriptPath, nil
}

func writeUnixWorkerLauncher(scriptDir string, cfg workerLaunchConfig) (string, error) {
	scriptPath := filepath.Join(scriptDir, "launch-worker.sh")
	args := []string{
		"--internal-update-worker",
		"--app-root", cfg.AppRoot,
		"--package-root", cfg.PackageRoot,
		"--backup-root", cfg.BackupPath,
		"--target", cfg.Target,
		"--version", cfg.Version,
	}
	if strings.TrimSpace(cfg.HealthURL) != "" {
		args = append(args, "--health-url", cfg.HealthURL)
	}
	lines := []string{
		"#!/usr/bin/env sh",
		"set -eu",
		"parent_pid='" + strconv.Itoa(cfg.ParentPID) + "'",
		"exe_path='" + escapeSingleQuotes(cfg.Executable) + "'",
		"if [ \"$parent_pid\" -gt 0 ] 2>/dev/null; then",
		"  while kill -0 \"$parent_pid\" 2>/dev/null; do",
		"    sleep 1",
		"  done",
		"  sleep 2",
		"fi",
		"nohup \"$exe_path\" " + shellJoin(args) + " >/dev/null 2>&1 &",
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
		return "", err
	}
	return scriptPath, nil
}

func shellJoin(args []string) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, "'"+escapeSingleQuotes(arg)+"'")
	}
	return strings.Join(parts, " ")
}

func launcherCommandForPath(path string) string {
	return filepath.Clean(path)
}

func startDetachedCommand(command string, args ...string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("launcher command is empty")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		invocation := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", command}
		invocation = append(invocation, args...)
		cmd = exec.Command("powershell", invocation...)
	} else {
		invocation := append([]string{command}, args...)
		cmd = exec.Command(invocation[0], invocation[1:]...)
	}
	return cmd.Start()
}
