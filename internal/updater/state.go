package updater

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
)

type UpdateState struct {
	LastDownloadedPackage string          `json:"last_downloaded_package,omitempty"`
	LastPreparedVersion   string          `json:"last_prepared_version,omitempty"`
	LastPreparedTarget    string          `json:"last_prepared_target,omitempty"`
	LastApplyResult       *ApplyResult    `json:"last_apply_result,omitempty"`
	LastDownloadResult    *DownloadResult `json:"last_download_result,omitempty"`
	LastRollbackResult    *RollbackResult `json:"last_rollback_result,omitempty"`
	LastRestartResult     *RestartResult  `json:"last_restart_result,omitempty"`
	History               []HistoryEntry  `json:"history,omitempty"`
	UpdatedAt             time.Time       `json:"updated_at,omitempty"`
}

type HistoryEntry struct {
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Version   string    `json:"version,omitempty"`
	Target    string    `json:"target,omitempty"`
	Message   string    `json:"message,omitempty"`
	Artifact  string    `json:"artifact,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

const maxHistoryEntries = 20

func (s *Service) LoadState() (UpdateState, error) {
	path := updateStatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UpdateState{}, nil
		}
		return UpdateState{}, err
	}
	var state UpdateState
	if err := json.Unmarshal(data, &state); err != nil {
		return UpdateState{}, err
	}
	return state, nil
}

func (s *Service) saveState(state UpdateState) error {
	path := updateStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func (s *Service) saveStateWithHistory(previous UpdateState, next UpdateState, entry HistoryEntry) error {
	entry.CreatedAt = entry.CreatedAt.UTC()
	if len(previous.History) > 0 {
		next.History = append([]HistoryEntry{}, previous.History...)
	}
	next.History = append(next.History, entry)
	if len(next.History) > maxHistoryEntries {
		next.History = append([]HistoryEntry{}, next.History[len(next.History)-maxHistoryEntries:]...)
	}
	next.UpdatedAt = entry.CreatedAt
	return s.saveState(next)
}

func updateStatePath() string {
	return appfs.ResolveFromRoot("updates", "state.json")
}
