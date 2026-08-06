package googleprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	DriveServerTypeID  = "googledrive"
	DriveStrategyID    = "googledrive"
	SheetsServerTypeID = "googlespreadsheet"
	SheetsStrategyID   = "googlespreadsheet"
)

func ServerTypes() []domain.ServerType {
	return []domain.ServerType{
		{ID: DriveServerTypeID, Key: "googledrive", Name: "Google Drive", Description: "Envio de arquivos JSON para pastas do Google Drive."},
		{ID: SheetsServerTypeID, Key: "googlespreadsheet", Name: "Google Spreadsheet", Description: "Adicao ou atualizacao de linhas em planilhas do Google Sheets."},
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           DriveStrategyID,
			Name:         "Google Drive",
			Strategy:     "googledrive",
			ServerTypeID: DriveServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "serviceAccountJson", Label: "Service Account JSON", Type: "textarea", Placeholder: "{ ... }"},
			},
		},
		{
			ID:           SheetsStrategyID,
			Name:         "Google Spreadsheet",
			Strategy:     "googlespreadsheet",
			ServerTypeID: SheetsServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "serviceAccountJson", Label: "Service Account JSON", Type: "textarea", Placeholder: "{ ... }"},
			},
		},
	}
}
