package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
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

func (s *Server) handleScriptFileContent(r *http.Request) (any, error) {
	content, err := s.deps.Scripts.FileContent(r.Context(), r.PathValue("id"), r.PathValue("filename"))
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": content}, nil
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

func (s *Server) handleScriptUploadInspect(r *http.Request) (any, error) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid multipart form")
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return ResponsePayload{Status: http.StatusBadRequest, Body: map[string]any{"error": "No file uploaded"}}, nil
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", "spark-edge_upload_*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, file); err != nil {
		_ = tempFile.Close()
		return nil, err
	}
	if err := tempFile.Close(); err != nil {
		return nil, err
	}

	result, err := s.deps.Scripts.InspectZip(tempFile.Name())
	if err != nil {
		return nil, err
	}
	if !result.HasSparkit {
		_ = os.RemoveAll(result.TempFolder)
		return ResponsePayload{
			Status: http.StatusBadRequest,
			Body:   map[string]any{"error": "O script e invalido. E obrigatorio ter um arquivo requirements.txt contendo a biblioteca \"sparkit\"."},
		}, nil
	}

	return map[string]any{"data": result}, nil
}

func (s *Server) handleScriptUploadFinalize(r *http.Request) (any, error) {
	var req scripts.FinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := s.deps.Scripts.FinalizeUpload(r.Context(), req)
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{
		"data": map[string]any{
			"script": publicScript(result.Script),
			"schema": result.Schema,
		},
	}, nil
}

func (s *Server) handleScriptSamplesList(r *http.Request) (any, error) {
	samples, err := s.deps.Scripts.ListSamples()
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": samples}, nil
}

func (s *Server) handleScriptSampleSchema(r *http.Request) (any, error) {
	schema, err := s.deps.Scripts.SampleSchema(r.Context(), r.PathValue("name"))
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": schema}, nil
}

func (s *Server) handleScriptPlaygroundRun(r *http.Request) (any, error) {
	var req scripts.PlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	result, err := s.deps.Scripts.RunPlayground(r.Context(), req)
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": publicScriptExecutionResult(result)}, nil
}

func scriptError(err error) (any, error) {
	if errors.Is(err, scripts.ErrInvalidScript) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid script")
	}
	if errors.Is(err, scripts.ErrScriptFileNotFound) {
		return ResponsePayload{Status: http.StatusNotFound, Body: map[string]any{"error": "File not found"}}, nil
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Script not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "Script already exists"}, nil
	}
	return nil, NewHTTPError(http.StatusInternalServerError, err.Error())
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

func publicScriptExecutionResult(result domain.ScriptResult) map[string]any {
	if result.Data != nil {
		return result.Data
	}

	payload := map[string]any{
		"stdout": nil,
		"stderr": nil,
	}

	if strings.TrimSpace(result.Stdout) != "" {
		payload["stdout"] = result.Stdout
	}
	if strings.TrimSpace(result.Stderr) != "" {
		payload["stderr"] = result.Stderr
	}

	return payload
}
