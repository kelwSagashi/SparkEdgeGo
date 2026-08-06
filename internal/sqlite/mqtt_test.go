package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

func TestMqttCommandsRepositoryPersistsStatusLifecycle(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	command, err := store.MqttCommands.Save(ctx, "cmd-1", "ping", map[string]any{"ok": true})
	if err != nil {
		t.Fatal(err)
	}
	if command.Status != domain.MqttCommandPending {
		t.Fatalf("unexpected command status %#v", command)
	}
	running, err := store.MqttCommands.UpdateStatus(ctx, "cmd-1", domain.MqttCommandRunning, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if running.StartedAt == nil {
		t.Fatal("expected started_at to be set")
	}
	done, err := store.MqttCommands.UpdateStatus(ctx, "cmd-1", domain.MqttCommandDone, map[string]any{"pong": true}, "")
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != domain.MqttCommandDone || done.Result["pong"] != true || done.FinishedAt == nil {
		t.Fatalf("unexpected done command %#v", done)
	}
}

func TestMqttQueueRepositoryListsIncrementsAndDeletes(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item, err := store.MqttQueue.Enqueue(ctx, "spark/edge-1/logs", `{"ok":true}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MqttQueue.IncrementAttempt(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	items, err := store.MqttQueue.ListPending(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Attempts != 1 || items[0].LastAttemptAt == nil {
		t.Fatalf("unexpected queue items %#v", items)
	}
	if err := store.MqttQueue.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	items, err = store.MqttQueue.ListPending(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty queue, got %#v", items)
	}
}
