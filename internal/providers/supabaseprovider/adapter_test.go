package supabaseprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestAdapterInsertSendsPostgRESTRequest(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/v1/sensors" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("apikey") != "service-key" || r.Header.Get("Authorization") != "Bearer service-key" {
			t.Fatalf("missing supabase auth headers")
		}
		if r.Header.Get("Prefer") != "return=minimal" {
			t.Fatalf("unexpected prefer header %q", r.Header.Get("Prefer"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	adapter := newTestAdapter(t, server.URL, "insert")
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
	if received["temperature"] != float64(42) {
		t.Fatalf("unexpected payload %#v", received)
	}
}

func TestAdapterSelectMapsPayloadToFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("device_id") != "eq.edge-1" {
			t.Fatalf("unexpected query %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	adapter := newTestAdapter(t, server.URL, "select")
	if err := adapter.Send(context.Background(), map[string]any{"device_id": "edge-1"}); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterUpsertUsesMergeDuplicatesPreference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Prefer") != "resolution=merge-duplicates" {
			t.Fatalf("unexpected prefer header %q", r.Header.Get("Prefer"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	adapter := newTestAdapter(t, server.URL, "upsert")
	if err := adapter.Send(context.Background(), map[string]any{"id": "row-1"}); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterDiscoverReadsSchemaDefinitions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/" {
			t.Fatalf("unexpected schema path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"definitions": map[string]any{
				"sensors": map[string]any{
					"properties": map[string]any{
						"temperature": map[string]any{"type": "number"},
					},
				},
			},
		})
	}))
	defer server.Close()

	adapter := newTestAdapter(t, server.URL, "insert")
	resources, err := adapter.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 || resources[0].Name != "sensors" || len(resources[0].Fields) != 1 {
		t.Fatalf("unexpected resources %#v", resources)
	}
}

func newTestAdapter(t *testing.T, baseURL string, method string) *Adapter {
	t.Helper()
	adapter, err := New(providers.Config{
		Resource:    map[string]any{"config": map[string]any{"table": "sensors"}},
		Operation:   map[string]any{"config": map[string]any{"method": method}},
		Credentials: map[string]any{"data": map[string]any{"url": baseURL, "apiKey": "service-key"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}
