package cloudsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	BaseURL           string
	EdgeID            string
	SyncToken         string
	Enabled           bool
	SchemaVersion     string
	MaxAttempts       int
	MaxBatchSize      int
	HighPriorityDelay time.Duration
	LowPriorityDelay  time.Duration
}

type Service struct {
	queue  QueueStore
	client *http.Client
	config Config
}

func NewService(queue QueueStore, cfg Config) *Service {
	if strings.TrimSpace(cfg.SchemaVersion) == "" {
		cfg.SchemaVersion = "edge-cloud.v1"
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 12
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 50
	}
	if cfg.HighPriorityDelay <= 0 {
		cfg.HighPriorityDelay = 15 * time.Second
	}
	if cfg.LowPriorityDelay <= 0 {
		cfg.LowPriorityDelay = 45 * time.Second
	}
	return &Service{
		queue:  queue,
		client: &http.Client{Timeout: 20 * time.Second},
		config: cfg,
	}
}

func (s *Service) UpdateConfig(cfg Config) {
	if s == nil {
		return
	}
	if strings.TrimSpace(cfg.SchemaVersion) == "" {
		cfg.SchemaVersion = "edge-cloud.v1"
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 12
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 50
	}
	if cfg.HighPriorityDelay <= 0 {
		cfg.HighPriorityDelay = 15 * time.Second
	}
	if cfg.LowPriorityDelay <= 0 {
		cfg.LowPriorityDelay = 45 * time.Second
	}
	s.config = cfg
}

func (s *Service) Configured() bool {
	return s != nil && s.queue != nil && s.config.Enabled && strings.TrimSpace(s.config.BaseURL) != "" && strings.TrimSpace(s.config.SyncToken) != ""
}

