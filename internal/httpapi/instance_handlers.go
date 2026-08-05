package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleInstancesList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstancesActiveList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListActive(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstancesByProjectList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListByProject(r.Context(), r.PathValue("project_id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstanceGet(r *http.Request) (any, error) {
	result, err := s.deps.Instances.GetWithDestinations(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{
		"data": map[string]any{
			"instance":     publicInstance(result.Instance),
			"destinations": result.Destinations,
		},
		"error": nil,
	}, nil
}

func (s *Server) handleInstanceCreate(r *http.Request) (any, error) {
	var req instances.Payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified {
		req.CreatedBy = identity.UserID
	}

	instance, err := s.deps.Instances.Create(r.Context(), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceUpdate(r *http.Request) (any, error) {
	var req instances.Payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceDelete(r *http.Request) (any, error) {
	if err := s.deps.Instances.Delete(r.Context(), r.PathValue("id")); err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleInstanceTrigger(r *http.Request) (any, error) {
	return map[string]any{
		"data": map[string]any{
			"status": "queued",
			"note":   "runtime trigger will be connected after destinations and executions are migrated",
		},
	}, nil
}

func (s *Server) handleInstanceExecutionsList(r *http.Request) (any, error) {
	return map[string]any{"data": []any{}, "error": nil}, nil
}

func instanceError(err error) (any, error) {
	if errors.Is(err, instances.ErrInvalidInstance) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid instance")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Instance not found"}, nil
	}
	return nil, err
}

func publicInstances(items []domain.Instance) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicInstance(item))
	}
	return result
}

func publicInstance(instance domain.Instance) map[string]any {
	return map[string]any{
		"id":                              instance.ID,
		"name":                            instance.Name,
		"description":                     instance.Description,
		"tags":                            instance.Tags,
		"status":                          instance.Status,
		"active":                          instance.Active,
		"project_id":                      instance.ProjectID,
		"device_id":                       instance.DeviceID,
		"script_id":                       instance.ScriptID,
		"include_device_data":             instance.IncludeDeviceData,
		"script_parameters":               instance.ScriptParameters,
		"trigger_type":                    instance.TriggerType,
		"trigger_config":                  instance.TriggerConfig,
		"fallback_enabled":                instance.FallbackEnabled,
		"fallback_strategy":               instance.FallbackStrategy,
		"fallback_retry_interval_seconds": instance.FallbackRetryIntervalSeconds,
		"on_error_action":                 instance.OnErrorAction,
		"on_error_config":                 instance.OnErrorConfig,
		"created_by":                      instance.CreatedBy,
		"created_at":                      instance.CreatedAt,
		"updated_at":                      instance.UpdatedAt,
	}
}
