package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

func (s *Server) handleUsersList(r *http.Request) (any, error) {
	items, err := s.deps.Users.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicUsers(items), "error": nil}, nil
}

func (s *Server) handleUserGet(r *http.Request) (any, error) {
	user, err := s.deps.Users.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return userError(err)
	}
	return map[string]any{"data": publicUser(user), "error": nil}, nil
}

func (s *Server) handleUserProjectGet(r *http.Request) (any, error) {
	result, err := s.deps.Users.FindProjectUserByName(r.Context(), r.PathValue("id"), r.PathValue("project"))
	if err != nil {
		return userError(err)
	}
	return map[string]any{
		"data": map[string]any{
			"user":    publicUser(result.User),
			"project": publicProject(result.Project),
		},
		"error": nil,
	}, nil
}

func (s *Server) handleUserCreate(r *http.Request) (any, error) {
	var req users.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := s.deps.Users.Create(r.Context(), req)
	if err != nil {
		return userError(err)
	}
	return map[string]any{"data": publicUser(user), "error": nil}, nil
}

func (s *Server) handleUserUpdate(r *http.Request) (any, error) {
	var req users.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := s.deps.Users.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return userError(err)
	}
	return map[string]any{"data": publicUser(user), "error": nil}, nil
}

func (s *Server) handleUserDelete(r *http.Request) (any, error) {
	if err := s.deps.Users.Delete(r.Context(), r.PathValue("id")); err != nil {
		return userError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleUserCreateAPIKey(r *http.Request) (any, error) {
	apiKey, err := s.deps.Users.CreateAPIKey(r.Context(), r.PathValue("id"))
	if err != nil {
		return userError(err)
	}
	return map[string]any{
		"data":  map[string]any{"userId": r.PathValue("id"), "apiKey": apiKey},
		"error": nil,
	}, nil
}

func userError(err error) (any, error) {
	if errors.Is(err, users.ErrInvalidUser) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid user")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "User not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "User already exists"}, nil
	}
	return nil, err
}

func publicUsers(items []domain.User) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicUser(item))
	}
	return result
}
