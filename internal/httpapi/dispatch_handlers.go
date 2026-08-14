package httpapi

import (
	"encoding/json"
	"net/http"
)

type eventDispatchRequest struct {
	EventName string         `json:"event_name"`
	Payload   map[string]any `json:"payload"`
}

type stateDispatchRequest struct {
	Payload map[string]any `json:"payload"`
}

func (s *Server) handleEventDispatch(r *http.Request) (any, error) {
	if s.deps.DispatchEvent == nil {
		return nil, NewHTTPError(http.StatusNotImplemented, "event dispatcher unavailable")
	}
	var req eventDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	result, err := s.deps.DispatchEvent(r.Context(), req.EventName, req.Payload)
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": result, "error": nil}, nil
}

func (s *Server) handleStateDispatch(r *http.Request) (any, error) {
	if s.deps.DispatchStateChange == nil {
		return nil, NewHTTPError(http.StatusNotImplemented, "state dispatcher unavailable")
	}
	var req stateDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	result, err := s.deps.DispatchStateChange(r.Context(), req.Payload)
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": result, "error": nil}, nil
}
