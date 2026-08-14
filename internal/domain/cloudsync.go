package domain

import "time"

type CloudSyncStatus string

const (
	CloudSyncPending CloudSyncStatus = "pending"
	CloudSyncSent    CloudSyncStatus = "sent"
	CloudSyncFailed  CloudSyncStatus = "failed"
)

type CloudSyncItem struct {
	ID            string
	EventType     string
	Priority      int
	Payload       map[string]any
	Status        CloudSyncStatus
	Attempts      int
	LastAttemptAt *time.Time
	NextAttemptAt *time.Time
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
