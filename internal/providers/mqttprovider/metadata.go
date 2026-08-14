package mqttprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	ServerTypeID = "mqtt"
	StrategyID   = "mqtt"
)

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "mqtt",
		Name:        "MQTT Broker",
		Description: "Publicacao de mensagens JSON em topicos MQTT genericos.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyID,
			Name:         "MQTT",
			Strategy:     "mqtt",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "brokerUrl", Label: "Broker URL", Type: "text", Placeholder: "tcp://localhost:1883"},
				{Key: "clientId", Label: "Client ID", Type: "text", Placeholder: "sparkedge-client"},
				{Key: "username", Label: "Username", Type: "text"},
				{Key: "password", Label: "Password", Type: "password"},
			},
		},
	}
}
