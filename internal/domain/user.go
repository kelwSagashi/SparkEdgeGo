package domain

import "time"

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleEditor UserRole = "editor"
	RoleViewer UserRole = "viewer"
)

type User struct {
	ID           string
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	Role         UserRole
	IsActive     bool
	APIKey       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
