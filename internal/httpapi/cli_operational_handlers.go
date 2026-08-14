package httpapi

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleCliOperationalSummary(r *http.Request) (any, error) {
	retention := sqlite.CurrentRetentionSnapshot()
	connectivity := cliConnectivitySnapshot(r, s)

	mqttStats := map[string]any{
		"configured": s != nil && s.deps.MQTT != nil,
	}
	if s != nil && s.deps.DB != nil && s.deps.DB.MqttQueue != nil {
		stats, err := s.deps.DB.MqttQueue.Stats(r.Context())
		if err != nil {
			return nil, err
		}
		if createdAt, ok := stats["oldest_pending_created_at"].(time.Time); ok && !createdAt.IsZero() {
			stats["oldest_pending_age_seconds"] = int64(time.Since(createdAt).Seconds())
		}
		stats["retention"] = retention["mqtt"]
		stats["usage_pct"] = percentOf(
			toInt64(stats["total"]),
			toInt64(retention["mqtt"].(map[string]any)["max_items"]),
		)
		stats["severity"] = mqttSeverity(stats, connectivity)
		mqttStats = stats
	}

	fallbackStats := map[string]any{"configured": s != nil && s.deps.Runtime != nil}
	if s != nil && s.deps.DB != nil && s.deps.DB.LocalFallback != nil {
		stats, err := s.deps.DB.LocalFallback.Stats(r.Context())
		if err != nil {
			return nil, err
		}
		if createdAt, ok := stats["oldest_pending_created_at"].(time.Time); ok && !createdAt.IsZero() {
			stats["oldest_pending_age_seconds"] = int64(time.Since(createdAt).Seconds())
		}
		stats["retention"] = retention["local_fallback"]
		stats["usage"] = map[string]any{
			"sent_pct_of_sent_window": percentOf(
				toInt64(stats["sent"]),
				toInt64(retention["local_fallback"].(map[string]any)["keep_sent_items"]),
			),
			"failed_pct_of_failed_window": percentOf(
				toInt64(stats["failed"]),
				toInt64(retention["local_fallback"].(map[string]any)["keep_failed_items"]),
			),
		}
		stats["severity"] = fallbackSeverity(stats)
		fallbackStats = stats
	}

	cloudSyncStats := map[string]any{"configured": s != nil && s.deps.CloudSync != nil}
	if s != nil && s.deps.CloudSync != nil {
		stats, err := s.deps.CloudSync.Stats(r.Context())
		if err != nil {
			return nil, err
		}
		stats["retention"] = retention["cloud_sync"]
		stats["usage"] = map[string]any{
			"pending_total_pct_of_failed_window": percentOf(
				toInt64(stats["pending"])+toInt64(stats["failed"]),
				toInt64(retention["cloud_sync"].(map[string]any)["keep_failed_items"]),
			),
			"sent_pct_of_sent_window": percentOf(
				toInt64(stats["sent"]),
				toInt64(retention["cloud_sync"].(map[string]any)["keep_sent_items"]),
			),
		}
		stats["severity"] = cloudSyncSeverity(stats, connectivity)
		cloudSyncStats = stats
	}

	reasons := uniqueSortedStrings(
		collectReasons(connectivity["reasons"]),
		collectSeverityReason("mqtt_queue", mqttStats["severity"]),
		collectSeverityReason("local_fallback", fallbackStats["severity"]),
		collectSeverityReason("cloud_sync", cloudSyncStats["severity"]),
	)

	overallSeverity := maxSeverityLabel(
		asString(cloudSyncStats["severity"]),
		asString(mqttStats["severity"]),
		asString(fallbackStats["severity"]),
		connectivitySeverity(connectivity),
	)

	return map[string]any{
		"severity": overallSeverity,
		"status":   connectivity["status"],
		"mode":     connectivity["mode"],
		"reasons":  reasons,
		"connectivity": map[string]any{
			"status":                     connectivity["status"],
			"mode":                       connectivity["mode"],
			"reasons":                    connectivity["reasons"],
			"heartbeat_interval_seconds": connectivity["heartbeat_interval_seconds"],
			"stats_interval_seconds":     connectivity["stats_interval_seconds"],
			"mqtt_connected":             connectivity["mqtt_connected"],
			"cloud_sync_configured":      connectivity["cloud_sync_configured"],
			"policy":                     connectivity["policy"],
		},
		"queues": map[string]any{
			"mqtt":           mqttStats,
			"cloud_sync":     cloudSyncStats,
			"local_fallback": fallbackStats,
		},
		"retention": retention,
		"generated_at": time.Now().UTC(),
	}, nil
}

func cloudSyncSeverity(stats map[string]any, connectivity map[string]any) string {
	failed := toInt64(stats["failed"])
	pending := toInt64(stats["pending"]) + toInt64(stats["failed"])
	oldestAge := toInt64(stats["oldest_pending_age_seconds"])
	if failed > 0 || pending >= 25 || oldestAge >= 600 || strings.EqualFold(asString(connectivity["status"]), "degraded") {
		return "critical"
	}
	if pending >= 5 || oldestAge >= 120 || strings.EqualFold(asString(connectivity["status"]), "intermittent") {
		return "warning"
	}
	return "healthy"
}

func mqttSeverity(stats map[string]any, connectivity map[string]any) string {
	total := toInt64(stats["total"])
	oldestAge := toInt64(stats["oldest_pending_age_seconds"])
	if !toBool(connectivity["mqtt_connected"]) {
		return "critical"
	}
	if total >= 10 || oldestAge >= 600 {
		return "critical"
	}
	if total > 0 || oldestAge >= 120 {
		return "warning"
	}
	return "healthy"
}

func fallbackSeverity(stats map[string]any) string {
	failed := toInt64(stats["failed"])
	pending := toInt64(stats["pending"])
	sending := toInt64(stats["sending"])
	oldestAge := toInt64(stats["oldest_pending_age_seconds"])
	if failed > 0 || pending >= 25 || oldestAge >= 1800 {
		return "critical"
	}
	if pending > 0 || sending > 0 || oldestAge >= 600 {
		return "warning"
	}
	return "healthy"
}

func connectivitySeverity(snapshot map[string]any) string {
	switch strings.ToLower(asString(snapshot["status"])) {
	case "offline", "degraded":
		return "critical"
	case "intermittent":
		return "warning"
	default:
		return "healthy"
	}
}

func maxSeverityLabel(values ...string) string {
	best := "healthy"
	bestScore := -1
	for _, value := range values {
		score := severityScore(value)
		if score > bestScore {
			best = normalizeSeverity(value)
			bestScore = score
		}
	}
	return best
}

func severityScore(value string) int {
	switch normalizeSeverity(value) {
	case "critical":
		return 3
	case "warning":
		return 2
	case "healthy":
		return 1
	default:
		return 0
	}
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical", "degraded", "offline":
		return "critical"
	case "warning", "intermittent":
		return "warning"
	default:
		return "healthy"
	}
}

func collectReasons(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := asString(item); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func collectSeverityReason(prefix string, severity any) []string {
	switch normalizeSeverity(asString(severity)) {
	case "critical":
		return []string{prefix + "_critical"}
	case "warning":
		return []string{prefix + "_warning"}
	default:
		return nil
	}
}

func uniqueSortedStrings(values ...[]string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, group := range values {
		for _, value := range group {
			normalized := strings.TrimSpace(strings.ToLower(value))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			result = append(result, normalized)
		}
	}
	sort.Strings(result)
	return result
}

func asString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func toBool(value any) bool {
	if typed, ok := value.(bool); ok {
		return typed
	}
	return false
}
