package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type contextKey string

const userContextKey contextKey = "sparkedge-user"

type AuthIdentity struct {
	Source   string
	Token    string
	Verified bool
	UserID   string
	Email    string
	Role     domain.UserRole
}

type IdentityVerifier interface {
	VerifyToken(ctx context.Context, token string) (domain.User, error)
	VerifyAPIKey(ctx context.Context, apiKey string) (domain.User, error)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		origin := r.Header.Get("Origin")
		if strings.HasPrefix(origin, "http://localhost:5") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler, verifier IdentityVerifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := resolveAuthIdentity(r)
		if identity != nil {
			verifyIdentity(r.Context(), identity, verifier)
			r = r.WithContext(context.WithValue(r.Context(), userContextKey, *identity))
		}

		if isProtectedPath(r.URL.Path) && (identity == nil || !identity.Verified) {
			Respond(w, http.StatusUnauthorized, map[string]any{"message": "Unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func verifyIdentity(ctx context.Context, identity *AuthIdentity, verifier IdentityVerifier) {
	if verifier == nil {
		return
	}

	var user domain.User
	var err error

	switch identity.Source {
	case "cookie", "bearer":
		user, err = verifier.VerifyToken(ctx, identity.Token)
	case "api-key":
		user, err = verifier.VerifyAPIKey(ctx, identity.Token)
	default:
		return
	}

	if err != nil || user.ID == "" || !user.IsActive {
		return
	}

	identity.Verified = true
	identity.UserID = user.ID
	identity.Email = user.Email
	identity.Role = user.Role
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				RespondError(w, NewHTTPError(http.StatusInternalServerError, "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func resolveAuthIdentity(r *http.Request) *AuthIdentity {
	if cookie, err := r.Cookie("spark_edge_token"); err == nil && cookie.Value != "" {
		return &AuthIdentity{Source: "cookie", Token: cookie.Value}
	}

	if bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "); bearer != "" && bearer != r.Header.Get("Authorization") {
		return &AuthIdentity{Source: "bearer", Token: bearer}
	}

	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return &AuthIdentity{Source: "api-key", Token: apiKey}
	}

	return nil
}

func isProtectedPath(path string) bool {
	for _, prefix := range protectedPrefixes() {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func protectedPrefixes() []string {
	return []string{
		"/api/instances",
		"/api/instance-advanced",
		"/api/scripts",
		"/api/devices",
		"/api/servers",
		"/api/users",
		"/api/server-types",
		"/api/credentials",
		"/api/tags",
		"/api/executions",
		"/api/fallback",
		"/api/projects",
	}
}

func CurrentIdentity(ctx context.Context) (AuthIdentity, bool) {
	identity, ok := ctx.Value(userContextKey).(AuthIdentity)
	return identity, ok
}
