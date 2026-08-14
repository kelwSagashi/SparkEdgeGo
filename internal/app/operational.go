package app

import (
	"context"
	"runtime"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/connectivity"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/system"
)

func (a *App) collectOperationalSnapshot(ctx context.Context) map[string]any {
	snapshot := map[string]any{}
	for key, value := range systemStats() {
		snapshot[key] = value
	}

	activeInstances, runningInstances := a.instanceOperationalCounts(ctx)
	queueSizes, oldestPendingAgeSeconds, queueDetails := a.queueOperationalState(ctx)
	cloudSync := a.cloudSyncOperationalState(ctx)
	connectivitySnapshot := a.connectivitySnapshot(queueSizes, cloudSync, oldestPendingAgeSeconds)
	health := operationalHealthSnapshot(connectivitySnapshot, queueDetails, oldestPendingAgeSeconds)

	snapshot["status"] = connectivitySnapshot.Status
	snapshot["edge_version"] = envOrDefault("SPARKEDGE_VERSION", "go-dev")
	snapshot["os"] = runtime.GOOS
	snapshot["hardware"] = envOrDefault("SPARKEDGE_HARDWARE", runtime.GOARCH)
	snapshot["active_instances"] = map[string]any{
		"total":   activeInstances,
		"running": runningInstances,
	}
	snapshot["queue_sizes"] = queueSizes
	snapshot["queues"] = queueDetails
	snapshot["oldest_pending_age_seconds"] = oldestPendingAgeSeconds
	snapshot["connectivity"] = map[string]any{
		"mode":                       string(connectivitySnapshot.Mode),
		"status":                     connectivitySnapshot.Status,
		"severity":                   severityFromConnectivityStatus(connectivitySnapshot.Status),
		"reasons":                    connectivitySnapshot.Reasons,
		"mqtt_connected":             a != nil && a.MQTT != nil && a.MQTT.IsConnected(),
		"cloud_sync_configured":      a != nil && a.CloudSync != nil && a.CloudSync.Configured(),
		"cloud_sync_queue_depth":     queueSizes["cloud_sync"],
		"heartbeat_interval_seconds": connectivitySnapshot.HeartbeatIntervalSeconds,
		"stats_interval_seconds":     connectivitySnapshot.StatsIntervalSeconds,
		"policy":                     a.RuntimeCfg.Connectivity.Normalize(),
	}
	snapshot["cloud_sync"] = cloudSync
	snapshot["health"] = health

	return snapshot
}

