package auth

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestRegisterLoginAndVerifyToken(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Users, store.Projects, "test-secret")

	user, err := service.Register(ctx, RegisterRequest{
		Email:     "USER@example.com",
		Password:  "secret123",
		FirstName: "Ada",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "admin" {
		t.Fatalf("expected admin role, got %s", user.Role)
	}

	personal, err := store.Projects.FindByOwner(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(personal) != 1 || personal[0].Key != "PERSONAL" {
		t.Fatalf("expected personal project, got %#v", personal)
	}

	login, err := service.Login(ctx, LoginRequest{Email: "user@example.com", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if login.Token == "" {
		t.Fatal("expected token")
	}
	if parts := strings.Split(login.Token, "."); len(parts) != 3 {
		t.Fatalf("expected JWT with three parts, got %d parts", len(parts))
	}

	verified, err := service.VerifyToken(ctx, login.Token)
	if err != nil {
		t.Fatal(err)
	}
	if verified.ID != user.ID {
		t.Fatalf("expected verified user %s, got %s", user.ID, verified.ID)
	}
}

func TestVerifyAPIKey(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Users, store.Projects, "test-secret")

	user, err := service.Register(ctx, RegisterRequest{Email: "api@example.com", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}

	verified, err := service.VerifyAPIKey(ctx, user.APIKey)
	if err != nil {
		t.Fatal(err)
	}
	if verified.ID != user.ID {
		t.Fatalf("expected verified user %s, got %s", user.ID, verified.ID)
	}
}
