package domain

import "time"

type ExecutionLog struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type InstanceExecution struct {
	ID              string
	InstanceID      string
	Status          ExecutionStatus
	TriggerType     TriggerType
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationMS      *int
	Logs            []ExecutionLog
	Output          string
	ErrorMessage    string
	DestinationSent bool
	FallbackUsed    bool
	CreatedAt       time.Time
}
