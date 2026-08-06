package system

import (
	"runtime"
	"time"
)

var startedAt = time.Now().UTC()

func CollectStats() map[string]any {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"cpu":       0,
		"memory":    float64(mem.Alloc) / 1024 / 1024,
		"uptime":    int(time.Since(startedAt).Seconds()),
		"disk":      0,
		"network": map[string]any{
			"latency": nil,
		},
	}
}
