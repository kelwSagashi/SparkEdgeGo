package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

const authCookieName = "spark_edge_token"

func (s *Server) handleRegister(r *http.Request) (any, error) {
	if s.deps.Auth == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "auth service unavailable")
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := s.deps.Auth.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return map[string]any{"data": nil, "error": "User already exists"}, nil
		}
		if errors.Is(err, auth.ErrInvalidCredential) {
			return nil, NewHTTPError(http.StatusBadRequest, "email and password are required")
		}
		return nil, err
	}

	return map[string]any{"data": publicUser(user), "error": nil}, nil
}

func (s *Server) handleLogin(r *http.Request) (any, error) {
	if s.deps.Auth == nil {
		return nil, NewHTTPError(http.StatusServiceUnavailable, "auth service unavailable")
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := s.deps.Auth.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredential) || errors.Is(err, sqlite.ErrNotFound) {
			return map[string]any{
				"data":  map[string]any{"token": nil, "user": nil},
				"error": "Invalid credentials",
			}, nil
		}
		return nil, err
	}

	return loginResponseWriter(result.Token, publicUser(result.User)), nil
}

func (s *Server) handleMe(r *http.Request) (any, error) {
	identity, ok := CurrentIdentity(r.Context())
	if !ok || !identity.Verified {
		return map[string]any{"data": nil}, nil
	}

	return map[string]any{"data": map[string]any{
		"id":     identity.UserID,
		"email":  identity.Email,
		"role":   identity.Role,
		"source": identity.Source,
	}}, nil
}

func loginResponseWriter(token string, user map[string]any) ResponsePayload {
	return ResponsePayload{
		Status: http.StatusOK,
		Headers: map[string]string{
			"Set-Cookie": (&http.Cookie{
				Name:     authCookieName,
				Value:    token,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   7 * 24 * 60 * 60,
			}).String(),
		},
		Body: map[string]any{
			"data":  map[string]any{"token": token, "user": user},
			"error": nil,
		},
	}
}

func publicUser(user domain.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
