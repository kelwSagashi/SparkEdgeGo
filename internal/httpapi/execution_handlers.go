package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/executions"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleExecutionsList(r *http.Request) (any, error) {
	items, err := s.deps.Executions.ListAll(r.Context(), limitFromQuery(r, 100))
	if err != nil {
		return executionError(err)
	}
	return map[string]any{"data": publicExecutions(items), "error": nil}, nil
}

func (s *Server) handleExecutionGet(r *http.Request) (any, error) {
	item, err := s.deps.Executions.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return executionError(err)
	}
	return map[string]any{"data": publicExecution(item), "error": nil}, nil
}

func (s *Server) handleExecutionsByInstanceList(r *http.Request) (any, error) {
	items, err := s.deps.Executions.ListByInstance(r.Context(), r.PathValue("instance_id"), limitFromQuery(r, 50))
	if err != nil {
		return executionError(err)
	}
	return map[string]any{"data": publicExecutions(items), "error": nil}, nil
}

func executionError(err error) (any, error) {
	if errors.Is(err, executions.ErrInvalidExecution) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid execution")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Execution not found"}, nil
	}
	return nil, err
}

func publicExecutions(items []domain.InstanceExecution) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicExecution(item))
	}
	return result
}

func publicExecution(execution domain.InstanceExecution) map[string]any {
	return map[string]any{
		"id":               execution.ID,
		"instance_id":      execution.InstanceID,
		"status":           execution.Status,
		"trigger_type":     execution.TriggerType,
		"started_at":       execution.StartedAt,
		"finished_at":      execution.FinishedAt,
		"duration_ms":      execution.DurationMS,
		"logs":             execution.Logs,
		"output":           execution.Output,
		"error_message":    execution.ErrorMessage,
		"destination_sent": execution.DestinationSent,
		"fallback_used":    execution.FallbackUsed,
		"created_at":       execution.CreatedAt,
	}
}

func limitFromQuery(r *http.Request, fallback int) int {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return fallback
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return fallback
	}
	return limit
}
