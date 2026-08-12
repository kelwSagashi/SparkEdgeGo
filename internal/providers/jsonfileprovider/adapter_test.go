package jsonfileprovider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func TestSendWritesNDJSONLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter, err := New(providers.Config{
		Credentials: map[string]any{"data": map[string]any{"basePath": dir}},
		Resource:    map[string]any{"config": map[string]any{"fileName": "telemetry/output.ndjson"}},
		Operation:   map[string]any{"config": map[string]any{"format": "ndjson"}},
	})
	if err != nil {
		t.Fatalf("expected adapter, got error: %v", err)
	}

	if err := adapter.Send(context.Background(), map[string]any{"status": "ok"}); err != nil {
		t.Fatalf("expected write to succeed, got error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "telemetry", "output.ndjson"))
	if err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}
	if string(data) != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected file content: %q", string(data))
	}
}

func TestSendWritesJSONArray(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter, err := New(providers.Config{
		Credentials: map[string]any{"data": map[string]any{"basePath": dir}},
		Resource:    map[string]any{"config": map[string]any{"fileName": "telemetry/output.json"}},
		Operation:   map[string]any{"config": map[string]any{"format": "json_array", "mode": "append"}},
	})
	if err != nil {
		t.Fatalf("expected adapter, got error: %v", err)
	}

	if err := adapter.Send(context.Background(), map[string]any{"status": "ok"}); err != nil {
		t.Fatalf("expected first write to succeed, got error: %v", err)
	}
	if err := adapter.Send(context.Background(), map[string]any{"status": "warning"}); err != nil {
		t.Fatalf("expected second write to succeed, got error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "telemetry", "output.json"))
	if err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}

	var payload []map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("expected valid json array, got error: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload))
	}
}
