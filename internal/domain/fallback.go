package domain

import "time"

type FallbackItemStatus string

const (
	FallbackPending FallbackItemStatus = "pending"
	FallbackSending FallbackItemStatus = "sending"
	FallbackSent    FallbackItemStatus = "sent"
	FallbackFailed  FallbackItemStatus = "failed"
)

type LocalFallbackItem struct {
	ID            string
	InstanceID    string
	DestinationID string
	ExecutionID   string
	Status        FallbackItemStatus
	Payload       string
	Filepath      string
	RetryCount    int
	LastRetryAt   *time.Time
	NextRetryAt   *time.Time
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
