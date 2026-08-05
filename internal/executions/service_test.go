package executions

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestExecutionsServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Executions)
	startedAt := time.Now().UTC()
	execution, err := service.Create(ctx, sqlite.CreateInstanceExecutionParams{
		InstanceID:  "instance-1",
		Status:      domain.ExecutionRunning,
		TriggerType: domain.TriggerManual,
		StartedAt:   &startedAt,
		Logs: []domain.ExecutionLog{
			{Level: "info", Message: "started", Timestamp: startedAt},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if execution.ID == "" || execution.Status != domain.ExecutionRunning {
		t.Fatalf("unexpected execution %#v", execution)
	}

	list, err := service.ListByInstance(ctx, "instance-1", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || len(list[0].Logs) != 1 {
		t.Fatalf("unexpected execution list %#v", list)
	}

	finishedAt := time.Now().UTC()
	duration := 123
	output := `{"ok":true}`
	destinationSent := true
	updated, err := service.UpdateStatus(ctx, execution.ID, sqlite.UpdateInstanceExecutionStatusParams{
		Status:          domain.ExecutionSuccess,
		FinishedAt:      &finishedAt,
		DurationMS:      &duration,
		Output:          &output,
		DestinationSent: &destinationSent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.ExecutionSuccess || updated.DurationMS == nil || *updated.DurationMS != duration || !updated.DestinationSent {
		t.Fatalf("unexpected updated execution %#v", updated)
	}
}
