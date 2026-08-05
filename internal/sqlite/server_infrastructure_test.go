package sqlite

import (
	"context"
	"path/filepath"
	"testing"
)

func TestResourceOperationsResolveTarget(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	credential, err := store.Credentials.Upsert(ctx, UpsertCredentialParams{
		Name:       "HTTP Token",
		AuthTypeID: "auth-bearer",
		Data:       map[string]any{"token": "secret"},
		ProjectID:  "project-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	server, err := store.Servers.Upsert(ctx, UpsertServerParams{
		Name:         "Telemetry API",
		Type:         "http",
		DriverKey:    "http",
		CredentialID: credential.ID,
		Headers:      map[string]any{"X-App": "SparkEdge"},
		ProjectID:    "project-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	resource, err := store.ServerResources.Upsert(ctx, UpsertServerResourceParams{
		ServerID: server.ID,
		Name:     "Measurements",
		Type:     "endpoint",
		Config:   map[string]any{"path": "/measurements"},
	})
	if err != nil {
		t.Fatal(err)
	}
	operation, err := store.ResourceOperations.Upsert(ctx, UpsertResourceOperationParams{
		ResourceID:  resource.ID,
		Name:        "Insert measurement",
		Type:        "http_request",
		Config:      map[string]any{"method": "POST"},
		InputSchema: map[string]any{"temperature": "number"},
	})
	if err != nil {
		t.Fatal(err)
	}

	target, err := store.ResourceOperations.ResolveTarget(ctx, operation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if target.Operation.ID != operation.ID || target.Resource.ID != resource.ID || target.Server.ID != server.ID {
		t.Fatalf("unexpected target %#v", target)
	}
	if target.Credential == nil || target.Credential.Data["token"] != "secret" {
		t.Fatalf("unexpected credential %#v", target.Credential)
	}
}
