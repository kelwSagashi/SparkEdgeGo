package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/executions"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	apperuntime "github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
)

func TestDependentInstanceReadyRequiresSuccessfulDependencies(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-workflow.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tagService := tags.NewService(store.Tags, store.InstanceTags)
	instanceService := instances.NewService(store.Instances, tagService, store.Destinations, store.DataMappings)
	executionService := executions.NewService(store.Executions)
	application := &App{
		Instances:        instanceService,
		Executions:       executionService,
		workflowInflight: map[string]bool{},
	}

	first := mustCreateInstance(t, ctx, instanceService, "Source A", nil, nil)
	second := mustCreateInstance(t, ctx, instanceService, "Source B", nil, nil)
	target := mustCreateInstance(t, ctx, instanceService, "Target", []string{first.ID, second.ID}, nil)

	ready, reason, err := application.dependentInstanceReady(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if ready || reason == "" {
		t.Fatalf("expected target to wait for dependencies, got ready=%v reason=%q", ready, reason)
	}

	mustCreateExecution(t, ctx, executionService, first.ID, domain.ExecutionSuccess, time.Now().UTC().Add(-3*time.Minute))
	ready, _, err = application.dependentInstanceReady(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if ready {
		t.Fatalf("expected target to remain blocked until every dependency succeeds")
	}

	mustCreateExecution(t, ctx, executionService, second.ID, domain.ExecutionFailed, time.Now().UTC().Add(-2*time.Minute))
	ready, _, err = application.dependentInstanceReady(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if ready {
		t.Fatalf("expected failed dependency to keep target blocked")
	}

	mustCreateExecution(t, ctx, executionService, second.ID, domain.ExecutionSuccess, time.Now().UTC().Add(-time.Minute))
	ready, reason, err = application.dependentInstanceReady(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if !ready || reason != "" {
		t.Fatalf("expected target to be ready after successful dependencies, got ready=%v reason=%q", ready, reason)
	}
}

func TestDependentInstanceReadyRespectsDebounce(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-workflow-debounce.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tagService := tags.NewService(store.Tags, store.InstanceTags)
	instanceService := instances.NewService(store.Instances, tagService, store.Destinations, store.DataMappings)
	executionService := executions.NewService(store.Executions)
	application := &App{
		Instances:        instanceService,
		Executions:       executionService,
		workflowInflight: map[string]bool{},
	}

	source := mustCreateInstance(t, ctx, instanceService, "Source", nil, nil)
	target := mustCreateInstance(t, ctx, instanceService, "Target", []string{source.ID}, map[string]any{"debounce_seconds": float64(300)})

	mustCreateExecution(t, ctx, executionService, source.ID, domain.ExecutionSuccess, time.Now().UTC().Add(-2*time.Minute))
	mustCreateExecution(t, ctx, executionService, target.ID, domain.ExecutionSuccess, time.Now().UTC())

	ready, reason, err := application.dependentInstanceReady(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if ready || reason == "" {
		t.Fatalf("expected debounce to block target, got ready=%v reason=%q", ready, reason)
	}
}

func TestBuildDependentTriggerInputIncludesWorkflowContext(t *testing.T) {
	parentExecution := domain.InstanceExecution{
		ID:          "exec-1",
		InstanceID:  "instance-1",
		TriggerType: domain.TriggerManual,
		InputPayload: map[string]any{
			"asset_id": "pump-a",
		},
		OutputPayload: map[string]any{
			"temperature": float64(42),
		},
	}
	parentResult := apperuntime.TriggerResult{
		Status: domain.ExecutionSuccess,
		Output: map[string]any{
			"temperature": float64(42),
		},
	}
	target := domain.Instance{
		ID:   "instance-2",
		Name: "Downstream",
	}

	input := buildDependentTriggerInput(parentExecution, parentResult, target)
	workflow, ok := input["workflow"].(map[string]any)
	if !ok {
		t.Fatalf("expected workflow context, got %#v", input)
	}
	upstream, ok := input["upstream"].(map[string]any)
	if !ok {
		t.Fatalf("expected upstream context, got %#v", input)
	}

	if workflow["source_instance_id"] != "instance-1" || workflow["target_instance_id"] != "instance-2" {
		t.Fatalf("unexpected workflow context %#v", workflow)
	}
	if upstream["execution_id"] != "exec-1" {
		t.Fatalf("unexpected upstream execution context %#v", upstream)
	}
	output, ok := upstream["output"].(map[string]any)
	if !ok || output["temperature"] != float64(42) {
		t.Fatalf("expected upstream output payload, got %#v", upstream)
	}
}

func mustCreateInstance(t *testing.T, ctx context.Context, service *instances.Service, name string, dependsOn []string, orchestrationConfig map[string]any) domain.Instance {
	t.Helper()

	active := true
	instance, err := service.Create(ctx, instances.Payload{
		Name:                name,
		ProjectID:           "project-1",
		Active:              &active,
		DependsOn:           dependsOn,
		OrchestrationConfig: orchestrationConfig,
	})
	if err != nil {
		t.Fatal(err)
	}
	return instance
}

func mustCreateExecution(t *testing.T, ctx context.Context, service *executions.Service, instanceID string, status domain.ExecutionStatus, startedAt time.Time) domain.InstanceExecution {
	t.Helper()

	execution, err := service.Create(ctx, sqlite.CreateInstanceExecutionParams{
		InstanceID:  instanceID,
		Status:      status,
		TriggerType: domain.TriggerWorkflow,
		StartedAt:   &startedAt,
		FinishedAt:  &startedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	return execution
}
