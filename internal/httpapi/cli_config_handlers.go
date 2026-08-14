package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/cloudsync"
	"github.com/kelwSagashi/sparkedge-go/internal/config"
	"github.com/kelwSagashi/sparkedge-go/internal/updater"
)

func (s *Server) handleCliConfigGet(_ *http.Request) (any, error) {
	effective, _, err := s.deps.Config.Load()
	if err != nil {
		return nil, err
	}
	return effective, nil
}

func (s *Server) handleCliConfigUpdate(r *http.Request) (any, error) {
	var req config.Update
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Cloud == nil && req.DB == nil && req.Auth == nil && req.Server == nil && req.Update == nil && req.Connectivity == nil && req.Retention == nil {
		return nil, NewHTTPError(http.StatusBadRequest, "Nenhum campo valido para atualizar.")
	}

	effective, err := s.deps.Config.Save(req)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, err.Error())
	}
	_, runtimeCfg, loadErr := s.deps.Config.Load()
	if loadErr == nil {
		s.deps.RuntimeCfg = runtimeCfg
		if s.deps.Updater != nil {
			s.deps.Updater.UpdateConfig(updater.Config{
				Enabled:         runtimeCfg.Update.Enabled,
				Provider:        runtimeCfg.Update.Provider,
				Repo:            runtimeCfg.Update.Repo,
				Channel:         runtimeCfg.Update.Channel,
				AllowPrerelease: runtimeCfg.Update.AllowPrerelease,
				ServiceName:     runtimeCfg.Update.ServiceName,
				RestartCommand:  runtimeCfg.Update.RestartCommand,
			})
		}
		if s.deps.CloudSync != nil {
			edgeID := ""
			if s.deps.Edge != nil {
				if provisioned, edgeErr := s.deps.Edge.Load(r.Context()); edgeErr == nil {
					edgeID = provisioned.EdgeID
				}
			}
			s.deps.CloudSync.UpdateConfig(cloudsync.Config{
				BaseURL:           runtimeCfg.CloudURL,
				EdgeID:            edgeID,
				SyncToken:         runtimeCfg.CloudSyncToken,
				Enabled:           true,
				SchemaVersion:     "edge-cloud.v1",
				MaxAttempts:       12,
				MaxBatchSize:      50,
				HighPriorityDelay: 15 * time.Second,
				LowPriorityDelay:  45 * time.Second,
			})
		}
	}

	return map[string]any{
		"success": true,
		"message": "Configuracao salva em config.yml. Ajustes do updater e cloud sync ja foram reaplicados; outras mudancas podem exigir reinicio.",
		"config":  effective,
	}, nil
}
