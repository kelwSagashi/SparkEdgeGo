package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

func (s *Server) handleAdaptersMetadata(r *http.Request) (any, error) {
	serverTypes, err := s.deps.ServerInfra.ListServerTypes(r.Context())
	if err != nil {
		return nil, err
	}
	authTypes, err := s.deps.ServerInfra.ListAuthTypes(r.Context(), "")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"data": map[string]any{
			"server_types": publicServerTypes(serverTypes),
			"auth_types":   publicAuthTypes(authTypes),
		},
		"error": nil,
	}, nil
}

func (s *Server) handleAdapterDiscover(r *http.Request) (any, error) {
	var req struct {
		Credentials map[string]any `json:"credentials"`
		Data        map[string]any `json:"data"`
		Server      map[string]any `json:"server"`
		Resource    map[string]any `json:"resource"`
		Operation   map[string]any `json:"operation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	credentials := req.Credentials
	if credentials == nil {
		credentials = req.Data
	}
	adapterID := r.PathValue("id")
	adapter, ok, err := s.deps.Providers.Create(adapterID, providers.Config{
		Server:      defaultMap(req.Server, map[string]any{"type": adapterID, "driver_key": adapterID}),
		Resource:    defaultMap(req.Resource, map[string]any{}),
		Operation:   defaultMap(req.Operation, map[string]any{"type": "discover", "config": map[string]any{}}),
		Credentials: map[string]any{"auth_type_id": adapterID, "data": defaultMap(credentials, map[string]any{})},
	})
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	if !ok {
		return map[string]any{"success": false, "error": "No adapter registered for adapter ID: " + adapterID}, nil
	}
	resources, err := adapter.Discover(r.Context())
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	return map[string]any{"success": true, "resources": resources}, nil
}

func defaultMap(value map[string]any, fallback map[string]any) map[string]any {
	if value == nil {
		return fallback
	}
	return value
}
