package sqlite

import (
	"context"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/config"
	"gorm.io/gorm"
)

type RetentionPolicy struct {
	MQTTQueueMaxItems            int
	MQTTQueueMaxAge              time.Duration
	CloudSyncSentRetention       time.Duration
	CloudSyncFailedRetention     time.Duration
	CloudSyncKeepSentItems       int
	CloudSyncKeepFailedItems     int
	LocalFallbackSentRetention   time.Duration
	LocalFallbackFailedRetention time.Duration
	LocalFallbackKeepSentItems   int
	LocalFallbackKeepFailedItems int
}

var retentionPolicy = defaultRetentionPolicy()

func ConfigureRetention(section config.RetentionSection) {
	retentionPolicy = RetentionPolicy{
		MQTTQueueMaxItems:            max(100, section.MQTTQueueMaxItems),
		MQTTQueueMaxAge:              time.Duration(max(1, section.MQTTQueueMaxAgeHours)) * time.Hour,
		CloudSyncSentRetention:       time.Duration(max(1, section.CloudSyncSentRetentionHours)) * time.Hour,
		CloudSyncFailedRetention:     time.Duration(max(1, section.CloudSyncFailedRetentionHours)) * time.Hour,
		CloudSyncKeepSentItems:       max(100, section.CloudSyncKeepSentItems),
		CloudSyncKeepFailedItems:     max(100, section.CloudSyncKeepFailedItems),
		LocalFallbackSentRetention:   time.Duration(max(1, section.LocalFallbackSentRetentionHours)) * time.Hour,
		LocalFallbackFailedRetention: time.Duration(max(1, section.LocalFallbackFailedRetentionHours)) * time.Hour,
		LocalFallbackKeepSentItems:   max(100, section.LocalFallbackKeepSentItems),
		LocalFallbackKeepFailedItems: max(100, section.LocalFallbackKeepFailedItems),
	}
}

func currentRetentionPolicy() RetentionPolicy {
	return retentionPolicy
}

func CurrentRetentionSnapshot() map[string]any {
	policy := currentRetentionPolicy()
	return map[string]any{
		"mqtt": map[string]any{
			"max_items":       policy.MQTTQueueMaxItems,
			"max_age_hours":   int(policy.MQTTQueueMaxAge.Hours()),
			"max_age_seconds": int(policy.MQTTQueueMaxAge.Seconds()),
		},
		"cloud_sync": map[string]any{
			"sent_retention_hours":   int(policy.CloudSyncSentRetention.Hours()),
			"failed_retention_hours": int(policy.CloudSyncFailedRetention.Hours()),
			"keep_sent_items":        policy.CloudSyncKeepSentItems,
			"keep_failed_items":      policy.CloudSyncKeepFailedItems,
		},
		"local_fallback": map[string]any{
			"sent_retention_hours":   int(policy.LocalFallbackSentRetention.Hours()),
			"failed_retention_hours": int(policy.LocalFallbackFailedRetention.Hours()),
			"keep_sent_items":        policy.LocalFallbackKeepSentItems,
			"keep_failed_items":      policy.LocalFallbackKeepFailedItems,
		},
	}
}

func defaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MQTTQueueMaxItems:            1000,
		MQTTQueueMaxAge:              14 * 24 * time.Hour,
		CloudSyncSentRetention:       7 * 24 * time.Hour,
		CloudSyncFailedRetention:     30 * 24 * time.Hour,
		CloudSyncKeepSentItems:       1000,
		CloudSyncKeepFailedItems:     1000,
		LocalFallbackSentRetention:   7 * 24 * time.Hour,
		LocalFallbackFailedRetention: 30 * 24 * time.Hour,
		LocalFallbackKeepSentItems:   1000,
		LocalFallbackKeepFailedItems: 1000,
	}
}

func deleteOldestRows[T any](ctx context.Context, db *gorm.DB, model *T, where string, args []any, keep int) error {
	if keep < 0 {
		keep = 0
	}
	subquery := db.WithContext(ctx).
		Model(model).
		Where(where, args...).
		Order("created_at DESC").
		Offset(keep).
		Select("id")

	return db.WithContext(ctx).
		Where("id IN (?)", subquery).
		Delete(model).
		Error
}

func deleteRowsOlderThan[T any](ctx context.Context, db *gorm.DB, model *T, column string, before time.Time, where string, args []any) error {
	query := db.WithContext(ctx).Model(model).Where(column+" < ?", before)
	if where != "" {
		query = query.Where(where, args...)
	}
	return query.Delete(model).Error
}
