package domain

import "time"

type ServerType struct {
	ID          string
	Key         string
	Name        string
	Description string
}

type AuthType struct {
	ID           string
	Name         string
	Strategy     string
	Fields       []AuthTypeField
	ServerTypeID string
}

type AuthTypeField struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Type        string         `json:"type"`
	Placeholder string         `json:"placeholder"`
	Grid        string         `json:"grid"`
	Options     []AuthOption   `json:"options"`
	Extra       map[string]any `json:"extra,omitempty"`
}

type AuthOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type Credential struct {
	ID         string
	Name       string
	AuthTypeID string
	Data       map[string]any
	OwnerID    string
	ProjectID  string
	CreatedAt  time.Time
}

type Server struct {
	ID           string
	Name         string
	Type         string
	ServerTypeID string
	DriverKey    string
	CredentialID string
	Headers      map[string]any
	ProjectID    string
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ServerResource struct {
	ID        string
	ServerID  string
	Name      string
	Type      string
	Config    map[string]any
	CreatedAt time.Time
}

type ResourceOperation struct {
	ID           string
	ResourceID   string
	Name         string
	Type         string
	Config       map[string]any
	InputSchema  map[string]any
	OutputSchema map[string]any
	CreatedAt    time.Time
}
