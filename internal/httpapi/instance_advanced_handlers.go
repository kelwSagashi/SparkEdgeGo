package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
)

func (s *Server) handleInstanceAdvancedByProjectList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListByProject(r.Context(), r.PathValue("projectId"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstances(items), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedExecutionGet(r *http.Request) (any, error) {
	item, err := s.deps.Executions.FindByID(r.Context(), r.PathValue("executionId"))
	if err != nil {
		return executionError(err)
	}
	return map[string]any{"data": publicExecution(item), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedDestinationsList(r *http.Request) (any, error) {
	items, err := s.deps.Instances.ListDestinations(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicDestinationsWithMappings(items), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedDestinationAdd(r *http.Request) (any, error) {
	var req instances.DestinationPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	item, err := s.deps.Instances.AddDestination(r.Context(), r.PathValue("id"), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicDestinationWithMapping(item), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedDestinationUpdate(r *http.Request) (any, error) {
	var req instances.DestinationPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	item, err := s.deps.Instances.UpdateDestination(r.Context(), r.PathValue("destinationId"), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicDestinationWithMapping(item), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedDestinationDelete(r *http.Request) (any, error) {
	if err := s.deps.Instances.DeleteDestination(r.Context(), r.PathValue("destinationId")); err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedAvailableFields(r *http.Request) (any, error) {
	instance, err := s.deps.Instances.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	script, err := s.deps.Scripts.FindByID(r.Context(), instance.ScriptID)
	if err != nil {
		return scriptError(err)
	}
	return map[string]any{"data": map[string]any{"schema_config": script.SchemaConfig, "script_id": script.ID}, "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedMappingTest(_ *http.Request) (any, error) {
	return map[string]any{"data": map[string]any{"result": "test passed"}, "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedMappingSet(r *http.Request) (any, error) {
	var req instances.MappingData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	item, err := s.deps.Instances.SetDataMapping(r.Context(), r.PathValue("destinationId"), req)
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicDataMapping(item), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedActiveUpdate(r *http.Request) (any, error) {
	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), instances.Payload{Active: req.Active})
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedStatusGet(r *http.Request) (any, error) {
	instance, err := s.deps.Instances.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": map[string]any{"status": instance.Status, "active": instance.Active}, "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedTriggerConfigUpdate(r *http.Request) (any, error) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), instances.Payload{TriggerConfig: body})
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedScriptParamsUpdate(r *http.Request) (any, error) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	params := body
	if nested, ok := body["params"].(map[string]any); ok {
		params = nested
	}
	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), instances.Payload{ScriptParameters: params})
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func (s *Server) handleInstanceAdvancedFallbackConfigUpdate(r *http.Request) (any, error) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	instance, err := s.deps.Instances.Update(r.Context(), r.PathValue("id"), instances.Payload{FallbackConfig: body})
	if err != nil {
		return instanceError(err)
	}
	return map[string]any{"data": publicInstance(instance), "error": nil}, nil
}

func publicDestinationWithMapping(item domain.InstanceDestinationWithMapping) map[string]any {
	result := map[string]any{
		"destination": publicInstanceDestination(item.Destination),
		"mapping":     nil,
	}
	if item.Mapping != nil {
		result["mapping"] = publicDataMapping(*item.Mapping)
	}
	return result
}
