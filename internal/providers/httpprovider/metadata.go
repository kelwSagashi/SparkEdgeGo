package httpprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const ServerTypeID = "http"

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "http",
		Name:        "Servidor HTTP REST",
		Description: "Servicos HTTP/REST que podem ser consultados via requisicao JSON.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyNoAuth,
			Name:         "No Auth",
			Strategy:     "http_noauth",
			ServerTypeID: ServerTypeID,
			Fields:       []domain.AuthTypeField{},
		},
		{
			ID:           StrategyAPIKey,
			Name:         "API Key",
			Strategy:     "http_apikey",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "key", Label: "Key Name", Type: "text", Placeholder: "X-API-Key"},
				{Key: "value", Label: "Key Value", Type: "password"},
				{Key: "in", Label: "In", Type: "select", Options: []domain.AuthOption{
					{Label: "Header", Value: "header"},
					{Label: "Query Param", Value: "query"},
				}},
			},
		},
		{
			ID:           StrategyBasic,
			Name:         "Basic Auth",
			Strategy:     "http_basicauth",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "username", Label: "Username", Type: "text"},
				{Key: "password", Label: "Password", Type: "password"},
			},
		},
		{
			ID:           StrategyBearer,
			Name:         "HTTP Bearer Token",
			Strategy:     "http_bearer",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "token", Label: "Token (Bearer)", Type: "password"},
			},
		},
	}
}
