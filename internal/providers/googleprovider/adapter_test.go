package googleprovider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestSheetsAdapterAppendsPayloadValues(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/spreadsheets/sheet-1/values/Sheet1!A1:append" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected query %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	adapter := newTestSheetsAdapter(t, "append")
	adapter.baseURL = api.URL
	adapter.tokenURL = tokenServer(t)
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
}

func TestSheetsAdapterUpdatesWhenConfigured(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	adapter := newTestSheetsAdapter(t, "update")
	adapter.baseURL = api.URL
	adapter.tokenURL = tokenServer(t)
	if err := adapter.Send(context.Background(), map[string]any{"temperature": float64(42)}); err != nil {
		t.Fatal(err)
	}
}

func TestDriveAdapterCreatesJSONFile(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/files" || r.URL.Query().Get("uploadType") != "multipart" {
			t.Fatalf("unexpected upload URL %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart content type")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	adapter, err := NewDrive(providers.Config{
		Resource:    map[string]any{"config": map[string]any{"folderId": "folder-1"}},
		Operation:   map[string]any{"config": map[string]any{"fileName": "data.json"}},
		Credentials: map[string]any{"data": map[string]any{"serviceAccountJson": testServiceAccountJSON(t)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	adapter.uploadURL = api.URL
	adapter.tokenURL = tokenServer(t)
	if err := adapter.Send(context.Background(), map[string]any{"ok": true}); err != nil {
		t.Fatal(err)
	}
}

func newTestSheetsAdapter(t *testing.T, action string) *SheetsAdapter {
	t.Helper()
	adapter, err := NewSheets(providers.Config{
		Resource:    map[string]any{"config": map[string]any{"spreadsheetId": "sheet-1"}},
		Operation:   map[string]any{"config": map[string]any{"range": "Sheet1!A1", "action": action}},
		Credentials: map[string]any{"data": map[string]any{"serviceAccountJson": testServiceAccountJSON(t)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func tokenServer(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "test-token"})
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func testServiceAccountJSON(t *testing.T) string {
	t.Helper()
	return `{"client_email":"edge@example.iam.gserviceaccount.com","private_key":` + quoteJSON(testPrivateKey(t)) + `}`
}

func quoteJSON(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
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
