package jsonfileprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	ServerTypeID = "jsonfile"
	StrategyID   = "jsonfile"
)

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "jsonfile",
		Name:        "Arquivo Local",
		Description: "Grava payloads em arquivos JSON locais para auditoria, buffer e exportacao offline.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyID,
			Name:         "Arquivo Local",
			Strategy:     StrategyID,
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "basePath", Label: "Diretorio base", Type: "text", Placeholder: "./data/exports"},
			},
		},
	}
}
