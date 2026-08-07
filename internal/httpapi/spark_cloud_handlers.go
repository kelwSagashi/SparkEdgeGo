package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func (s *Server) handleSparkCloudLogin(r *http.Request) (any, error) {
	return map[string]any{"token": "spark-cloud-token-" + randomCloudID(16)}, nil
}

func (s *Server) handleSparkCloudEdgeRegister(r *http.Request) (any, error) {
	var req struct {
		Name string   `json:"name"`
		Lat  any      `json:"lat"`
		Lng  any      `json:"lng"`
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	edgeID := "edge-" + randomCloudID(4)
	edgeName := strings.TrimSpace(req.Name)
	if edgeName == "" {
		edgeName = "Edge " + edgeID
	}

	return map[string]any{
		"edge_id":   edgeID,
		"edge_name": edgeName,
		"mqtt": map[string]any{
			"url":      sparkCloudMQTTURL(s.deps.RuntimeCfg.MQTTURL),
			"username": "spark-user-" + edgeID,
			"password": "spark-pass-" + randomCloudID(4),
		},
	}, nil
}

func randomCloudID(bytesLen int) string {
	data := make([]byte, bytesLen)
	if _, err := rand.Read(data); err != nil {
		return "local"
	}
	return hex.EncodeToString(data)
}

func sparkCloudMQTTURL(runtimeValue string) string {
	if value := strings.TrimSpace(runtimeValue); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("MQTT_URL")); value != "" {
		return value
	}
	return "mqtt://localhost:1883"
}
