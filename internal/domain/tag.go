package domain

import "time"

type Tag struct {
	ID        string
	Name      string
	Color     string
	ProjectID string
	CreatedAt time.Time
}

type InstanceTag struct {
	ID         string
	InstanceID string
	TagID      string
	CreatedAt  time.Time
}
