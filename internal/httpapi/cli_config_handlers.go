package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/config"
)

func (s *Server) handleCliConfigGet(_ *http.Request) (any, error) {
	effective, _, err := s.deps.Config.Load()
	if err != nil {
		return nil, err
	}
	return effective, nil
}

func (s *Server) handleCliConfigUpdate(r *http.Request) (any, error) {
	var req config.Update
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Cloud == nil && req.DB == nil && req.Auth == nil && req.Server == nil {
		return nil, NewHTTPError(http.StatusBadRequest, "Nenhum campo valido para atualizar.")
	}

	effective, err := s.deps.Config.Save(req)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return map[string]any{
		"success": true,
		"message": "Configuracao salva em config.yml. Reinicie o servico para aplicar todas as mudancas.",
		"config":  effective,
	}, nil
}
