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
	UpdatedAt             time.Time       `json:"updated_at,omitempty"`
}

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

func updateStatePath() string {
	return appfs.ResolveFromRoot("updates", "state.json")
}
