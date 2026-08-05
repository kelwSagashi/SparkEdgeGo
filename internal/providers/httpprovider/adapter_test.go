package httpprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestBearerAdapterSendsJSONPayload(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/measurements" {
			t.Fatalf("expected /measurements, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Device") != "edge-1" {
			t.Fatalf("expected server header to be forwarded")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	adapter, err := New(StrategyBearer, providers.Config{
		Server:      map[string]any{"headers": map[string]any{"X-Device": "edge-1"}},
		Resource:    map[string]any{"config": map[string]any{"baseUrl": server.URL}},
		Operation:   map[string]any{"config": map[string]any{"method": "POST", "path": "/measurements"}},
		Credentials: map[string]any{"data": map[string]any{"token": "secret-token"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
	if received["temperature"] != float64(42) {
		t.Fatalf("unexpected payload %#v", received)
	}
}

func TestAPIKeyAdapterCanUseQueryParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != "abc123" {
			t.Fatalf("expected api key query param, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	registry := providers.NewRegistry()
	Register(registry)
	adapter, ok, err := registry.Create(StrategyAPIKey, providers.Config{
		Resource:    map[string]any{"config": map[string]any{"baseUrl": server.URL}},
		Operation:   map[string]any{"config": map[string]any{"method": "GET", "path": "/health"}},
		Credentials: map[string]any{"data": map[string]any{"key": "token", "value": "abc123", "in": "query"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected api key adapter to be registered")
	}
	if err := adapter.Test(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}
