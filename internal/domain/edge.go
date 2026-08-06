package domain

import "time"

type EdgeIdentity struct {
	ID          string
	EdgeID      string
	EdgeName    string
	Provisioned bool
	CreatedAt   time.Time
}

type EdgeCredentials struct {
	ID        string
	Type      string
	BrokerURL string
	Username  string
	Password  string
	UpdatedAt time.Time
}

type EdgeConfig struct {
	ID             string
	EdgeName       string
	Lat            string
	Lng            string
	LocationSource string
	Tags           []string
	OS             string
	OSVersion      string
	EdgeVersion    string
	Hardware       string
	Environment    string
	Description    string
	UpdatedAt      time.Time
}

type ProvisionedEdge struct {
	EdgeID      string         `json:"edge_id"`
	EdgeName    string         `json:"edge_name"`
	MQTT        EdgeMQTTConfig `json:"mqtt"`
	Provisioned bool           `json:"provisioned"`
}

type EdgeMQTTConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
