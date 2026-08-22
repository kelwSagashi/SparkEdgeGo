package domain

import "time"

type CloudSyncStatus string

const (
	CloudSyncPending CloudSyncStatus = "pending"
	CloudSyncSent    CloudSyncStatus = "sent"
	CloudSyncFailed  CloudSyncStatus = "failed"
)

type CloudSyncItem struct {
	ID            string          `json:"id"`
	EventType     string          `json:"event_type"`
	Priority      int             `json:"priority"`
	Payload       map[string]any  `json:"payload"`
	Status        CloudSyncStatus `json:"status"`
	Attempts      int             `json:"attempts"`
	LastAttemptAt *time.Time      `json:"last_attempt_at,omitempty"`
	NextAttemptAt *time.Time      `json:"next_retry_at,omitempty"`
	LastError     string          `json:"last_error,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
