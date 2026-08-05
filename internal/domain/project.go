package domain

import "time"

type ProjectVisibility string

const (
	ProjectVisibilityPrivate ProjectVisibility = "private"
	ProjectVisibilityPublic  ProjectVisibility = "public"
)

type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "owner"
	ProjectRoleEditor ProjectRole = "editor"
	ProjectRoleViewer ProjectRole = "viewer"
)

type Project struct {
	ID          string
	Name        string
	Key         string
	Description string
	Visibility  ProjectVisibility
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectMember struct {
	ID        string
	ProjectID string
	UserID    string
	Role      ProjectRole
	CreatedAt time.Time
}

type ProjectUser struct {
	User    User
	Project Project
}
