package httpapi

import "net/http"

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
		return map[string]any{"success": false, "sent": sent, "error": err.Error()}, nil
	}
	return map[string]any{"success": true, "sent": sent}, nil
}
