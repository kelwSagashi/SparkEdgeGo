package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleScriptsList(r *http.Request) (any, error) {
	items, err := s.deps.Scripts.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicScripts(items), "error": nil}, nil
}

func (s *Server) handleScriptGet(r *http.Request) (any, error) {
	script, err := s.deps.Scripts.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": publicScript(script), "error": nil}, nil
}

func (s *Server) handleScriptCreate(r *http.Request) (any, error) {
	var req scripts.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	script, err := s.deps.Scripts.Create(r.Context(), req)
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": publicScript(script), "error": nil}, nil
}

func (s *Server) handleScriptUpdate(r *http.Request) (any, error) {
	var req scripts.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	script, err := s.deps.Scripts.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": publicScript(script), "error": nil}, nil
}

func (s *Server) handleScriptDelete(r *http.Request) (any, error) {
	if err := s.deps.Scripts.Delete(r.Context(), r.PathValue("id")); err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func scriptError(err error) (any, error) {
	if errors.Is(err, scripts.ErrInvalidScript) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid script")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Script not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "Script already exists"}, nil
	}
	return nil, err
}

func publicScripts(items []domain.DownloadedScript) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicScript(item))
	}
	return result
}

func publicScript(script domain.DownloadedScript) map[string]any {
	return map[string]any{
		"id":                script.ID,
		"name":              script.Name,
		"description":       script.Description,
		"author":            script.Author,
		"version":           script.Version,
		"source":            script.Source,
		"github_repo":       script.GitHubRepo,
		"github_ref":        script.GitHubRef,
		"local_path":        script.LocalPath,
		"main_file":         script.MainFile,
		"venv_path":         script.VenvPath,
		"requirements_file": script.RequirementsFile,
		"venv_ready":        script.VenvReady,
		"language":          script.Language,
		"tags":              script.Tags,
		"schema_config":     script.SchemaConfig,
		"created_at":        script.CreatedAt,
		"updated_at":        script.UpdatedAt,
	}
}
