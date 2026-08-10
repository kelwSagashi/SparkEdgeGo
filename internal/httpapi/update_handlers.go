package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleUpdateCheck(r *http.Request) (any, error) {
	if s.deps.Updater == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "update service unavailable")
	}

	result, err := s.deps.Updater.Check(r.Context())
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"data":  result,
		"error": nil,
	}, nil
}

func (s *Server) handleUpdateStatus(_ *http.Request) (any, error) {
	if s.deps.Updater == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "update service unavailable")
	}

	state, err := s.deps.Updater.LoadState()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"data":  state,
		"error": nil,
	}, nil
}

func (s *Server) handleUpdateDownload(r *http.Request) (any, error) {
	if s.deps.Updater == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "update service unavailable")
	}

	result, err := s.deps.Updater.DownloadLatest(r.Context())
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"data":  result,
		"error": nil,
	}, nil
}

func (s *Server) handleUpdateApply(r *http.Request) (any, error) {
	if s.deps.Updater == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "update service unavailable")
	}

	var req struct {
		DownloadedPath string `json:"downloaded_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := s.deps.Updater.ApplyDownloaded(r.Context(), req.DownloadedPath)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"data":  result,
		"error": nil,
	}, nil
}
