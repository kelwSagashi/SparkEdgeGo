package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (s *Server) handleCliConfigGet(_ *http.Request) (any, error) {
	jwtSecret := envOr("JWT_SECRET", "dev-secret")
	return map[string]any{
		"cloud": map[string]any{
			"url":      envOr("SPARK_CLOUD_URL", "http://localhost:3000"),
			"mqtt_url": envOr("MQTT_URL", ""),
		},
		"db": map[string]any{
			"file": s.deps.DB.Path,
		},
		"auth": map[string]any{
			"jwt_secret": maskSecret(jwtSecret),
			"is_default": jwtSecret == "dev-secret",
		},
		"server": map[string]any{
			"port": envOr("SPARKEDGE_HTTP_ADDR", ":3009"),
		},
	}, nil
}

func (s *Server) handleCliConfigUpdate(r *http.Request) (any, error) {
	var req struct {
		Cloud  map[string]any `json:"cloud"`
		DB     map[string]any `json:"db"`
		Auth   map[string]any `json:"auth"`
		Server map[string]any `json:"server"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if secret, ok := stringConfig(req.Auth, "jwt_secret"); ok && secret != "" && len(secret) < 8 {
		return nil, NewHTTPError(http.StatusBadRequest, "jwt_secret deve ter pelo menos 8 caracteres.")
	}
	if port, ok := stringConfig(req.Server, "port"); ok && port != "" {
		if numeric := strings.TrimPrefix(port, ":"); numeric != "" {
			value, err := strconv.Atoi(numeric)
			if err != nil || value < 1 || value > 65535 {
				return nil, NewHTTPError(http.StatusBadRequest, "Porta invalida. Use um valor entre 1 e 65535.")
			}
		}
	}
	if len(req.Cloud) == 0 && len(req.DB) == 0 && len(req.Auth) == 0 && len(req.Server) == 0 {
		return nil, NewHTTPError(http.StatusBadRequest, "Nenhum campo valido para atualizar.")
	}
	return map[string]any{"success": true, "message": "Configuracao validada. Reinicie o servico com as variaveis de ambiente atualizadas para aplicar as mudancas."}, nil
}

func envOr(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", min(len(secret)-4, 20))
}

func stringConfig(values map[string]any, key string) (string, bool) {
	if values == nil {
		return "", false
	}
	value, ok := values[key]
	if !ok {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), true
	default:
		return strings.TrimSpace(strconv.FormatFloat(toFloat64(typed), 'f', -1, 64)), true
	}
}

func toFloat64(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	default:
		return 0
	}
}
