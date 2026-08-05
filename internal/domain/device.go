package domain

import "time"

type DeviceConnectionMethod string

const (
	DeviceConnectionNone   DeviceConnectionMethod = "none"
	DeviceConnectionHTTP   DeviceConnectionMethod = "http"
	DeviceConnectionMQTT   DeviceConnectionMethod = "mqtt"
	DeviceConnectionSerial DeviceConnectionMethod = "serial"
)

type DeviceOtherField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Device struct {
	ID                  string
	DeviceID            string
	Name                string
	Brand               string
	SerialNumber        string
	ConnectionMethod    DeviceConnectionMethod
	IPAddress           string
	Location            string
	Description         string
	Others              []DeviceOtherField
	ResourceOperationID string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
