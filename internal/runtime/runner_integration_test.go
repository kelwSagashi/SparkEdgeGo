package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/httpprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestRunnerTriggerDispatchesMappedPayloadToHTTPProvider(t *testing.T) {
	var (
		mu          sync.Mutex
		requestPath string
		payload     map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		mu.Lock()
		requestPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	registry := providers.NewRegistry()
	httpprovider.Register(registry)

	runner := NewRunner(Dependencies{
		Sparkit:            &fakeSparkitExecutor{},
		Providers:          registry,
		ResourceOperations: fakeResourceOperationsRepo{targets: map[string]sqlite.OperationTarget{"operation-1": httpTarget(server.URL, "/ingest")}},
	})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		ExecutionID: "exec-1",
		Instance: domain.Instance{
			ID:               "instance-1",
			Name:             "Temperature Monitor",
			ScriptParameters: map[string]any{"mode": "full"},
		},
		Script: domain.DownloadedScript{
			ID:        "script-1",
			Name:      "sample-script",
			LocalPath: ".",
			MainFile:  "main.py",
		},
		Destinations: []domain.InstanceDestinationWithMapping{
			{
				Destination: domain.InstanceDestination{
					ID:                  "destination-1",
					ResourceOperationID: "operation-1",
					Enabled:             true,
				},
				Mapping: &domain.DataMapping{
					Mapping: map[string]any{
						"temperature": "$.temperature",
						"status":      "{{ok || 'offline'}}",
					},
					PayloadTemplate: map[string]any{
						"instance": "{{instance.name}}",
						"script":   "{{script_metadata.name}}",
					},
					CustomFields: []domain.MappingCustomField{
						{Key: "summary", Value: "temp={{temperature}}"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionSuccess {
		t.Fatalf("expected success, got %#v", result)
	}
	if !result.DestinationSent || result.FallbackUsed {
		t.Fatalf("unexpected dispatch flags: %#v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if requestPath != "/ingest" {
		t.Fatalf("expected request path /ingest, got %q", requestPath)
	}
	if payload["instance"] != "Temperature Monitor" || payload["script"] != "sample-script" {
		t.Fatalf("unexpected template payload: %#v", payload)
	}
	if payload["temperature"] != float64(42) || payload["status"] != true || payload["summary"] != "temp=42" {
		t.Fatalf("unexpected mapped payload: %#v", payload)
	}
}

func TestRunnerFallbackAndFlushRetry(t *testing.T) {
	var (
		mu       sync.Mutex
		failMode = true
		hits     int
		payload  map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		mu.Lock()
		hits++
		currentFailMode := failMode
		if !currentFailMode {
			_ = json.NewDecoder(r.Body).Decode(&payload)
		}
		mu.Unlock()
		if currentFailMode {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	registry := providers.NewRegistry()
	httpprovider.Register(registry)

	fallback := newFakeFallbackStore()
	destinations := fakeDestinationsRepo{
		items: map[string]domain.InstanceDestination{
			"destination-1": {
				ID:                  "destination-1",
				ResourceOperationID: "operation-1",
				Enabled:             true,
			},
		},
	}
	runner := NewRunner(Dependencies{
		Sparkit:            &fakeSparkitExecutor{},
		Providers:          registry,
		ResourceOperations: fakeResourceOperationsRepo{targets: map[string]sqlite.OperationTarget{"operation-1": httpTarget(server.URL, "/retry")}},
		Fallback:           fallback,
		Destinations:       destinations,
	})

	result, err := runner.Trigger(context.Background(), TriggerRequest{
		ExecutionID: "exec-2",
		Instance: domain.Instance{
			ID:              "instance-1",
			Name:            "Retry Edge",
			FallbackEnabled: true,
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
		Destinations: []domain.InstanceDestinationWithMapping{
			{
				Destination: destinations.items["destination-1"],
				Mapping: &domain.DataMapping{
					Mapping: map[string]any{"temperature": "$.temperature"},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.ExecutionSuccess || !result.FallbackUsed {
		t.Fatalf("expected success with fallback, got %#v", result)
	}
	items, err := fallback.ListPending(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one fallback item, got %#v", items)
	}

	mu.Lock()
	failMode = false
	mu.Unlock()

	sent, err := runner.FlushFallback(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if sent != 1 {
		t.Fatalf("expected one retried item sent, got %d", sent)
	}
	item, err := fallback.FindByID(context.Background(), items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != domain.FallbackSent {
		t.Fatalf("expected fallback item sent, got %#v", item)
	}
	mu.Lock()
	defer mu.Unlock()
	if hits < 2 {
		t.Fatalf("expected retry request to happen, got %d hits", hits)
	}
	if payload["temperature"] != float64(42) {
		t.Fatalf("unexpected retried payload %#v", payload)
	}
}

func TestRunnerCircuitBreakerBlocksDestinationAfterThreshold(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	registry := providers.NewRegistry()
	httpprovider.Register(registry)

	runner := NewRunner(Dependencies{
		Sparkit:            &fakeSparkitExecutor{},
		Providers:          registry,
		ResourceOperations: fakeResourceOperationsRepo{targets: map[string]sqlite.OperationTarget{"operation-1": httpTarget(server.URL, "/breaker")}},
		CircuitBreakers:    newFakeCircuitBreakerStore(),
	})

	request := TriggerRequest{
		ExecutionID: "exec-breaker",
		Instance: domain.Instance{
			ID:              "instance-1",
			Name:            "Breaker Edge",
			FallbackEnabled: false,
		},
		Script: domain.DownloadedScript{
			LocalPath: ".",
			MainFile:  "main.py",
		},
		Destinations: []domain.InstanceDestinationWithMapping{
			{
				Destination: domain.InstanceDestination{
					ID:                  "destination-1",
					ResourceOperationID: "operation-1",
					Enabled:             true,
					RetryPolicy: domain.RetryPolicy{
						MaxRetries:                    1,
						RetryInterval:                 1,
						TimeoutSeconds:                2,
						CircuitBreakerThreshold:       1,
						CircuitBreakerCooldownSeconds: 60,
					},
				},
				Mapping: &domain.DataMapping{
					Mapping: map[string]any{"temperature": "$.temperature"},
				},
			},
		},
	}

	firstResult, err := runner.Trigger(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if firstResult.Status != domain.ExecutionFailed {
		t.Fatalf("expected first execution to fail, got %#v", firstResult)
	}

	secondResult, err := runner.Trigger(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if secondResult.Status != domain.ExecutionFailed {
		t.Fatalf("expected second execution to fail due to breaker, got %#v", secondResult)
	}
	if secondResult.Error != "destination-1: circuit breaker open for destination" {
		t.Fatalf("expected breaker message, got %#v", secondResult.Error)
	}
}

type fakeCircuitBreakerStore struct {
	items map[string]domain.CircuitBreakerState
}

func newFakeCircuitBreakerStore() *fakeCircuitBreakerStore {
	return &fakeCircuitBreakerStore{items: map[string]domain.CircuitBreakerState{}}
}

func (f *fakeCircuitBreakerStore) GetByDestination(_ context.Context, destinationID string) (domain.CircuitBreakerState, error) {
	item, ok := f.items[destinationID]
	if !ok {
		return domain.CircuitBreakerState{}, sqlite.ErrNotFound
	}
	return item, nil
}

func (f *fakeCircuitBreakerStore) Upsert(_ context.Context, state domain.CircuitBreakerState) (domain.CircuitBreakerState, error) {
	f.items[state.DestinationID] = state
	return state, nil
}

func (f *fakeCircuitBreakerStore) Delete(_ context.Context, destinationID string) error {
	delete(f.items, destinationID)
	return nil
}

type fakeResourceOperationsRepo struct {
	targets map[string]sqlite.OperationTarget
}

func (f fakeResourceOperationsRepo) ResolveTarget(_ context.Context, id string) (sqlite.OperationTarget, error) {
	target, ok := f.targets[id]
	if !ok {
		return sqlite.OperationTarget{}, sqlite.ErrNotFound
	}
	return target, nil
}

type fakeDestinationsRepo struct {
	items map[string]domain.InstanceDestination
}

func (f fakeDestinationsRepo) FindByID(_ context.Context, id string) (domain.InstanceDestination, error) {
	item, ok := f.items[id]
	if !ok {
		return domain.InstanceDestination{}, sqlite.ErrNotFound
	}
	return item, nil
}

type fakeFallbackStore struct {
	mu    sync.Mutex
	items map[string]domain.LocalFallbackItem
}

func newFakeFallbackStore() *fakeFallbackStore {
	return &fakeFallbackStore{items: map[string]domain.LocalFallbackItem{}}
}

func (f *fakeFallbackStore) Create(_ context.Context, params sqlite.CreateLocalFallbackParams) (domain.LocalFallbackItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now().UTC()
	item := domain.LocalFallbackItem{
		ID:            "fallback-" + params.DestinationID,
		InstanceID:    params.InstanceID,
		DestinationID: params.DestinationID,
		ExecutionID:   params.ExecutionID,
		Status:        domain.FallbackPending,
		Payload:       params.Payload,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeFallbackStore) ListPending(_ context.Context) ([]domain.LocalFallbackItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := make([]domain.LocalFallbackItem, 0, len(f.items))
	for _, item := range f.items {
		if item.Status == domain.FallbackPending {
			items = append(items, item)
		}
	}
	return items, nil
}

func (f *fakeFallbackStore) MarkAsSending(_ context.Context, id string) (domain.LocalFallbackItem, error) {
	return f.update(id, func(item domain.LocalFallbackItem) domain.LocalFallbackItem {
		item.Status = domain.FallbackSending
		return item
	})
}

func (f *fakeFallbackStore) MarkAsSent(_ context.Context, id string) (domain.LocalFallbackItem, error) {
	return f.update(id, func(item domain.LocalFallbackItem) domain.LocalFallbackItem {
		item.Status = domain.FallbackSent
		return item
	})
}

func (f *fakeFallbackStore) IncrementRetry(_ context.Context, id string, lastError string) (domain.LocalFallbackItem, error) {
	return f.update(id, func(item domain.LocalFallbackItem) domain.LocalFallbackItem {
		now := time.Now().UTC()
		item.Status = domain.FallbackPending
		item.RetryCount++
		item.LastRetryAt = &now
		item.LastError = lastError
		return item
	})
}

func (f *fakeFallbackStore) MarkAsFailed(_ context.Context, id string, lastError string) (domain.LocalFallbackItem, error) {
	return f.update(id, func(item domain.LocalFallbackItem) domain.LocalFallbackItem {
		item.Status = domain.FallbackFailed
		item.LastError = lastError
		return item
	})
}

func (f *fakeFallbackStore) FindByID(_ context.Context, id string) (domain.LocalFallbackItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok {
		return domain.LocalFallbackItem{}, sqlite.ErrNotFound
	}
	return item, nil
}

func (f *fakeFallbackStore) update(id string, update func(domain.LocalFallbackItem) domain.LocalFallbackItem) (domain.LocalFallbackItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok {
		return domain.LocalFallbackItem{}, sqlite.ErrNotFound
	}
	item = update(item)
	item.UpdatedAt = time.Now().UTC()
	f.items[id] = item
	return item, nil
}

func httpTarget(baseURL string, path string) sqlite.OperationTarget {
	return sqlite.OperationTarget{
		Server: domain.Server{
			ID:        "server-1",
			Name:      "HTTP Server",
			Type:      "http",
			DriverKey: "http",
			Headers:   map[string]any{},
		},
		Resource: domain.ServerResource{
			ID:       "resource-1",
			ServerID: "server-1",
			Name:     "HTTP Resource",
			Type:     "http",
			Config:   map[string]any{"baseUrl": baseURL},
		},
		Operation: domain.ResourceOperation{
			ID:         "operation-1",
			ResourceID: "resource-1",
			Name:       "Send",
			Type:       "http",
			Config:     map[string]any{"method": "POST", "path": path},
		},
	}
}
