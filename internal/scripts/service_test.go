package scripts

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestScriptsServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Scripts)
	script, err := service.Create(ctx, CreateRequest{
		Name:      "Temperature Collector",
		Author:    "SparkEdge",
		LocalPath: filepath.Join(t.TempDir(), "script"),
		MainFile:  "main.py",
		Tags:      []string{"sensor", "temperature"},
		SchemaConfig: map[string]any{
			"inputs": []any{},
			"outputs": []any{
				map[string]any{"name": "temperature", "type": "number"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if script.Source != domain.ScriptSourceLocal {
		t.Fatalf("expected local source, got %s", script.Source)
	}
	if script.Language != domain.ScriptLanguagePython {
		t.Fatalf("expected python language, got %s", script.Language)
	}

	found, err := service.FindByID(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Tags) != 2 || found.SchemaConfig["outputs"] == nil {
		t.Fatalf("expected tags and schema_config to round-trip, got %#v", found)
	}

	version := "2.0.0"
	updated, err := service.Update(ctx, script.ID, UpdateRequest{Version: &version})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != version {
		t.Fatalf("expected version %q, got %q", version, updated.Version)
	}

	items, err := service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one script, got %d", len(items))
	}

	if err := service.Delete(ctx, script.ID); err != nil {
		t.Fatal(err)
	}
	items, err = service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no scripts after delete, got %d", len(items))
	}
}
