package firebaseprovider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestAdapterAddsDocumentByDefault(t *testing.T) {
	var received map[string]any
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/projects/project-1/databases/(default)/documents/sensors" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	adapter := newTestAdapter(t, "add", "")
	adapter.baseURL = api.URL
	adapter.tokenURL = tokenServer(t)
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
	fields := received["fields"].(map[string]any)
	temperature := fields["temperature"].(map[string]any)
	if temperature["doubleValue"] != float64(42) {
		t.Fatalf("unexpected firestore payload %#v", received)
	}
}

func TestAdapterSetsDocumentWhenDocIDIsConfigured(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/projects/project-1/databases/(default)/documents/sensors/latest" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	adapter := newTestAdapter(t, "set", "latest")
	adapter.baseURL = api.URL
	adapter.tokenURL = tokenServer(t)
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
}

func newTestAdapter(t *testing.T, operation string, docID string) *Adapter {
	t.Helper()
	adapter, err := New(providers.Config{
		Resource: map[string]any{"config": map[string]any{"collection": "sensors"}},
		Operation: map[string]any{"config": map[string]any{
			"operation": operation,
			"docId":     docID,
		}},
		Credentials: map[string]any{"data": map[string]any{
			"projectId":   "project-1",
			"clientEmail": "edge@example.iam.gserviceaccount.com",
			"privateKey":  testPrivateKey(t),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func tokenServer(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected token POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "test-token"})
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func testPrivateKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}))
}
