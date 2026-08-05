package users

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestUsersServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Users)
	user, err := service.Create(ctx, CreateRequest{
		Email:     "USER@example.com",
		FirstName: "Grace",
		Role:      domain.RoleEditor,
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}

	items, err := service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one user, got %d", len(items))
	}

	lastName := "Hopper"
	updated, err := service.Update(ctx, user.ID, UpdateRequest{LastName: &lastName})
	if err != nil {
		t.Fatal(err)
	}
	if updated.LastName != lastName {
		t.Fatalf("expected last name %q, got %q", lastName, updated.LastName)
	}

	apiKey, err := service.CreateAPIKey(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if apiKey == "" || apiKey == user.APIKey {
		t.Fatal("expected a new api key")
	}

	if err := service.Delete(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	items, err = service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no users after delete, got %d", len(items))
	}
}

func TestFindProjectUserByName(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	user, err := store.Users.Create(ctx, sqlite.CreateUserParams{
		Email:    "owner@example.com",
		Role:     domain.RoleAdmin,
		IsActive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Projects.Upsert(ctx, sqlite.CreateProjectParams{
		Name:    "PERSONAL",
		Key:     "PERSONAL",
		OwnerID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	service := NewService(store.Users)
	result, err := service.FindProjectUserByName(ctx, user.ID, "PERSONAL")
	if err != nil {
		t.Fatal(err)
	}
	if result.User.ID != user.ID || result.Project.Name != "PERSONAL" {
		t.Fatalf("unexpected project user result: %#v", result)
	}
}
