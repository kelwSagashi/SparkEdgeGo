package cloudsync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

func TestFlushBatchesItemsByEdgeID(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	requests := make([]map[string]any, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		mu.Lock()
		requests = append(requests, payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	queue := &fakeQueueStore{
		items: []domain.CloudSyncItem{
			newCloudSyncItem("1", "edge-a", "instance_execution", 60),
			newCloudSyncItem("2", "edge-a", "remote_job", 80),
			newCloudSyncItem("3", "edge-b", "instance_execution", 60),
		},
	}
	service := NewService(queue, Config{
		BaseURL:   server.URL,
		EdgeID:    "",
		SyncToken: "token",
		Enabled:   true,
	})

	result, err := service.Flush(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result["sent"] != 3 || result["failed"] != 0 {
		t.Fatalf("unexpected flush result %#v", result)
	}
	if len(queue.sent) != 3 {
		t.Fatalf("expected items to be marked as sent, got %#v", queue.sent)
	}
	if len(requests) != 2 {
		t.Fatalf("expected one batched request per edge id, got %d", len(requests))
	}
	firstEvents, _ := requests[0]["events"].([]any)
	secondEvents, _ := requests[1]["events"].([]any)
	if len(firstEvents) != 2 && len(secondEvents) != 2 {
		t.Fatalf("expected one batch to contain two events, got %#v", requests)
	}
}

func TestStatsExposePriorityAndTypeBreakdown(t *testing.T) {
	ctx := context.Background()
	queue := &fakeQueueStore{
		stats: map[string]any{
			"pending": int64(2),
			"failed":  int64(1),
			"sent":    int64(5),
			"by_status": map[string]any{
				"pending": int64(2),
				"failed":  int64(1),
				"sent":    int64(5),
			},
			"by_event_type": map[string]any{
				"instance_execution": int64(2),
				"remote_job":         int64(1),
			},
			"by_priority_band": map[string]any{
				"critical": int64(1),
				"high":     int64(2),
				"normal":   int64(0),
				"low":      int64(0),
			},
		},
	}
	service := NewService(queue, Config{
		BaseURL:   "https://example.test",
		EdgeID:    "edge-a",
		SyncToken: "token",
		Enabled:   true,
	})

	stats, err := service.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stats["pending_total"] != int64(3) {
		t.Fatalf("expected pending_total to be derived, got %#v", stats)
	}
	byType, ok := stats["by_event_type"].(map[string]any)
	if !ok || byType["instance_execution"] != int64(2) {
		t.Fatalf("expected event type breakdown, got %#v", stats)
	}
	byPriority, ok := stats["by_priority_band"].(map[string]any)
	if !ok || byPriority["critical"] != int64(1) {
		t.Fatalf("expected priority breakdown, got %#v", stats)
	}
}

type fakeQueueStore struct {
	items   []domain.CloudSyncItem
	stats   map[string]any
	sent    []string
	failed  []string
	deleted []string
}

func (f *fakeQueueStore) Enqueue(context.Context, string, int, map[string]any) (domain.CloudSyncItem, error) {
	return domain.CloudSyncItem{}, nil
}

func (f *fakeQueueStore) FindByID(_ context.Context, id string) (domain.CloudSyncItem, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.CloudSyncItem{}, nil
}

func (f *fakeQueueStore) ListPending(context.Context, int) ([]domain.CloudSyncItem, error) {
	return f.items, nil
}

func (f *fakeQueueStore) ListRecent(context.Context, int) ([]domain.CloudSyncItem, error) {
	return f.items, nil
}

func (f *fakeQueueStore) MarkSent(_ context.Context, id string) error {
	f.sent = append(f.sent, id)
	return nil
}

func (f *fakeQueueStore) MarkFailed(_ context.Context, id string, _ string, _ *time.Time) error {
	f.failed = append(f.failed, id)
	return nil
}

func (f *fakeQueueStore) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeQueueStore) Stats(context.Context) (map[string]any, error) {
	return f.stats, nil
}

func newCloudSyncItem(id string, edgeID string, eventType string, priority int) domain.CloudSyncItem {
	now := time.Now().UTC()
	return domain.CloudSyncItem{
		ID:        id,
		EventType: eventType,
		Priority:  priority,
		Status:    domain.CloudSyncPending,
		Payload: map[string]any{
			"message_id": id,
			"edge_id":    edgeID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
