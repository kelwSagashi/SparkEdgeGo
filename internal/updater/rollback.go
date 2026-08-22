package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
)

type RollbackResult struct {
	Version         string    `json:"version"`
	Target          string    `json:"target"`
	BackupPath      string    `json:"backup_path"`
	ScriptPath      string    `json:"script_path,omitempty"`
	Applied         bool      `json:"applied"`
	PreparedOnly    bool      `json:"prepared_only"`
	RestartRequired bool      `json:"restart_required"`
	Message         string    `json:"message"`
	RestoredFiles   []string  `json:"restored_files"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *Service) RollbackLatest(_ context.Context) (RollbackResult, error) {
	state, err := s.LoadState()
	if err != nil {
		return RollbackResult{}, err
	}
	if state.LastApplyResult == nil {
		return RollbackResult{}, errors.New("nenhuma aplicacao assistida foi registrada para rollback")
	}

	last := state.LastApplyResult
	result := RollbackResult{
		Version:         last.Version,
		Target:          last.Target,
		BackupPath:      last.BackupPath,
		RestartRequired: true,
		UpdatedAt:       time.Now().UTC(),
	}

	if strings.HasPrefix(last.Target, "windows-") {
		if strings.TrimSpace(last.RollbackPath) == "" {
			return RollbackResult{}, errors.New("rollback script is unavailable for this Windows update")
		}
		result.PreparedOnly = true
		result.ScriptPath = last.RollbackPath
		result.Message = "Rollback preparado. Pare o SparkEdge e execute o script de rollback para restaurar a versao anterior."
		_ = s.saveStateWithHistory(state, UpdateState{
			LastDownloadedPackage: state.LastDownloadedPackage,
			LastPreparedVersion:   state.LastPreparedVersion,
			LastPreparedTarget:    state.LastPreparedTarget,
			LastApplyResult:       state.LastApplyResult,
			LastDownloadResult:    state.LastDownloadResult,
			LastExecuteResult:     state.LastExecuteResult,
			LastRollbackResult:    &result,
			LastRestartResult:     state.LastRestartResult,
		}, HistoryEntry{
			Type:      "rollback",
			Status:    "prepared",
			Version:   result.Version,
			Target:    result.Target,
			Message:   result.Message,
			Artifact:  result.BackupPath,
			CreatedAt: result.UpdatedAt,
		})
		return result, nil
	}

	restoredFiles, err := restoreBackup(last.BackupPath, appfs.AppRoot(), last.Target)
	if err != nil {
		return RollbackResult{}, err
	}
	result.Applied = true
	result.RestoredFiles = restoredFiles
	result.ScriptPath = filepath.Join(last.BackupPath, "rollback-update.sh")
	result.Message = "Arquivos restaurados a partir do backup. Reinicie o SparkEdge para voltar para a versao anterior."
	_ = s.saveStateWithHistory(state, UpdateState{
		LastDownloadedPackage: state.LastDownloadedPackage,
		LastPreparedVersion:   state.LastPreparedVersion,
		LastPreparedTarget:    state.LastPreparedTarget,
		LastApplyResult:       state.LastApplyResult,
		LastDownloadResult:    state.LastDownloadResult,
		LastExecuteResult:     state.LastExecuteResult,
		LastRollbackResult:    &result,
		LastRestartResult:     state.LastRestartResult,
	}, HistoryEntry{
		Type:      "rollback",
		Status:    "applied",
		Version:   result.Version,
		Target:    result.Target,
		Message:   result.Message,
		Artifact:  result.BackupPath,
		CreatedAt: result.UpdatedAt,
	})
	return result, nil
}

func restoreBackup(backupRoot string, appRoot string, target string) ([]string, error) {
	restored := make([]string, 0, len(replaceableRelativePaths(target))+1)
	for _, relative := range replaceableRelativePaths(target) {
		source := filepath.Join(backupRoot, relative)
		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("restore source missing for %s", relative)
		}
		if err := copyPath(source, filepath.Join(appRoot, relative)); err != nil {
			return nil, fmt.Errorf("restore %s: %w", relative, err)
		}
		restored = append(restored, relative)
	}

	configSource := filepath.Join(backupRoot, "config.yml")
	if _, err := os.Stat(configSource); err == nil {
		if err := copyPath(configSource, filepath.Join(appRoot, "config.yml")); err != nil {
			return nil, err
		}
		restored = append(restored, "config.yml")
	}

	return restored, nil
}
