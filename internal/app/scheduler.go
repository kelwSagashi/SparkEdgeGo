package app

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

func (a *App) StartScheduler(ctx context.Context, pollingInterval time.Duration) {
	if pollingInterval <= 0 {
		pollingInterval = 30 * time.Second
	}
	var processing atomic.Bool
	go func() {
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()
		a.pollScheduledInstances(ctx, &processing)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.pollScheduledInstances(ctx, &processing)
			}
		}
	}()
}

func (a *App) pollScheduledInstances(ctx context.Context, processing *atomic.Bool) {
	if a == nil || !processing.CompareAndSwap(false, true) {
		return
	}
	defer processing.Store(false)

	_, _ = a.Runtime.FlushFallback(ctx, 10)

	instances, err := a.Instances.ListActive(ctx)
	if err != nil {
		log.Printf("[InstanceScheduler] list active failed: %v", err)
		return
	}
	now := time.Now().UTC()
	for _, instance := range instances {
		if ctx.Err() != nil {
			return
		}
		if !shouldRunScheduledInstance(ctx, a, instance, now) {
			continue
		}
		go func(instanceID string) {
			if _, _, err := a.triggerInstance(ctx, instanceID, map[string]any{}, domain.TriggerInterval); err != nil {
				log.Printf("[InstanceScheduler] trigger %s failed: %v", instanceID, err)
			}
		}(instance.ID)
	}
}

func shouldRunScheduledInstance(ctx context.Context, a *App, instance domain.Instance, now time.Time) bool {
	if instance.Status == domain.InstanceStatusRunning {
		return false
	}
	if instance.TriggerType != domain.TriggerInterval && instance.TriggerType != domain.TriggerIntervalAndWebhook {
		return false
	}
	intervalSeconds := intFromMapDefault(instance.TriggerConfig, "interval_seconds", "intervalSeconds", 300)
	executions, err := a.Executions.ListByInstance(ctx, instance.ID, 1)
	if err != nil || len(executions) == 0 {
		return true
	}
	last := executions[0].CreatedAt
	if executions[0].StartedAt != nil {
		last = *executions[0].StartedAt
	}
	return now.Sub(last).Seconds() >= float64(intervalSeconds)
}

func intFromMapDefault(values map[string]any, snake string, camel string, fallback int) int {
	for _, key := range []string{snake, camel} {
		switch value := values[key].(type) {
		case int:
			if value > 0 {
				return value
			}
		case float64:
			if value > 0 {
				return int(value)
			}
		}
	}
	return fallback
}
