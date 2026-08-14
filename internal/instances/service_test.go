package instances

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
)

func TestInstancesServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tagService := tags.NewService(store.Tags, store.InstanceTags)
	service := NewService(store.Instances, tagService, store.Destinations, store.DataMappings)

	active := true
	includeDeviceData := true
	instance, err := service.Create(ctx, Payload{
		Name:              "Temperature Monitor",
		ProjectID:         "project-1",
		DeviceID:          "device-1",
		ScriptID:          "script-1",
		Active:            &active,
		IncludeDeviceData: &includeDeviceData,
		Tags:              []string{"critical", "factory"},
		ScriptInputs:      map[string]any{"ip": "127.0.0.1"},
		ScriptParameters: []any{
			map[string]any{"key": "port", "value": float64(502)},
		},
		FallbackConfig: map[string]any{
			"enabled":                true,
			"strategy":               "active_queue",
			"retry_interval_seconds": float64(120),
		},
		ErrorConfig: map[string]any{"action": "retry"},
		Destinations: []DestinationPayload{
			{
				ResourceOperationID: "resource-op-1",
				Enabled:             &active,
				Priority:            1,
				RetryPolicy: map[string]any{
					"max_retries":    float64(5),
					"retry_interval": float64(30),
				},
				Mapping: &MappingData{
					Mapping:         map[string]any{"temperature": "$.stdout.temperature"},
					PayloadTemplate: map[string]any{"device": "{{device.name}}"},
					CustomFields: []domain.MappingCustomField{
						{Key: "source", Value: "sparkedge"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if instance.Status != domain.InstanceStatusIdle || instance.TriggerType != domain.TriggerInterval {
		t.Fatalf("unexpected defaults: %#v", instance)
	}
	if instance.ScriptParameters["ip"] != "127.0.0.1" || instance.ScriptParameters["port"] != float64(502) {
		t.Fatalf("unexpected script parameters %#v", instance.ScriptParameters)
	}
	if !instance.FallbackEnabled || instance.FallbackStrategy != domain.FallbackActiveQueue || instance.FallbackRetryIntervalSeconds != 120 {
		t.Fatalf("unexpected fallback config %#v", instance)
	}

	linkedTags, err := tagService.FindByInstance(ctx, instance.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(linkedTags) != 2 {
		t.Fatalf("expected two linked tags, got %#v", linkedTags)
	}

	paused := domain.InstanceStatusPaused
	updated, err := service.Update(ctx, instance.ID, Payload{Status: paused})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.InstanceStatusPaused {
		t.Fatalf("expected paused status, got %s", updated.Status)
	}

	activeItems, err := service.ListActive(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(activeItems) != 1 {
		t.Fatalf("expected one active instance, got %d", len(activeItems))
	}

	byProject, err := service.ListByProject(ctx, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(byProject) != 1 {
		t.Fatalf("expected one project instance, got %d", len(byProject))
	}

	withDestinations, err := service.GetWithDestinations(ctx, instance.ID)
	if err != nil {
		t.Fatal(err)
	}
	if withDestinations.Instance.ID != instance.ID {
		t.Fatalf("unexpected with destinations result %#v", withDestinations)
	}
	if len(withDestinations.Destinations) != 1 {
		t.Fatalf("expected one destination, got %#v", withDestinations.Destinations)
	}
	destination := withDestinations.Destinations[0]
	if destination.Destination.ResourceOperationID != "resource-op-1" || destination.Destination.RetryPolicy.MaxRetries != 5 {
		t.Fatalf("unexpected destination %#v", destination.Destination)
	}
	if destination.Mapping == nil || destination.Mapping.Mapping["temperature"] != "$.stdout.temperature" {
		t.Fatalf("unexpected mapping %#v", destination.Mapping)
	}

	if err := service.Delete(ctx, instance.ID); err != nil {
		t.Fatal(err)
	}
	items, err := service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no instances after delete, got %d", len(items))
	}
}

func TestInstancesServiceRejectsInvalidDependenciesAndCycles(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-deps-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tagService := tags.NewService(store.Tags, store.InstanceTags)
	service := NewService(store.Instances, tagService, store.Destinations, store.DataMappings)

	first, err := service.Create(ctx, Payload{
		Name:      "Collector A",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	second, err := service.Create(ctx, Payload{
		Name:      "Collector B",
		ProjectID: "project-1",
		DependsOn: []string{first.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.Create(ctx, Payload{
		Name:      "Collector C",
		ProjectID: "project-1",
		DependsOn: []string{"missing-instance"},
	}); err == nil {
		t.Fatalf("expected missing dependency validation error")
	}

	if _, err := service.Update(ctx, first.ID, Payload{
		DependsOn: []string{second.ID},
	}); err == nil {
		t.Fatalf("expected cycle validation error")
	}

	if _, err := service.Update(ctx, first.ID, Payload{
		DependsOn: []string{first.ID},
	}); err == nil {
		t.Fatalf("expected self dependency validation error")
	}
}
