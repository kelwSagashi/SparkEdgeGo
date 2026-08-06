package domain

import "time"

type MqttCommandStatus string

const (
	MqttCommandPending MqttCommandStatus = "pending"
	MqttCommandRunning MqttCommandStatus = "running"
	MqttCommandDone    MqttCommandStatus = "done"
	MqttCommandError   MqttCommandStatus = "error"
	MqttCommandIgnored MqttCommandStatus = "ignored"
)

type MqttCommand struct {
	ID         string
	CommandID  string
	Type       string
	Payload    map[string]any
	Status     MqttCommandStatus
	Result     map[string]any
	Error      string
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}

type MqttQueueItem struct {
	ID            string
	Topic         string
	Payload       string
	Attempts      int
	LastAttemptAt *time.Time
	CreatedAt     time.Time
}
