package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
)

func (s *Server) handleTagsList(r *http.Request) (any, error) {
	items, err := s.deps.Tags.ListAll(r.Context(), r.URL.Query().Get("project_id"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicTags(items), "error": nil}, nil
}

func (s *Server) handleTagsSearch(r *http.Request) (any, error) {
	items, err := s.deps.Tags.Search(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("project_id"))
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicTags(items), "error": nil}, nil
}

func (s *Server) handleTagCreate(r *http.Request) (any, error) {
	var req tags.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	tag, err := s.deps.Tags.Create(r.Context(), req)
	if err != nil {
		return tagError(err)
	}
	return map[string]any{"data": publicTag(tag), "error": nil}, nil
}

func (s *Server) handleTagDelete(r *http.Request) (any, error) {
	tag, err := s.deps.Tags.Delete(r.Context(), r.PathValue("id"))
	if err != nil {
		return tagError(err)
	}
	return map[string]any{"data": publicTag(tag), "error": nil}, nil
}

func tagError(err error) (any, error) {
	if errors.Is(err, tags.ErrInvalidTag) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid tag")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Tag not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "Tag already exists"}, nil
	}
	return nil, err
}

func publicTags(items []domain.Tag) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicTag(item))
	}
	return result
}

func publicTag(tag domain.Tag) map[string]any {
	return map[string]any{
		"id":         tag.ID,
		"name":       tag.Name,
		"color":      tag.Color,
		"project_id": tag.ProjectID,
		"created_at": tag.CreatedAt,
	}
}
