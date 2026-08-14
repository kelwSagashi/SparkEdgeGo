package connectivity

import "strings"

type State string

const (
	StateHealthy      State = "healthy"
	StateIntermittent State = "intermittent"
	StateDegraded     State = "degraded"
	StateOffline      State = "offline"
)

type Policy struct {
	IntermittentPendingAgeSeconds   int `json:"intermittent_pending_age_seconds" yaml:"intermittent_pending_age_seconds"`
	DegradedPendingAgeSeconds       int `json:"degraded_pending_age_seconds" yaml:"degraded_pending_age_seconds"`
	IntermittentCloudSyncQueueDepth int `json:"intermittent_cloud_sync_queue_depth" yaml:"intermittent_cloud_sync_queue_depth"`
	DegradedCloudSyncQueueDepth     int `json:"degraded_cloud_sync_queue_depth" yaml:"degraded_cloud_sync_queue_depth"`
	DegradedMQTTQueueDepth          int `json:"degraded_mqtt_queue_depth" yaml:"degraded_mqtt_queue_depth"`
	HeartbeatHealthySeconds         int `json:"heartbeat_healthy_seconds" yaml:"heartbeat_healthy_seconds"`
	HeartbeatDegradedSeconds        int `json:"heartbeat_degraded_seconds" yaml:"heartbeat_degraded_seconds"`
	StatsHealthySeconds             int `json:"stats_healthy_seconds" yaml:"stats_healthy_seconds"`
	StatsDegradedSeconds            int `json:"stats_degraded_seconds" yaml:"stats_degraded_seconds"`
}

type Snapshot struct {
	Mode                     State    `json:"mode"`
	Status                   string   `json:"status"`
	Reasons                  []string `json:"reasons"`
	HeartbeatIntervalSeconds int      `json:"heartbeat_interval_seconds"`
	StatsIntervalSeconds     int      `json:"stats_interval_seconds"`
}

func DefaultPolicy() Policy {
	return Policy{
		IntermittentPendingAgeSeconds:   120,
		DegradedPendingAgeSeconds:       600,
		IntermittentCloudSyncQueueDepth: 5,
		DegradedCloudSyncQueueDepth:     25,
		DegradedMQTTQueueDepth:          10,
		HeartbeatHealthySeconds:         30,
		HeartbeatDegradedSeconds:        90,
		StatsHealthySeconds:             120,
		StatsDegradedSeconds:            300,
	}
}

func (p Policy) Normalize() Policy {
	defaults := DefaultPolicy()
	if p.IntermittentPendingAgeSeconds <= 0 {
		p.IntermittentPendingAgeSeconds = defaults.IntermittentPendingAgeSeconds
	}
	if p.DegradedPendingAgeSeconds <= 0 {
		p.DegradedPendingAgeSeconds = defaults.DegradedPendingAgeSeconds
	}
	if p.IntermittentCloudSyncQueueDepth <= 0 {
		p.IntermittentCloudSyncQueueDepth = defaults.IntermittentCloudSyncQueueDepth
	}
	if p.DegradedCloudSyncQueueDepth <= 0 {
		p.DegradedCloudSyncQueueDepth = defaults.DegradedCloudSyncQueueDepth
	}
	if p.DegradedMQTTQueueDepth <= 0 {
		p.DegradedMQTTQueueDepth = defaults.DegradedMQTTQueueDepth
	}
	if p.HeartbeatHealthySeconds <= 0 {
		p.HeartbeatHealthySeconds = defaults.HeartbeatHealthySeconds
	}
	if p.HeartbeatDegradedSeconds <= 0 {
		p.HeartbeatDegradedSeconds = defaults.HeartbeatDegradedSeconds
	}
	if p.StatsHealthySeconds <= 0 {
		p.StatsHealthySeconds = defaults.StatsHealthySeconds
	}
	if p.StatsDegradedSeconds <= 0 {
		p.StatsDegradedSeconds = defaults.StatsDegradedSeconds
	}
	return p
}

func (p Policy) Evaluate(
	mqttConnected bool,
	cloudSyncConfigured bool,
	cloudSyncPending int,
	cloudSyncFailed int,
	mqttQueueDepth int,
	oldestPendingAgeSeconds int,
) Snapshot {
	p = p.Normalize()
	reasons := make([]string, 0, 4)

	if !mqttConnected {
		return Snapshot{
			Mode:                     StateOffline,
			Status:                   "offline",
			Reasons:                  []string{"mqtt_disconnected"},
			HeartbeatIntervalSeconds: p.HeartbeatDegradedSeconds,
			StatsIntervalSeconds:     p.StatsDegradedSeconds,
		}
	}

	if cloudSyncConfigured {
		if cloudSyncFailed > 0 {
			reasons = append(reasons, "cloud_sync_failed_items")
		}
		if cloudSyncPending >= p.DegradedCloudSyncQueueDepth {
			reasons = append(reasons, "cloud_sync_queue_high")
		} else if cloudSyncPending >= p.IntermittentCloudSyncQueueDepth {
			reasons = append(reasons, "cloud_sync_queue_growing")
		}
	}

	if mqttQueueDepth >= p.DegradedMQTTQueueDepth {
		reasons = append(reasons, "mqtt_queue_high")
	}

	if oldestPendingAgeSeconds >= p.DegradedPendingAgeSeconds {
		reasons = append(reasons, "pending_age_critical")
	} else if oldestPendingAgeSeconds >= p.IntermittentPendingAgeSeconds {
		reasons = append(reasons, "pending_age_elevated")
	}

	mode := StateHealthy
	if containsReason(reasons, "cloud_sync_failed_items") ||
		containsReason(reasons, "cloud_sync_queue_high") ||
		containsReason(reasons, "mqtt_queue_high") ||
		containsReason(reasons, "pending_age_critical") {
		mode = StateDegraded
	} else if len(reasons) > 0 {
		mode = StateIntermittent
	}

	status := "online"
	switch mode {
	case StateIntermittent:
		status = "intermittent"
	case StateDegraded:
		status = "degraded"
	}

	heartbeatInterval := p.HeartbeatHealthySeconds
	statsInterval := p.StatsHealthySeconds
	if mode != StateHealthy {
		heartbeatInterval = p.HeartbeatDegradedSeconds
		statsInterval = p.StatsDegradedSeconds
	}

	return Snapshot{
		Mode:                     mode,
		Status:                   status,
		Reasons:                  uniqueReasons(reasons),
		HeartbeatIntervalSeconds: heartbeatInterval,
		StatsIntervalSeconds:     statsInterval,
	}
}

func containsReason(reasons []string, target string) bool {
	for _, reason := range reasons {
		if strings.EqualFold(reason, target) {
			return true
		}
	}
	return false
}

func uniqueReasons(reasons []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		normalized := strings.TrimSpace(strings.ToLower(reason))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
