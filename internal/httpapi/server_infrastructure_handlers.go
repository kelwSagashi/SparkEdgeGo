package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/serverinfra"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleServerTypesList(r *http.Request) (any, error) {
	items, err := s.deps.ServerInfra.ListServerTypes(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicServerTypes(items), "error": nil}, nil
}

func (s *Server) handleCredentialsList(r *http.Request) (any, error) {
	ownerID := ""
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified {
		ownerID = identity.UserID
	}
	items, err := s.deps.ServerInfra.ListCredentials(r.Context(), ownerID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicCredentials(items), "error": nil}, nil
}

func (s *Server) handleCredentialGet(r *http.Request) (any, error) {
	item, err := s.deps.ServerInfra.FindCredential(r.Context(), r.PathValue("id"))
	if err != nil {
		return infraError(err)
	}
	return map[string]any{"data": publicCredential(item), "error": nil}, nil
}

func (s *Server) handleCredentialCreate(r *http.Request) (any, error) {
	return s.upsertCredential(r, "")
}
func (s *Server) handleCredentialUpdate(r *http.Request) (any, error) {
	return s.upsertCredential(r, r.PathValue("id"))
}
func (s *Server) upsertCredential(r *http.Request, id string) (any, error) {
	var req serverinfra.CredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.ID = idOr(req.ID, id)
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified {
		req.OwnerID = identity.UserID
	}
	item, err := s.deps.ServerInfra.UpsertCredential(r.Context(), req)
	if err != nil {
		return infraError(err)
	}
	return map[string]any{"data": publicCredential(item), "error": nil}, nil
}

func (s *Server) handleCredentialDelete(r *http.Request) (any, error) {
	if err := s.deps.ServerInfra.DeleteCredential(r.Context(), r.PathValue("id")); err != nil {
		return infraError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleServersList(r *http.Request) (any, error) {
	items, err := s.deps.ServerInfra.ListServers(r.Context(), r.URL.Query().Get("project_id"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicServers(items), "error": nil}, nil
}
func (s *Server) handleServerGet(r *http.Request) (any, error) {
	item, err := s.deps.ServerInfra.FindServer(r.Context(), r.PathValue("id"))
	if err != nil {
		return infraError(err)
	}
	return map[string]any{"data": publicServer(item), "error": nil}, nil
}
func (s *Server) handleServerCreate(r *http.Request) (any, error) { return s.upsertServer(r, "") }
func (s *Server) handleServerUpdate(r *http.Request) (any, error) {
	return s.upsertServer(r, r.PathValue("id"))
}
func (s *Server) upsertServer(r *http.Request, id string) (any, error) {
	var req serverinfra.ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.ID = idOr(req.ID, id)
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified {
		req.CreatedBy = identity.UserID
	}
	item, err := s.deps.ServerInfra.UpsertServer(r.Context(), req)
	if err != nil {
		return infraError(err)
	}
	return map[string]any{"data": publicServer(item), "error": nil}, nil
}
func (s *Server) handleServerDelete(r *http.Request) (any, error) {
	if err := s.deps.ServerInfra.DeleteServer(r.Context(), r.PathValue("id")); err != nil {
		return infraError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}
func (s *Server) handleServerResourcesList(r *http.Request) (any, error) {
	items, err := s.deps.ServerInfra.ListResources(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": items, "error": nil}, nil
}
func (s *Server) handleResourceOperationsList(r *http.Request) (any, error) {
	items, err := s.deps.ServerInfra.ListOperations(r.Context(), r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": items, "error": nil}, nil
}

func idOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
func infraError(err error) (any, error) {
	if errors.Is(err, serverinfra.ErrInvalidInput) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid server infrastructure input")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Resource not found"}, nil
	}
	return nil, err
}
func publicServerTypes(items []domain.ServerType) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{"id": item.ID, "key": item.Key, "name": item.Name, "description": item.Description})
	}
	return result
}
func publicCredentials(items []domain.Credential) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicCredential(item))
	}
	return result
}
func publicCredential(item domain.Credential) map[string]any {
	return map[string]any{"id": item.ID, "name": item.Name, "auth_type_id": item.AuthTypeID, "data": item.Data, "owner_id": item.OwnerID, "project_id": item.ProjectID, "created_at": item.CreatedAt}
}
func publicServers(items []domain.Server) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicServer(item))
	}
	return result
}
func publicServer(item domain.Server) map[string]any {
	return map[string]any{"id": item.ID, "name": item.Name, "type": item.Type, "server_type_id": item.ServerTypeID, "driver_key": item.DriverKey, "credential_id": item.CredentialID, "headers": item.Headers, "project_id": item.ProjectID, "created_by": item.CreatedBy, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
}
