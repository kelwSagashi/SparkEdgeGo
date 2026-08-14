package cloudsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type QueueStore interface {
	Enqueue(context.Context, string, int, map[string]any) (domain.CloudSyncItem, error)
	FindByID(context.Context, string) (domain.CloudSyncItem, error)
	ListPending(context.Context, int) ([]domain.CloudSyncItem, error)
	ListRecent(context.Context, int) ([]domain.CloudSyncItem, error)
	MarkSent(context.Context, string) error
	MarkFailed(context.Context, string, string, *time.Time) error
	Delete(context.Context, string) error
	Stats(context.Context) (map[string]any, error)
}

type Config struct {
	BaseURL   string
	EdgeID    string
	SyncToken string
	Enabled   bool
}

type Service struct {
	queue  QueueStore
	client *http.Client
	config Config
}

func NewService(queue QueueStore, cfg Config) *Service {
	return &Service{
		queue:  queue,
		client: &http.Client{Timeout: 20 * time.Second},
		config: cfg,
	}
}

func (s *Service) Configured() bool {
	return s != nil && s.queue != nil && s.config.Enabled && strings.TrimSpace(s.config.BaseURL) != "" && strings.TrimSpace(s.config.EdgeID) != "" && strings.TrimSpace(s.config.SyncToken) != ""
}

func (s *Service) EnqueueEvent(ctx context.Context, eventType string, priority int, payload map[string]any) (domain.CloudSyncItem, error) {
	if s == nil || s.queue == nil {
		return domain.CloudSyncItem{}, errors.New("cloud sync queue is not configured")
	}
	return s.queue.Enqueue(ctx, eventType, priority, payload)
}

func (s *Service) EnqueueInstanceExecution(ctx context.Context, payload map[string]any) (domain.CloudSyncItem, error) {
	return s.EnqueueEvent(ctx, "instance_execution", 50, payload)
}

func (s *Service) EnqueueRemoteJob(ctx context.Context, payload map[string]any) (domain.CloudSyncItem, error) {
	return s.EnqueueEvent(ctx, "remote_job", 40, payload)
}

func (s *Service) Flush(ctx context.Context, limit int) (map[string]any, error) {
	if !s.Configured() {
		return map[string]any{"sent": 0, "failed": 0, "skipped": true}, nil
	}
	items, err := s.queue.ListPending(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]any{"sent": 0, "failed": 0, "items": 0}, nil
	}

	sent := 0
	failed := 0
	for _, item := range items {
		if err := s.sendItem(ctx, item); err != nil {
			failed++
			nextRetryAt := time.Now().UTC().Add(backoff(item.Attempts))
			_ = s.queue.MarkFailed(ctx, item.ID, err.Error(), &nextRetryAt)
			continue
		}
		sent++
		_ = s.queue.MarkSent(ctx, item.ID)
	}
	return map[string]any{"sent": sent, "failed": failed, "items": len(items)}, nil
}

func (s *Service) ListRecent(ctx context.Context, limit int) ([]domain.CloudSyncItem, error) {
	if s == nil || s.queue == nil {
		return nil, errors.New("cloud sync queue is not configured")
	}
	return s.queue.ListRecent(ctx, limit)
}

func (s *Service) RetryItem(ctx context.Context, id string) (map[string]any, error) {
	if s == nil || s.queue == nil {
		return nil, errors.New("cloud sync queue is not configured")
	}
	item, err := s.queue.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status == domain.CloudSyncSent {
		return map[string]any{
			"id":      item.ID,
			"status":  string(item.Status),
			"sent":    false,
			"message": "item already sent",
		}, nil
	}
	if !s.Configured() {
		return map[string]any{
			"id":      item.ID,
			"status":  string(item.Status),
			"sent":    false,
			"skipped": true,
			"message": "cloud sync is not configured",
		}, nil
	}
	if err := s.sendItem(ctx, item); err != nil {
		nextRetryAt := time.Now().UTC().Add(backoff(item.Attempts))
		_ = s.queue.MarkFailed(ctx, item.ID, err.Error(), &nextRetryAt)
		return map[string]any{
			"id":         item.ID,
			"status":     string(domain.CloudSyncFailed),
			"sent":       false,
			"last_error": err.Error(),
		}, nil
	}
	if err := s.queue.MarkSent(ctx, item.ID); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":      item.ID,
		"status":  string(domain.CloudSyncSent),
		"sent":    true,
		"message": "item sent successfully",
	}, nil
}

func (s *Service) DeleteItem(ctx context.Context, id string) error {
	if s == nil || s.queue == nil {
		return errors.New("cloud sync queue is not configured")
	}
	return s.queue.Delete(ctx, id)
}

func (s *Service) Stats(ctx context.Context) (map[string]any, error) {
	if s == nil || s.queue == nil {
		return nil, errors.New("cloud sync queue is not configured")
	}
	stats, err := s.queue.Stats(ctx)
	if err != nil {
		return nil, err
	}
	stats["configured"] = s.Configured()
	stats["base_url"] = s.config.BaseURL
	stats["edge_id"] = s.config.EdgeID
	if createdAt, ok := stats["oldest_pending_created_at"].(time.Time); ok && !createdAt.IsZero() {
		stats["oldest_pending_age_seconds"] = int64(time.Since(createdAt).Seconds())
	}
	stats["pending_total"] = asInt64(stats["pending"]) + asInt64(stats["failed"])
	return stats, nil
}

func (s *Service) sendItem(ctx context.Context, item domain.CloudSyncItem) error {
	endpoint := strings.TrimRight(s.config.BaseURL, "/") + "/edge-sync/events/batch"
	edgeID := s.config.EdgeID
	if edgeID == "" {
		if value, ok := item.Payload["edge_id"].(string); ok {
			edgeID = strings.TrimSpace(value)
		}
	}
	if edgeID == "" {
		return errors.New("cloud sync item missing edge_id")
	}
	payload := map[string]any{
		"edge_id": edgeID,
		"events": []map[string]any{
			{
				"message_id": item.Payload["message_id"],
				"type":       item.EventType,
				"payload":    item.Payload,
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-spark-token", s.config.SyncToken)
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("cloud sync failed with status %d", res.StatusCode)
	}
	return nil
}

func backoff(attempts int) time.Duration {
	if attempts < 1 {
		return 30 * time.Second
	}
	delay := time.Duration(attempts*attempts) * 30 * time.Second
	if delay > 30*time.Minute {
		return 30 * time.Minute
	}
	return delay
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
