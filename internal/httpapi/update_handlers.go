package httpapi

import (
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
