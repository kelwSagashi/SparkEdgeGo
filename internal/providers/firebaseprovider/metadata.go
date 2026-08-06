package firebaseprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	ServerTypeID = "firebase"
	StrategyID   = "firebase"
)

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "firebase",
		Name:        "Firebase Firestore",
		Description: "Insercao de documentos em colecoes do Firestore.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyID,
			Name:         "Firebase (Firestore)",
			Strategy:     "firebase",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "projectId", Label: "Project ID", Type: "text"},
				{Key: "clientEmail", Label: "Client Email", Type: "text"},
				{Key: "privateKey", Label: "Private Key", Type: "textarea"},
			},
		},
	}
}
