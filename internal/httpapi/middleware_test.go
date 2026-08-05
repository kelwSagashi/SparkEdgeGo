package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareAllowsOpenRoutesWithoutIdentity(t *testing.T) {
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected open route to pass, got status %d", res.Code)
	}
}

func TestAuthMiddlewareRejectsProtectedRouteWithoutVerifiedIdentity(t *testing.T) {
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/instances", nil)
	req.Header.Set("Authorization", "Bearer presented-but-not-verified")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected protected route to reject unverified identity, got status %d", res.Code)
	}
}

func TestResolveAuthIdentityOrder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.AddCookie(&http.Cookie{Name: "spark_edge_token", Value: "cookie-token"})
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.Header.Set("X-API-Key", "api-key-token")

	identity := resolveAuthIdentity(req)
	if identity == nil {
		t.Fatal("expected identity candidate")
	}
	if identity.Source != "cookie" {
		t.Fatalf("expected cookie to be resolved first, got %s", identity.Source)
	}
}
