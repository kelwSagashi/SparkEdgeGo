package httpapi

import (
	"errors"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleCloudSyncList(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return map[string]any{"data": []any{}}, nil
	}
	items, err := s.deps.CloudSync.ListRecent(r.Context(), 100)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Server) handleCloudSyncStats(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return map[string]any{"configured": false}, nil
	}
	return s.deps.CloudSync.Stats(r.Context())
}

func (s *Server) handleCloudSyncFlush(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	return s.deps.CloudSync.Flush(r.Context(), 50)
}

func (s *Server) handleCloudSyncRetry(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	result, err := s.deps.CloudSync.RetryItem(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "cloud sync item not found")
		}
		return nil, err
	}
	return result, nil
}

func (s *Server) handleCloudSyncDelete(r *http.Request) (any, error) {
	if s.deps.CloudSync == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "cloud sync unavailable")
	}
	if err := s.deps.CloudSync.DeleteItem(r.Context(), r.PathValue("id")); err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "cloud sync item not found")
		}
		return nil, err
	}
	return map[string]any{
		"success": true,
		"id":      r.PathValue("id"),
	}, nil
}
