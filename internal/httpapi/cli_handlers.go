package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/edge"
)

func (s *Server) handleCliOnboardingGet(r *http.Request) (any, error) {
	config, complete, err := s.deps.Edge.GetOnboarding(r.Context())
	if err != nil {
		return nil, err
	}
	var data any
	if config.ID != "" {
		data = map[string]any{
			"name":        config.EdgeName,
			"description": config.Description,
			"lat":         config.Lat,
			"lng":         config.Lng,
			"tags":        config.Tags,
		}
	}
	return map[string]any{"complete": complete, "data": data}, nil
}

func (s *Server) handleCliOnboardingSave(r *http.Request) (any, error) {
	var req edge.OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if _, err := s.deps.Edge.SaveOnboarding(r.Context(), req); err != nil {
		return cliEdgeError(err)
	}
	return map[string]any{"success": true}, nil
}

func (s *Server) handleCliStatus(r *http.Request) (any, error) {
	status, err := s.deps.Edge.Status(r.Context())
	if err != nil {
		return nil, err
	}
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified && s.deps.MQTT != nil && s.deps.MQTT.IsConnected() {
		_ = s.deps.MQTT.PublishContext(r.Context(), map[string]any{"id": identity.UserID})
	}
	return status, nil
}

func (s *Server) handleCliPair(r *http.Request) (any, error) {
	var req edge.PairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	registration, err := s.deps.Edge.Pair(r.Context(), req)
	if err != nil {
		return cliEdgeError(err)
	}
	publishUserContext(r, s)
	enqueueCloudConnectionEvent(r, s, registration.EdgeID, "paired", "Edge paired with Spark Cloud")
	return map[string]any{"success": true, "edge_id": registration.EdgeID, "edge_name": registration.EdgeName}, nil
}

func (s *Server) handleCliConnect(r *http.Request) (any, error) {
	var req edge.ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	registration, err := s.deps.Edge.Connect(r.Context(), req)
	if err != nil {
		return cliEdgeError(err)
	}
	publishUserContext(r, s)
	enqueueCloudConnectionEvent(r, s, registration.EdgeID, "registered", "Edge connected to Spark Cloud")
	return map[string]any{"success": true, "edge_id": registration.EdgeID, "edge_name": registration.EdgeName}, nil
}

func (s *Server) handleCliDisconnect(r *http.Request) (any, error) {
	if provisioned, err := s.deps.Edge.Load(r.Context()); err == nil && provisioned.EdgeID != "" {
		enqueueCloudConnectionEvent(r, s, provisioned.EdgeID, "mqtt_disconnected", "Edge disconnected from MQTT broker")
	}
	if err := s.deps.Edge.Disconnect(r.Context()); err != nil {
		return nil, err
	}
	return map[string]any{"success": true, "message": "Edge desconectado (identidade preservada)."}, nil
}

func (s *Server) handleCliReconnect(r *http.Request) (any, error) {
	if err := s.deps.Edge.Reconnect(r.Context()); err != nil {
		return cliEdgeError(err)
	}
	if provisioned, err := s.deps.Edge.Load(r.Context()); err == nil && provisioned.EdgeID != "" {
		enqueueCloudConnectionEvent(r, s, provisioned.EdgeID, "mqtt_reconnected", "Edge reconnected to MQTT broker")
	}
	return map[string]any{"success": true}, nil
}

func (s *Server) handleCliRemove(r *http.Request) (any, error) {
	if err := s.deps.Edge.Remove(r.Context()); err != nil {
		return cliEdgeError(err)
	}
	return map[string]any{"success": true, "message": "Edge resetado com sucesso (conexao removida)."}, nil
}

func (s *Server) handleCliMQTTConfigGet(r *http.Request) (any, error) {
	provisioned, err := s.deps.Edge.Load(r.Context())
	if errors.Is(err, edge.ErrNotProvisioned) {
		return map[string]any{"broker_url": "", "username": "", "has_password": false}, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"broker_url":   provisioned.MQTT.URL,
		"username":     provisioned.MQTT.Username,
		"has_password": provisioned.MQTT.Password != "",
		"password":     provisioned.MQTT.Password,
	}, nil
}

func cliEdgeError(err error) (any, error) {
	if errors.Is(err, edge.ErrInvalidOnboarding) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid edge onboarding")
	}
	if errors.Is(err, edge.ErrNotProvisioned) {
		return nil, NewHTTPError(http.StatusBadRequest, "Edge nao provisionado. Conecte-se ao Spark Cloud primeiro.")
	}
	return nil, err
}

func publishUserContext(r *http.Request, s *Server) {
	if identity, ok := CurrentIdentity(r.Context()); ok && identity.Verified && s.deps.MQTT != nil {
		_ = s.deps.MQTT.PublishContext(r.Context(), map[string]any{"id": identity.UserID})
	}
}

func publicProvisionedEdge(edge domain.ProvisionedEdge) map[string]any {
	return map[string]any{"edge_id": edge.EdgeID, "edge_name": edge.EdgeName, "provisioned": edge.Provisioned}
}

func enqueueCloudConnectionEvent(r *http.Request, s *Server, edgeID string, eventType string, message string) {
	if s == nil || s.deps.CloudSync == nil || edgeID == "" {
		return
	}
	now := time.Now().UTC()
	_, _ = s.deps.CloudSync.EnqueueEvent(r.Context(), "edge_connection", 80, map[string]any{
		"message_id":  fmt.Sprintf("%s-%d", eventType, now.UnixNano()),
		"edge_id":     edgeID,
		"type":        "edge_connection",
		"event":       eventType,
		"message":     message,
		"occurred_at": now.Format(time.RFC3339),
		"timestamp":   now.Format(time.RFC3339),
	})
}
