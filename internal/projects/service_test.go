package projects

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestCreateListUpdateMembersAndDeleteProject(t *testing.T) {
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

	service := NewService(store.Projects, store.ProjectMembers)
	project, err := service.Create(ctx, CreateRequest{
		Name:    "Lab",
		Key:     "LAB",
		OwnerID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if project.Visibility != domain.ProjectVisibilityPrivate {
		t.Fatalf("expected private visibility, got %s", project.Visibility)
	}

	projects, err := service.ListByOwner(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected one project, got %d", len(projects))
	}

	members, err := service.ListMembers(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 1 || members[0].Role != domain.ProjectRoleOwner {
		t.Fatalf("expected owner member, got %#v", members)
	}

	newName := "Lab Updated"
	updated, err := service.Update(ctx, project.ID, UpdateRequest{Name: &newName})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != newName {
		t.Fatalf("expected updated name %q, got %q", newName, updated.Name)
	}

	if err := service.Delete(ctx, project.ID); err != nil {
		t.Fatal(err)
	}
	projects, err = service.ListByOwner(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected no projects after delete, got %d", len(projects))
	}
}
