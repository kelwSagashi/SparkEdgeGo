package serverinfra

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/httpprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestSeedCatalogPersistsHTTPMetadata(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store)
	if err := service.SeedCatalog(ctx, []domain.ServerType{httpprovider.ServerType()}, httpprovider.AuthTypes()); err != nil {
		t.Fatal(err)
	}

	serverTypes, err := service.ListServerTypes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(serverTypes) != 1 || serverTypes[0].ID != httpprovider.ServerTypeID {
		t.Fatalf("unexpected server types %#v", serverTypes)
	}

	authTypes, err := service.ListAuthTypes(ctx, httpprovider.ServerTypeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(authTypes) != 4 {
		t.Fatalf("expected 4 http auth types, got %#v", authTypes)
	}
	if authTypes[0].Fields == nil {
		t.Fatalf("expected fields to be initialized")
	}
}
