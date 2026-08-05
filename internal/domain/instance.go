package domain

import "time"

type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerWebhook  TriggerType = "webhook"
	TriggerInterval TriggerType = "interval"
)

type ExecutionStatus string

const (
	ExecutionRunning ExecutionStatus = "running"
	ExecutionSuccess ExecutionStatus = "success"
	ExecutionFailed  ExecutionStatus = "failed"
	ExecutionTimeout ExecutionStatus = "timeout"
	ExecutionQueued  ExecutionStatus = "queued"
)

type Instance struct {
	ID                string
	Name              string
	ScriptID          string
	DeviceID          string
	IncludeDeviceData bool
	FallbackEnabled   bool
	FallbackStrategy  string
	ScriptParameters  map[string]any
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
