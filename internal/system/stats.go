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
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"cpu_pct":        0,
		"memory_mb":      float64(mem.Alloc) / 1024 / 1024,
		"uptime_seconds": int(time.Since(startedAt).Seconds()),
		"goroutines":     runtime.NumGoroutine(),
		"disk_pct":       0,
		"network": map[string]any{
			"latency_ms": nil,
		},
	}
}
