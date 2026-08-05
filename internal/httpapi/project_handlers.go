package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleProjectsList(r *http.Request) (any, error) {
	identity, ok := CurrentIdentity(r.Context())
	if !ok || !identity.Verified {
		return map[string]any{"data": []any{}, "error": nil}, nil
	}

	items, err := s.deps.Projects.ListByOwner(r.Context(), identity.UserID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicProjects(items), "error": nil}, nil
}

func (s *Server) handleProjectGet(r *http.Request) (any, error) {
	project, err := s.deps.Projects.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return projectError(err)
	}
	return map[string]any{"data": publicProject(project), "error": nil}, nil
}

func (s *Server) handleProjectCreate(r *http.Request) (any, error) {
	identity, ok := CurrentIdentity(r.Context())
	if !ok || !identity.Verified {
		return nil, NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req projects.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.OwnerID = identity.UserID

	project, err := s.deps.Projects.Create(r.Context(), req)
	if err != nil {
		return projectError(err)
	}
	return map[string]any{"data": publicProject(project), "error": nil}, nil
}

func (s *Server) handleProjectUpdate(r *http.Request) (any, error) {
	var req projects.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	project, err := s.deps.Projects.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return projectError(err)
	}
	return map[string]any{"data": publicProject(project), "error": nil}, nil
}

func (s *Server) handleProjectDelete(r *http.Request) (any, error) {
	if err := s.deps.Projects.Delete(r.Context(), r.PathValue("id")); err != nil {
		return projectError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleProjectMembersList(r *http.Request) (any, error) {
	members, err := s.deps.Projects.ListMembers(r.Context(), r.PathValue("id"))
	if err != nil {
		return projectError(err)
	}
	return map[string]any{"data": publicProjectMembers(members), "error": nil}, nil
}

func (s *Server) handleProjectMemberAdd(r *http.Request) (any, error) {
	var req projects.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	member, err := s.deps.Projects.AddMember(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return projectError(err)
	}
	return map[string]any{"data": publicProjectMember(member), "error": nil}, nil
}

func projectError(err error) (any, error) {
	if errors.Is(err, projects.ErrInvalidProject) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid project")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Project not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "Project already exists"}, nil
	}
	return nil, err
}

func publicProjects(items []domain.Project) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicProject(item))
	}
	return result
}

func publicProject(project domain.Project) map[string]any {
	return map[string]any{
		"id":          project.ID,
		"name":        project.Name,
		"key":         project.Key,
		"description": project.Description,
		"visibility":  project.Visibility,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	}
}

func publicProjectMembers(items []domain.ProjectMember) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicProjectMember(item))
	}
	return result
}

func publicProjectMember(member domain.ProjectMember) map[string]any {
	return map[string]any{
		"id":         member.ID,
		"project_id": member.ProjectID,
		"user_id":    member.UserID,
		"role":       member.Role,
		"created_at": member.CreatedAt,
	}
}
