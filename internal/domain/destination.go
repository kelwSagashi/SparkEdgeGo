package domain

import "time"

type RetryPolicy struct {
	MaxRetries                    int    `json:"max_retries"`
	RetryInterval                 int    `json:"retry_interval"`
	TimeoutSeconds                int    `json:"timeout_seconds"`
	ContinueOnError               bool   `json:"continue_on_error"`
	IsolationMode                 string `json:"isolation_mode"`
	CircuitBreakerThreshold       int    `json:"circuit_breaker_threshold"`
	CircuitBreakerCooldownSeconds int    `json:"circuit_breaker_cooldown_seconds"`
}

type MappingCustomField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type InstanceDestination struct {
	ID                  string
	InstanceID          string
	ResourceOperationID string
	Enabled             bool
	Priority            int
	RetryPolicy         RetryPolicy
	CreatedAt           time.Time
}

type CircuitBreakerState struct {
	DestinationID       string
	ConsecutiveFailures int
	OpenedUntil         *time.Time
	UpdatedAt           time.Time
}

type DataMapping struct {
	ID                    string
	InstanceDestinationID string
	Mapping               map[string]any
	PayloadTemplate       map[string]any
	CustomFields          []MappingCustomField
	TransformScript       string
	CreatedAt             time.Time
}

type InstanceDestinationWithMapping struct {
	Destination  InstanceDestination
	Mapping      *DataMapping
	BreakerState *CircuitBreakerState
}
