package httpapi

import (
	"net/http"
	"strings"
)

func (s *Server) routeInstances(w http.ResponseWriter, r *http.Request) {
	segments := splitRouteSegments(r.URL.Path, "/api/instances/")
	switch {
	case r.Method == http.MethodGet && len(segments) == 1 && segments[0] == "active":
		Adapt(s.handleInstancesActiveList)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[0] == "project":
		r = withPathValue(r, "project_id", segments[1])
		Adapt(s.handleInstancesByProjectList)(w, r)
	case r.Method == http.MethodGet && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceGet)(w, r)
	case r.Method == http.MethodPut && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceUpdate)(w, r)
	case r.Method == http.MethodDelete && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceDelete)(w, r)
	case r.Method == http.MethodPost && len(segments) == 2 && segments[1] == "trigger":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceTrigger)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "executions":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceExecutionsList)(w, r)
	default:
		RespondError(w, NewHTTPError(http.StatusNotFound, "route not found"))
	}
}

func (s *Server) routeInstanceAdvanced(w http.ResponseWriter, r *http.Request) {
	segments := splitRouteSegments(r.URL.Path, "/api/instance-advanced/")
	switch {
	case r.Method == http.MethodGet && len(segments) == 2 && segments[0] == "project":
		r = withPathValue(r, "projectId", segments[1])
		Adapt(s.handleInstanceAdvancedByProjectList)(w, r)
	case len(segments) == 0:
		RespondError(w, NewHTTPError(http.StatusNotFound, "route not found"))
	case r.Method == http.MethodGet && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceGet)(w, r)
	case r.Method == http.MethodPut && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceUpdate)(w, r)
	case r.Method == http.MethodDelete && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceDelete)(w, r)
	case r.Method == http.MethodPost && len(segments) == 2 && segments[1] == "trigger":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceTrigger)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "executions":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceExecutionsList)(w, r)
	case r.Method == http.MethodGet && len(segments) == 3 && segments[1] == "executions":
		r = withPathValue(r, "id", segments[0])
		r = withPathValue(r, "executionId", segments[2])
		Adapt(s.handleInstanceAdvancedExecutionGet)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "destinations":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedDestinationsList)(w, r)
	case r.Method == http.MethodPost && len(segments) == 2 && segments[1] == "destinations":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedDestinationAdd)(w, r)
	case r.Method == http.MethodPut && len(segments) == 3 && segments[1] == "destinations":
		r = withPathValue(r, "id", segments[0])
		r = withPathValue(r, "destinationId", segments[2])
		Adapt(s.handleInstanceAdvancedDestinationUpdate)(w, r)
	case r.Method == http.MethodDelete && len(segments) == 3 && segments[1] == "destinations":
		r = withPathValue(r, "id", segments[0])
		r = withPathValue(r, "destinationId", segments[2])
		Adapt(s.handleInstanceAdvancedDestinationDelete)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "available-fields":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedAvailableFields)(w, r)
	case r.Method == http.MethodPost && len(segments) == 3 && segments[1] == "mappings" && segments[2] == "test":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedMappingTest)(w, r)
	case r.Method == http.MethodPut && len(segments) == 4 && segments[1] == "destinations" && segments[3] == "mapping":
		r = withPathValue(r, "id", segments[0])
		r = withPathValue(r, "destinationId", segments[2])
		Adapt(s.handleInstanceAdvancedMappingSet)(w, r)
	case r.Method == http.MethodPut && len(segments) == 2 && segments[1] == "active":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedActiveUpdate)(w, r)
	case r.Method == http.MethodGet && len(segments) == 2 && segments[1] == "status":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedStatusGet)(w, r)
	case r.Method == http.MethodPut && len(segments) == 2 && segments[1] == "trigger-config":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedTriggerConfigUpdate)(w, r)
	case r.Method == http.MethodPut && len(segments) == 2 && segments[1] == "script-params":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedScriptParamsUpdate)(w, r)
	case r.Method == http.MethodPut && len(segments) == 2 && segments[1] == "fallback-config":
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleInstanceAdvancedFallbackConfigUpdate)(w, r)
	default:
		RespondError(w, NewHTTPError(http.StatusNotFound, "route not found"))
	}
}

func splitRouteSegments(path string, prefix string) []string {
	if !strings.HasPrefix(path, prefix) {
		return nil
	}
	remainder := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if remainder == "" {
		return nil
	}
	return strings.Split(remainder, "/")
}

func withPathValue(r *http.Request, key string, value string) *http.Request {
	r.SetPathValue(key, value)
	return r
}

func (s *Server) routeScripts(w http.ResponseWriter, r *http.Request) {
	segments := splitRouteSegments(r.URL.Path, "/api/scripts/")
	switch {
	case r.Method == http.MethodGet && len(segments) == 2 && segments[0] == "samples" && segments[1] == "list":
		Adapt(s.handleScriptSamplesList)(w, r)
	case r.Method == http.MethodGet && len(segments) == 3 && segments[0] == "samples" && segments[2] == "schema":
		r = withPathValue(r, "name", segments[1])
		Adapt(s.handleScriptSampleSchema)(w, r)
	case r.Method == http.MethodGet && len(segments) == 3 && segments[1] == "contents":
		r = withPathValue(r, "id", segments[0])
		r = withPathValue(r, "filename", segments[2])
		Adapt(s.handleScriptFileContent)(w, r)
	case r.Method == http.MethodGet && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleScriptGet)(w, r)
	case r.Method == http.MethodPut && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleScriptUpdate)(w, r)
	case r.Method == http.MethodDelete && len(segments) == 1:
		r = withPathValue(r, "id", segments[0])
		Adapt(s.handleScriptDelete)(w, r)
	default:
		RespondError(w, NewHTTPError(http.StatusNotFound, "route not found"))
	}
}