func (s *Service) EnqueueEvent(ctx context.Context, eventType string, priority int, payload map[string]any) (domain.CloudSyncItem, error) {
	if s == nil || s.queue == nil {
		return domain.CloudSyncItem{}, errors.New("cloud sync queue is not configured")
	}
	normalized := s.normalizePayload(eventType, payload)
	if normalized == nil {
		normalized = map[string]any{}
	}
	if priority <= 0 {
		priority = defaultPriority(eventType)
	}
	return s.queue.Enqueue(ctx, eventType, priority, normalized)
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
	if limit <= 0 || limit > s.config.MaxBatchSize {
		limit = s.config.MaxBatchSize
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
	for _, batch := range s.groupItemsForBatch(items) {
		sendErr := s.sendBatch(ctx, batch)
		if sendErr == nil || isConflict(sendErr) {
			for _, item := range batch {
				sent++
				_ = s.queue.MarkSent(ctx, item.ID)
			}
			continue
		}

		for _, item := range batch {
			nextRetryAt := time.Now().UTC().Add(backoff(item.Attempts+1, item.Priority, s.config.HighPriorityDelay, s.config.LowPriorityDelay))
			if item.Attempts+1 >= s.config.MaxAttempts && !isRetryableError(sendErr) {
				nextRetryAt = time.Now().UTC().Add(24 * time.Hour)
			}
			_ = s.queue.MarkFailed(ctx, item.ID, sendErr.Error(), &nextRetryAt)
			failed++
		}
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
		nextRetryAt := time.Now().UTC().Add(backoff(item.Attempts+1, item.Priority, s.config.HighPriorityDelay, s.config.LowPriorityDelay))
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
	return s.sendBatch(ctx, []domain.CloudSyncItem{item})
}

func (s *Service) sendBatch(ctx context.Context, items []domain.CloudSyncItem) error {
	if len(items) == 0 {
		return nil
	}
	endpoint := strings.TrimRight(s.config.BaseURL, "/") + "/edge-sync/events/batch"
	edgeID := s.resolveEdgeID(items[0])
	if edgeID == "" {
		return errors.New("cloud sync item missing edge_id")
	}
	events := make([]map[string]any, 0, len(items))
	for _, item := range items {
		events = append(events, map[string]any{
			"message_id": item.Payload["message_id"],
			"type":       item.EventType,
			"payload":    item.Payload,
			"priority":   item.Priority,
			"attempts":   item.Attempts,
			"created_at": item.CreatedAt.Format(time.RFC3339),
		})
	}
	payload := map[string]any{
		"edge_id": edgeID,
		"events":  events,
		"batch": map[string]any{
			"items":            len(items),
			"schema_version":   s.config.SchemaVersion,
			"highest_priority": highestPriority(items),
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
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return &syncHTTPError{
			StatusCode: res.StatusCode,
			Body:       strings.TrimSpace(string(body)),
		}
	}
	return nil
}

func (s *Service) resolveEdgeID(item domain.CloudSyncItem) string {
	edgeID := strings.TrimSpace(s.config.EdgeID)
	if edgeID != "" {
		return edgeID
	}
	if value, ok := item.Payload["edge_id"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func (s *Service) groupItemsForBatch(items []domain.CloudSyncItem) [][]domain.CloudSyncItem {
	if len(items) == 0 {
		return nil
	}
	buckets := map[string][]domain.CloudSyncItem{}
	order := make([]string, 0)
	for _, item := range items {
		edgeID := s.resolveEdgeID(item)
		if edgeID == "" {
			edgeID = "__missing__"
		}
		if _, exists := buckets[edgeID]; !exists {
			order = append(order, edgeID)
		}
		buckets[edgeID] = append(buckets[edgeID], item)
	}
	result := make([][]domain.CloudSyncItem, 0, len(order))
	for _, edgeID := range order {
		result = append(result, buckets[edgeID])
	}
	return result
}

func backoff(attempts int, priority int, highPriorityDelay time.Duration, lowPriorityDelay time.Duration) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	base := lowPriorityDelay
	if priority >= 80 {
		base = highPriorityDelay
	} else if priority >= 50 {
		base = 25 * time.Second
	}
	delay := time.Duration(attempts*attempts) * base
	if delay > 30*time.Minute {
		return 30 * time.Minute
	}
	return delay
}

func (s *Service) normalizePayload(eventType string, payload map[string]any) map[string]any {
	if payload == nil {
		payload = map[string]any{}
	}
	if value, ok := payload["edge_id"].(string); !ok || strings.TrimSpace(value) == "" {
		payload["edge_id"] = s.config.EdgeID
	}
	if value, ok := payload["type"].(string); !ok || strings.TrimSpace(value) == "" {
		payload["type"] = eventType
	}
	if value, ok := payload["schema_version"].(string); !ok || strings.TrimSpace(value) == "" {
		payload["schema_version"] = s.config.SchemaVersion
	}
	occurredAt, ok := payload["occurred_at"].(string)
	if !ok || strings.TrimSpace(occurredAt) == "" {
		now := time.Now().UTC().Format(time.RFC3339)
		payload["occurred_at"] = now
		if _, exists := payload["timestamp"]; !exists {
			payload["timestamp"] = now
		}
	}
	messageID, _ := payload["message_id"].(string)
	if strings.TrimSpace(messageID) == "" {
		payload["message_id"] = buildMessageID(payload)
	}
	return payload
}

func buildMessageID(payload map[string]any) string {
	edgeID, _ := payload["edge_id"].(string)
	eventType, _ := payload["type"].(string)
	occurredAt, _ := payload["occurred_at"].(string)
	return fmt.Sprintf("%s:%s:%d:%s", strings.TrimSpace(edgeID), strings.TrimSpace(eventType), time.Now().UTC().UnixNano(), strings.TrimSpace(occurredAt))
}

func defaultPriority(eventType string) int {
	switch strings.TrimSpace(strings.ToLower(eventType)) {
	case "edge_connection", "command_response":
		return 90
	case "remote_job":
		return 80
	case "instance_execution", "heartbeat", "context":
		return 60
	case "stats", "metrics", "telemetry", "meta":
		return 40
	default:
		return 50
	}
}

func isRetryableError(err error) bool {
	var syncErr *syncHTTPError
	if errors.As(err, &syncErr) {
		if syncErr.StatusCode == http.StatusRequestTimeout || syncErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
		return syncErr.StatusCode >= 500
	}
	return true
}

func isConflict(err error) bool {
	var syncErr *syncHTTPError
	return errors.As(err, &syncErr) && syncErr.StatusCode == http.StatusConflict
}

func highestPriority(items []domain.CloudSyncItem) int {
	current := 0
	for _, item := range items {
		if item.Priority > current {
			current = item.Priority
		}
	}
	return current
}

type syncHTTPError struct {
	StatusCode int
	Body       string
}

func (e *syncHTTPError) Error() string {
	if strings.TrimSpace(e.Body) == "" {
		return fmt.Sprintf("cloud sync failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("cloud sync failed with status %d: %s", e.StatusCode, e.Body)
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
