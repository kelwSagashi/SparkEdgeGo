package tags

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestTagsServiceLifecycleAndInstanceSync(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Tags, store.InstanceTags)
	tag, err := service.Create(ctx, CreateRequest{
		Name:      "critical",
		Color:     "#ff0000",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if tag.Color != "#ff0000" {
		t.Fatalf("expected tag color, got %q", tag.Color)
	}

	updated, err := service.Create(ctx, CreateRequest{
		Name:      "critical",
		Color:     "#00ff00",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != tag.ID || updated.Color != "#00ff00" {
		t.Fatalf("expected upserted tag, got %#v", updated)
	}

	found, err := service.Search(ctx, "crit", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].ID != tag.ID {
		t.Fatalf("expected search result, got %#v", found)
	}

	tagIDs, err := service.FindOrCreateByNames(ctx, []string{"critical", "new tag", " "}, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tagIDs) != 2 {
		t.Fatalf("expected two tag ids, got %#v", tagIDs)
	}

	linked, err := service.SyncTags(ctx, "instance-1", tagIDs)
	if err != nil {
		t.Fatal(err)
	}
	if len(linked) != 2 {
		t.Fatalf("expected linked tags, got %#v", linked)
	}

	deleted, err := service.Delete(ctx, tag.ID)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.ID != tag.ID {
		t.Fatalf("expected deleted tag %s, got %s", tag.ID, deleted.ID)
	}
}
