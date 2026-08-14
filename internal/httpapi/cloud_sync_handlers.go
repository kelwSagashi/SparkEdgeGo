package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleCloudSyncList(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return map[string]any{"data": []any{}}, nil
	}
	items, err := s.deps.CloudSync.ListRecent(r.Context(), 100)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Server) handleCloudSyncStats(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return map[string]any{"configured": false}, nil
	}
	stats, err := s.deps.CloudSync.Stats(r.Context())
	if err != nil {
		return nil, err
	}
	retention := sqlite.CurrentRetentionSnapshot()
	stats["retention"] = retention["cloud_sync"]
	if s.deps.DB != nil && s.deps.DB.MqttQueue != nil {
		if mqttStats, err := s.deps.DB.MqttQueue.Stats(r.Context()); err == nil {
			if createdAt, ok := mqttStats["oldest_pending_created_at"].(time.Time); ok && !createdAt.IsZero() {
				mqttStats["oldest_pending_age_seconds"] = int64(time.Since(createdAt).Seconds())
			}
			if total := toInt64(mqttStats["total"]); total >= 0 {
				limit := toInt64(retention["mqtt"].(map[string]any)["max_items"])
				mqttStats["usage_pct"] = percentOf(total, limit)
			}
			mqttStats["retention"] = retention["mqtt"]
			stats["mqtt_queue"] = mqttStats
		}
	}
	stats["usage"] = map[string]any{
		"pending_total_pct_of_failed_window": percentOf(
			toInt64(stats["failed"]),
			toInt64(retention["cloud_sync"].(map[string]any)["keep_failed_items"]),
		),
		"sent_pct_of_sent_window": percentOf(
			toInt64(stats["sent"]),
			toInt64(retention["cloud_sync"].(map[string]any)["keep_sent_items"]),
		),
	}
	stats["connectivity"] = cliConnectivitySnapshot(r, s)
	return stats, nil
}

func (s *Server) handleCloudSyncFlush(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	return s.deps.CloudSync.Flush(r.Context(), 50)
}

func (s *Server) handleCloudSyncRetry(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	result, err := s.deps.CloudSync.RetryItem(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "cloud sync item not found")
		}
		return nil, err
	}
	return result, nil
}

func (s *Server) handleCloudSyncDelete(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	if err := s.deps.CloudSync.DeleteItem(r.Context(), r.PathValue("id")); err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "cloud sync item not found")
		}
		return nil, err
	}
	return map[string]any{
		"success": true,
		"id":      r.PathValue("id"),
	}, nil
}

func percentOf(value int64, limit int64) int64 {
	if limit <= 0 {
		return 0
	}
	return (value * 100) / limit
}