func systemStats() map[string]any {
	stats := system.CollectStats()
	if _, ok := stats["timestamp"]; !ok {
		stats["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	}
	return stats
}

func (a *App) instanceOperationalCounts(ctx context.Context) (int, int) {
	if a == nil || a.Instances == nil {
		return 0, 0
	}
	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		return 0, 0
	}
	running := 0
	for _, instance := range instances {
		if instance.Status == domain.InstanceStatusRunning {
			running++
		}
	}
	return len(instances), running
}

func (a *App) queueOperationalState(ctx context.Context) (map[string]any, int, map[string]any) {
	queueSizes := map[string]any{
		"mqtt":       0,
		"cloud_sync": 0,
	}
	queueDetails := map[string]any{
		"mqtt": map[string]any{
			"pending":                0,
			"max_items":              0,
			"max_age_seconds":        0,
			"usage_pct":              0,
			"severity":               "normal",
			"oldest_pending_seconds": 0,
		},
		"cloud_sync": map[string]any{
			"pending":                0,
			"failed":                 0,
			"pending_total":          0,
			"capacity_hint":          0,
			"usage_pct":              0,
			"severity":               "normal",
			"oldest_pending_seconds": 0,
		},
	}
	oldestPendingAgeSeconds := 0
	retention := sqlite.CurrentRetentionSnapshot()

	mqttRetention, _ := retention["mqtt"].(map[string]any)
	mqttMaxItems := int(asInt64(mqttRetention["max_items"]))
	mqttMaxAgeSeconds := int(asInt64(mqttRetention["max_age_seconds"]))

	if a != nil && a.DB != nil && a.DB.MqttQueue != nil {
		items, err := a.DB.MqttQueue.ListPending(ctx, 0)
		if err == nil {
			queueSizes["mqtt"] = len(items)
			mqttOldestAge := 0
			if len(items) > 0 {
				age := int(time.Since(items[0].CreatedAt).Seconds())
				if age > oldestPendingAgeSeconds {
					oldestPendingAgeSeconds = age
				}
				mqttOldestAge = age
			}
			queueDetails["mqtt"] = map[string]any{
				"pending":                len(items),
				"max_items":              mqttMaxItems,
				"max_age_seconds":        mqttMaxAgeSeconds,
				"usage_pct":              percentOfInt(len(items), mqttMaxItems),
				"severity":               maxSeverity(severityFromPercent(percentOfInt(len(items), mqttMaxItems)), severityFromAgeSeconds(mqttOldestAge)),
				"oldest_pending_seconds": mqttOldestAge,
			}
		}
	}

	cloudSyncRetention, _ := retention["cloud_sync"].(map[string]any)
	cloudSyncCapacityHint := max(
		int(asInt64(cloudSyncRetention["keep_sent_items"])),
		int(asInt64(cloudSyncRetention["keep_failed_items"])),
	)

	if a != nil && a.CloudSync != nil {
		stats, err := a.CloudSync.Stats(ctx)
		if err == nil {
			pending := int(asInt64(stats["pending"]) + asInt64(stats["failed"]))
			queueSizes["cloud_sync"] = pending
			age := int(asInt64(stats["oldest_pending_age_seconds"]))
			if age > oldestPendingAgeSeconds {
				oldestPendingAgeSeconds = age
			}
			queueDetails["cloud_sync"] = map[string]any{
				"pending":                int(asInt64(stats["pending"])),
				"failed":                 int(asInt64(stats["failed"])),
				"pending_total":          pending,
				"capacity_hint":          cloudSyncCapacityHint,
				"usage_pct":              percentOfInt(pending, cloudSyncCapacityHint),
				"severity":               maxSeverity(severityFromPercent(percentOfInt(pending, cloudSyncCapacityHint)), severityFromAgeSeconds(age)),
				"oldest_pending_seconds": age,
			}
		}
	}

	return queueSizes, oldestPendingAgeSeconds, queueDetails
}

func (a *App) cloudSyncOperationalState(ctx context.Context) map[string]any {
	result := map[string]any{
		"configured": false,
		"pending":    int64(0),
		"failed":     int64(0),
		"sent":       int64(0),
	}
	if a == nil || a.CloudSync == nil {
		return result
	}
	stats, err := a.CloudSync.Stats(ctx)
	if err != nil {
		return result
	}
	for key, value := range stats {
		result[key] = value
	}
	retention := sqlite.CurrentRetentionSnapshot()
	cloudSyncRetention, _ := retention["cloud_sync"].(map[string]any)
	pendingTotal := int(asInt64(result["pending"]) + asInt64(result["failed"]))
	capacityHint := max(
		int(asInt64(cloudSyncRetention["keep_sent_items"])),
		int(asInt64(cloudSyncRetention["keep_failed_items"])),
	)
	usagePct := percentOfInt(pendingTotal, capacityHint)
	result["pending_total"] = pendingTotal
	result["capacity_hint"] = capacityHint
	result["usage_pct"] = usagePct
	result["severity"] = maxSeverity(
		severityFromPercent(usagePct),
		severityFromAgeSeconds(int(asInt64(result["oldest_pending_age_seconds"]))),
	)
	return result
}

func (a *App) connectivitySnapshot(queueSizes map[string]any, cloudSync map[string]any, oldestPendingAgeSeconds int) connectivity.Snapshot {
	policy := a.RuntimeCfg.Connectivity.Normalize()
	cloudSyncPending := int(asInt64(cloudSync["pending"]) + asInt64(cloudSync["failed"]))
	cloudSyncFailed := int(asInt64(cloudSync["failed"]))
	mqttQueueDepth, _ := queueSizes["mqtt"].(int)
	mqttConnected := a != nil && a.MQTT != nil && a.MQTT.IsConnected()
	cloudSyncConfigured := a != nil && a.CloudSync != nil && a.CloudSync.Configured()

	return policy.Evaluate(
		mqttConnected,
		cloudSyncConfigured,
		cloudSyncPending,
		cloudSyncFailed,
		mqttQueueDepth,
		oldestPendingAgeSeconds,
	)
}

func asInt64(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func operationalHealthSnapshot(snapshot connectivity.Snapshot, queueDetails map[string]any, oldestPendingAgeSeconds int) map[string]any {
	reasons := append([]string{}, snapshot.Reasons...)
	severity := severityFromConnectivityStatus(snapshot.Status)

	mqtt, _ := queueDetails["mqtt"].(map[string]any)
	cloudSync, _ := queueDetails["cloud_sync"].(map[string]any)
	mqttSeverity := asString(mqtt["severity"], "normal")
	cloudSyncSeverity := asString(cloudSync["severity"], "normal")
	ageSeverity := severityFromAgeSeconds(oldestPendingAgeSeconds)

	severity = maxSeverity(severity, mqttSeverity)
	severity = maxSeverity(severity, cloudSyncSeverity)
	severity = maxSeverity(severity, ageSeverity)

	if severityRank(mqttSeverity) > 0 {
		reasons = append(reasons, "mqtt_queue_pressure")
	}
	if severityRank(cloudSyncSeverity) > 0 {
		reasons = append(reasons, "cloud_sync_backlog")
	}
	if severityRank(ageSeverity) > 0 {
		reasons = append(reasons, "oldest_pending_age")
	}

	return map[string]any{
		"severity":               severity,
		"reasons":                uniqueStrings(reasons),
		"oldest_pending_seconds": oldestPendingAgeSeconds,
		"mqtt_queue":             mqtt,
		"cloud_sync":             cloudSync,
	}
}

func severityFromConnectivityStatus(status string) string {
	switch status {
	case "offline":
		return "critical"
	case "degraded":
		return "warning"
	default:
		return "normal"
	}
}

func severityFromPercent(value int) string {
	switch {
	case value >= 85:
		return "critical"
	case value >= 60:
		return "warning"
	default:
		return "normal"
	}
}

func severityFromAgeSeconds(value int) string {
	switch {
	case value >= 1800:
		return "critical"
	case value >= 600:
		return "warning"
	default:
		return "normal"
	}
}

func severityRank(value string) int {
	switch value {
	case "critical":
		return 2
	case "warning":
		return 1
	default:
		return 0
	}
}

func maxSeverity(values ...string) string {
	current := "normal"
	for _, value := range values {
		if severityRank(value) > severityRank(current) {
			current = value
		}
	}
	return current
}

func percentOfInt(value, limit int) int {
	if limit <= 0 || value <= 0 {
		return 0
	}
	return int(float64(value) / float64(limit) * 100)
}

func asString(value any, fallback string) string {
	if typed, ok := value.(string); ok && typed != "" {
		return typed
	}
	return fallback
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
