package httpapi

import (
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleFallbackList(r *http.Request) (any, error) {
	items, err := s.deps.DB.LocalFallback.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicFallbackItems(items), "error": nil}, nil
}

func (s *Server) handleFallbackStats(r *http.Request) (any, error) {
	items, err := s.deps.DB.LocalFallback.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	stats := map[string]int{"pending": 0, "failed": 0, "sent": 0, "total": len(items)}
	for _, item := range items {
		stats[string(item.Status)]++
	}
	return map[string]any{"data": stats, "error": nil}, nil
}

func (s *Server) handleFallbackFlush(r *http.Request) (any, error) {
	sent, err := s.deps.Runtime.FlushFallback(r.Context(), 10)
	if err != nil {
		return map[string]any{"data": map[string]any{"sent": sent}, "success": false, "error": err.Error()}, nil
	}
	return map[string]any{"data": map[string]any{"sent": sent}, "success": true, "error": nil}, nil
}

func (s *Server) handleFallbackRetry(r *http.Request) (any, error) {
	success, err := s.deps.Runtime.RetryFallbackItem(r.Context(), r.PathValue("id"))
	if err != nil {
		if err == sqlite.ErrNotFound {
			return map[string]any{"error": "Item not found"}, nil
		}
		return map[string]any{"data": map[string]any{"success": false}, "error": err.Error()}, nil
	}
	return map[string]any{"data": map[string]any{"success": success}, "error": nil}, nil
}

func (s *Server) handleFallbackDelete(r *http.Request) (any, error) {
	if err := s.deps.DB.LocalFallback.Delete(r.Context(), r.PathValue("id")); err != nil {
		if err == sqlite.ErrNotFound {
			return map[string]any{"error": "Item not found"}, nil
		}
		return nil, err
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func publicFallbackItems(items []domain.LocalFallbackItem) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicFallbackItem(item))
	}
	return result
}

func publicFallbackItem(item domain.LocalFallbackItem) map[string]any {
	return map[string]any{
		"id":             item.ID,
		"instance_id":    item.InstanceID,
		"destination_id": item.DestinationID,
		"execution_id":   item.ExecutionID,
		"status":         item.Status,
		"payload":        item.Payload,
		"filepath":       item.Filepath,
		"retry_count":    item.RetryCount,
		"last_retry_at":  item.LastRetryAt,
		"next_retry_at":  item.NextRetryAt,
		"last_error":     item.LastError,
		"created_at":     item.CreatedAt,
		"updated_at":     item.UpdatedAt,
	}
}
