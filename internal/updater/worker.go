package updater

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WorkerArgs struct {
	AppRoot     string
	PackageRoot string
	BackupRoot  string
	Target      string
	Version     string
	HealthURL   string
}

func RunInternalWorker(args []string) (bool, error) {
	for _, arg := range args {
		if arg == "--internal-update-worker" {
			fs := flag.NewFlagSet("internal-update-worker", flag.ContinueOnError)
			fs.SetOutput(ioDiscard{})
			appRoot := fs.String("app-root", "", "")
			packageRoot := fs.String("package-root", "", "")
			backupRoot := fs.String("backup-root", "", "")
			target := fs.String("target", "", "")
			version := fs.String("version", "", "")
			healthURL := fs.String("health-url", "", "")
			if err := fs.Parse(args[1:]); err != nil {
				return true, err
			}
			return true, RunWorker(WorkerArgs{
				AppRoot:     strings.TrimSpace(*appRoot),
				PackageRoot: strings.TrimSpace(*packageRoot),
				BackupRoot:  strings.TrimSpace(*backupRoot),
				Target:      strings.TrimSpace(*target),
				Version:     strings.TrimSpace(*version),
				HealthURL:   strings.TrimSpace(*healthURL),
			})
		}
	}
	return false, nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func RunWorker(args WorkerArgs) error {
	if args.AppRoot == "" || args.PackageRoot == "" || args.BackupRoot == "" || args.Target == "" {
		return errors.New("worker args are incomplete")
	}

	statePath := updateStatePathForRoot(args.AppRoot)
	previous, _ := loadStateFromPath(statePath)
	preserved := []string{"config.yml", "sparkedge.db"}

	appliedFiles, err := applyPackageFiles(args.PackageRoot, args.AppRoot, args.Target)
	if err != nil {
		_ = saveStateWithHistoryAtPath(statePath, previous, previous, HistoryEntry{
			Type:      "apply",
			Status:    "failed",
			Version:   args.Version,
			Target:    args.Target,
			Message:   err.Error(),
			Artifact:  args.PackageRoot,
			CreatedAt: time.Now().UTC(),
		})
		return err
	}

	applyResult := &ApplyResult{
		Version:         args.Version,
		Target:          args.Target,
		StagingPath:     args.PackageRoot,
		BackupPath:      args.BackupRoot,
		Applied:         true,
		PreparedOnly:    false,
		RestartRequired: false,
		Message:         "Atualizacao aplicada no diretorio principal. Validando a subida da nova versao.",
		AppliedFiles:    appliedFiles,
		PreservedFiles:  preserved,
	}
	_ = saveStateWithHistoryAtPath(statePath, previous, UpdateState{
		LastDownloadedPackage: previous.LastDownloadedPackage,
		LastPreparedVersion:   args.Version,
		LastPreparedTarget:    args.Target,
		LastApplyResult:       applyResult,
		LastDownloadResult:    previous.LastDownloadResult,
		LastExecuteResult:     previous.LastExecuteResult,
		LastRollbackResult:    previous.LastRollbackResult,
		LastRestartResult:     previous.LastRestartResult,
	}, HistoryEntry{
		Type:      "apply",
		Status:    "applied",
		Version:   args.Version,
		Target:    args.Target,
		Message:   applyResult.Message,
		Artifact:  args.PackageRoot,
		CreatedAt: time.Now().UTC(),
	})

	process, err := startAppBinary(args.AppRoot, args.Target)
	if err != nil {
		return rollbackAndRestartPrevious(statePath, previous, args, fmt.Errorf("start new version: %w", err))
	}

	restartResult := &RestartResult{
		Executed:       true,
		ManualRequired: false,
		Command:        filepath.Join(args.AppRoot, executableNameForTarget(args.Target)),
		Message:        "Nova versao iniciada. Verificando a saude do processo.",
		UpdatedAt:      time.Now().UTC(),
	}
	_ = saveStateWithHistoryAtPath(statePath, previous, UpdateState{
		LastDownloadedPackage: previous.LastDownloadedPackage,
		LastPreparedVersion:   args.Version,
		LastPreparedTarget:    args.Target,
		LastApplyResult:       applyResult,
		LastDownloadResult:    previous.LastDownloadResult,
		LastExecuteResult:     previous.LastExecuteResult,
		LastRollbackResult:    previous.LastRollbackResult,
		LastRestartResult:     restartResult,
	}, HistoryEntry{
		Type:      "restart",
		Status:    "executed",
		Version:   args.Version,
		Target:    args.Target,
		Message:   restartResult.Message,
		Artifact:  restartResult.Command,
		CreatedAt: restartResult.UpdatedAt,
	})

	if err := waitForHealth(args.HealthURL, 90*time.Second); err != nil {
		_ = process.Kill()
		return rollbackAndRestartPrevious(statePath, previous, args, err)
	}

	successState, _ := loadStateFromPath(statePath)
	successRestart := *restartResult
	successRestart.Message = "Atualizacao concluida e nova versao respondeu no health check."
	successRestart.UpdatedAt = time.Now().UTC()
	return saveStateWithHistoryAtPath(statePath, successState, UpdateState{
		LastDownloadedPackage: successState.LastDownloadedPackage,
		LastPreparedVersion:   args.Version,
		LastPreparedTarget:    args.Target,
		LastApplyResult:       applyResult,
		LastDownloadResult:    successState.LastDownloadResult,
		LastExecuteResult:     successState.LastExecuteResult,
		LastRollbackResult:    successState.LastRollbackResult,
		LastRestartResult:     &successRestart,
	}, HistoryEntry{
		Type:      "health",
		Status:    "validated",
		Version:   args.Version,
		Target:    args.Target,
		Message:   successRestart.Message,
		Artifact:  strings.TrimSpace(args.HealthURL),
		CreatedAt: successRestart.UpdatedAt,
	})
}

func rollbackAndRestartPrevious(statePath string, previous UpdateState, args WorkerArgs, failure error) error {
	restoredFiles, restoreErr := restoreBackup(args.BackupRoot, args.AppRoot, args.Target)
	if restoreErr != nil {
		_ = saveStateWithHistoryAtPath(statePath, previous, previous, HistoryEntry{
			Type:      "rollback",
			Status:    "failed",
			Version:   args.Version,
			Target:    args.Target,
			Message:   restoreErr.Error(),
			Artifact:  args.BackupRoot,
			CreatedAt: time.Now().UTC(),
		})
		return errors.Join(failure, restoreErr)
	}

	rollbackResult := &RollbackResult{
		Version:         args.Version,
		Target:          args.Target,
		BackupPath:      args.BackupRoot,
		Applied:         true,
		PreparedOnly:    false,
		RestartRequired: false,
		Message:         "Falha ao validar a nova versao. O backup anterior foi restaurado automaticamente.",
		RestoredFiles:   restoredFiles,
		UpdatedAt:       time.Now().UTC(),
	}
	_ = saveStateWithHistoryAtPath(statePath, previous, UpdateState{
		LastDownloadedPackage: previous.LastDownloadedPackage,
		LastPreparedVersion:   previous.LastPreparedVersion,
		LastPreparedTarget:    previous.LastPreparedTarget,
		LastApplyResult:       previous.LastApplyResult,
		LastDownloadResult:    previous.LastDownloadResult,
		LastExecuteResult:     previous.LastExecuteResult,
		LastRollbackResult:    rollbackResult,
		LastRestartResult:     previous.LastRestartResult,
	}, HistoryEntry{
		Type:      "rollback",
		Status:    "applied",
		Version:   args.Version,
		Target:    args.Target,
		Message:   rollbackResult.Message,
		Artifact:  args.BackupRoot,
		CreatedAt: rollbackResult.UpdatedAt,
	})

	_, restartErr := startAppBinary(args.AppRoot, args.Target)
	if restartErr != nil {
		return errors.Join(failure, restartErr)
	}
	return failure
}

func startAppBinary(appRoot string, target string) (*os.Process, error) {
	executable := filepath.Join(appRoot, executableNameForTarget(target))
	cmd := exec.Command(executable)
	cmd.Dir = appRoot
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, nil
}

func waitForHealth(healthURL string, timeout time.Duration) error {
	healthURL = strings.TrimSpace(healthURL)
	if healthURL == "" {
		time.Sleep(4 * time.Second)
		return nil
	}

	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		res, err := client.Get(healthURL)
		if err == nil && res != nil {
			_ = res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("health endpoint returned status %d", res.StatusCode)
		} else if err != nil {
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	if lastErr == nil {
		lastErr = errors.New("health check timeout")
	}
	return lastErr
}
