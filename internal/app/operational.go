package app

import (
	"context"
	"runtime"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/system"
)

func (a *App) collectOperationalSnapshot(ctx context.Context) map[string]any {
	snapshot := map[string]any{}
	for key, value := range systemStats() {
		snapshot[key] = value
	}

	activeInstances, runningInstances := a.instanceOperationalCounts(ctx)
	queueSizes, oldestPendingAgeSeconds := a.queueOperationalState(ctx)
	cloudSync := a.cloudSyncOperationalState(ctx)

	status := "online"
	if failed, _ := cloudSync["failed"].(int64); failed > 0 {
		status = "degraded"
	}
	if pending, ok := queueSizes["cloud_sync"].(int); ok && pending > 25 {
		status = "degraded"
	}
	if pending, ok := queueSizes["mqtt"].(int); ok && pending > 10 {
		status = "degraded"
	}

	snapshot["status"] = status
	snapshot["edge_version"] = envOrDefault("SPARKEDGE_VERSION", "go-dev")
	snapshot["os"] = runtime.GOOS
	snapshot["hardware"] = envOrDefault("SPARKEDGE_HARDWARE", runtime.GOARCH)
	snapshot["active_instances"] = map[string]any{
		"total":   activeInstances,
		"running": runningInstances,
	}
	snapshot["queue_sizes"] = queueSizes
	snapshot["oldest_pending_age_seconds"] = oldestPendingAgeSeconds
	snapshot["connectivity"] = map[string]any{
		"mqtt_connected":         a != nil && a.MQTT != nil && a.MQTT.IsConnected(),
		"cloud_sync_configured":  a != nil && a.CloudSync != nil && a.CloudSync.Configured(),
		"cloud_sync_queue_depth": queueSizes["cloud_sync"],
	}
	snapshot["cloud_sync"] = cloudSync

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

func (a *App) queueOperationalState(ctx context.Context) (map[string]any, int) {
	queueSizes := map[string]any{
		"mqtt":       0,
		"cloud_sync": 0,
	}
	oldestPendingAgeSeconds := 0

	if a != nil && a.DB != nil && a.DB.MqttQueue != nil {
		items, err := a.DB.MqttQueue.ListPending(ctx, 0)
		if err == nil {
			queueSizes["mqtt"] = len(items)
			if len(items) > 0 {
				age := int(time.Since(items[0].CreatedAt).Seconds())
				if age > oldestPendingAgeSeconds {
					oldestPendingAgeSeconds = age
				}
			}
		}
	}

	if a != nil && a.CloudSync != nil {
		stats, err := a.CloudSync.Stats(ctx)
		if err == nil {
			pending := int(asInt64(stats["pending"]) + asInt64(stats["failed"]))
			queueSizes["cloud_sync"] = pending
			if age := int(asInt64(stats["oldest_pending_age_seconds"])); age > oldestPendingAgeSeconds {
				oldestPendingAgeSeconds = age
			}
		}
	}

	return queueSizes, oldestPendingAgeSeconds
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
	return result
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
