package supabaseprovider

import "github.com/kelwSagashi/sparkedge-go/internal/domain"

const (
	ServerTypeID = "supabase"
	StrategyID   = "supabase"
)

func ServerType() domain.ServerType {
	return domain.ServerType{
		ID:          ServerTypeID,
		Key:         "supabase",
		Name:        "Supabase",
		Description: "Integracao com Supabase via API PostgREST.",
	}
}

func AuthTypes() []domain.AuthType {
	return []domain.AuthType{
		{
			ID:           StrategyID,
			Name:         "Supabase",
			Strategy:     "supabase",
			ServerTypeID: ServerTypeID,
			Fields: []domain.AuthTypeField{
				{Key: "url", Label: "Supabase URL", Type: "text", Placeholder: "https://xyz.supabase.co"},
				{Key: "apiKey", Label: "API Key (Service Role ou Anon)", Type: "password"},
			},
		},
	}
}
