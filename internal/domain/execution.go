package domain

import "time"

type ExecutionLog struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ExecutionDestinationDetail struct {
	DestinationID       string         `json:"destination_id"`
	ResourceOperationID string         `json:"resource_operation_id"`
	ServerName          string         `json:"server_name,omitempty"`
	ResourceName        string         `json:"resource_name,omitempty"`
	OperationName       string         `json:"operation_name,omitempty"`
	Status              string         `json:"status"`
	Payload             map[string]any `json:"payload"`
	Error               string         `json:"error,omitempty"`
	UsedFallback        bool           `json:"used_fallback"`
	Timestamp           time.Time      `json:"timestamp"`
}

type InstanceExecution struct {
	ID                 string
	InstanceID         string
	Status             ExecutionStatus
	TriggerType        TriggerType
	StartedAt          *time.Time
	FinishedAt         *time.Time
	DurationMS         *int
	Logs               []ExecutionLog
	Output             string
	ErrorMessage       string
	DestinationSent    bool
	FallbackUsed       bool
	InputPayload       map[string]any
	OutputPayload      map[string]any
	DestinationDetails []ExecutionDestinationDetail
	CreatedAt          time.Time
}
