package mongoprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	ServerTypeID = "mongodb"
	StrategyID   = "mongo"
)

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "mongodb",
		Name:        "MongoDB",
		Description: "Insercao e consulta de documentos em colecoes do MongoDB.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyID,
			Name:         "MongoDB",
			Strategy:     "mongo",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "uri", Label: "Connection URI", Type: "text", Placeholder: "mongodb://user:pass@localhost:27017"},
				{Key: "database", Label: "Database Name", Type: "text", Placeholder: "my_database"},
			},
		},
	}
}
