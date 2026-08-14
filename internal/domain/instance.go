package domain

import "time"

type TriggerType string

const (
	TriggerManual             TriggerType = "manual"
	TriggerWebhook            TriggerType = "webhook"
	TriggerInterval           TriggerType = "interval"
	TriggerIntervalAndWebhook TriggerType = "interval_and_webhook"
	TriggerEvent              TriggerType = "event"
	TriggerMQTT               TriggerType = "mqtt"
	TriggerStateChange        TriggerType = "state_change"
	TriggerWorkflow           TriggerType = "workflow"
)

type InstanceStatus string

const (
	InstanceStatusIdle    InstanceStatus = "idle"
	InstanceStatusRunning InstanceStatus = "running"
	InstanceStatusPaused  InstanceStatus = "paused"
	InstanceStatusError   InstanceStatus = "error"
)

type FallbackStrategy string

const (
	FallbackBackgroundJob FallbackStrategy = "background_job"
	FallbackActiveQueue   FallbackStrategy = "active_queue"
)

type OnErrorAction string

const (
	OnErrorLogOnly       OnErrorAction = "log_only"
	OnErrorRetry         OnErrorAction = "retry"
	OnErrorNotifyWebhook OnErrorAction = "notify_webhook"
	OnErrorStop          OnErrorAction = "stop"
)

type ExecutionStatus string

const (
	ExecutionRunning ExecutionStatus = "running"
	ExecutionSuccess ExecutionStatus = "success"
	ExecutionFailed  ExecutionStatus = "failed"
	ExecutionTimeout ExecutionStatus = "timeout"
	ExecutionQueued  ExecutionStatus = "queued"
)

type ExecutionMode string

const (
	ExecutionModeSequential ExecutionMode = "sequential"
	ExecutionModeParallel   ExecutionMode = "parallel"
)

type Instance struct {
	ID                           string
	Name                         string
	Description                  string
	Tags                         []string
	Status                       InstanceStatus
	Active                       bool
	ProjectID                    string
	DeviceID                     string
	ScriptID                     string
	IncludeDeviceData            bool
	ScriptParameters             map[string]any
	TriggerType                  TriggerType
	TriggerConfig                map[string]any
	DependsOn                    []string
	ExecutionMode                ExecutionMode
	OrchestrationConfig          map[string]any
	FallbackEnabled              bool
	FallbackStrategy             FallbackStrategy
	FallbackRetryIntervalSeconds int
	OnErrorAction                OnErrorAction
	OnErrorConfig                map[string]any
	CreatedBy                    string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type Script struct {
	ID        string
	Name      string
	LocalPath string
	MainFile  string
	VenvPath  string
	Schema    map[string]any
	Ready     bool
}

type ExecutionContext struct {
	ExecutionID      string
	InstanceID       string
	Device           map[string]any
	Instance         Instance
	Script           Script
	ScriptParameters map[string]any
	TriggerType      TriggerType
	Timestamp        time.Time
}
