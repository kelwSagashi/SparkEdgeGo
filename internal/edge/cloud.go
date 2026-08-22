package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type CloudClient interface {
	Login(ctx context.Context, email string, password string) (CloudLoginResult, error)
	Register(ctx context.Context, login CloudLoginResult, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error)
	Pair(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error)
	Unpair(ctx context.Context, edge domain.ProvisionedEdge) error
}

type HTTPCloudClient struct {
	BaseURL string
	Client  *http.Client
}

type CloudRequestError struct {
	StatusCode int
	Path       string
	Message    string
}

type CloudLoginResult struct {
	Token          string
	OrganizationID string
}

func (e *CloudRequestError) Error() string {
	if e == nil {
		return "cloud request failed"
	}
	if strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("cloud request failed (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("cloud request failed (%d)", e.StatusCode)
}

func NewHTTPCloudClient(baseURL string) *HTTPCloudClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:3000"
	}
	return &HTTPCloudClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *HTTPCloudClient) Login(ctx context.Context, email string, password string) (CloudLoginResult, error) {
	var response struct {
		Token string `json:"token"`
		User  struct {
			Organizations []struct {
				ID string `json:"id"`
			} `json:"organizations"`
		} `json:"user"`
	}
	if err := c.postJSON(ctx, "/auth/login", "", nil, map[string]any{"email": email, "password": password}, &response); err != nil {
		return CloudLoginResult{}, err
	}
	if response.Token == "" {
		return CloudLoginResult{}, fmt.Errorf("cloud login response missing token")
	}
	organizationID := ""
	if len(response.User.Organizations) > 0 {
		organizationID = strings.TrimSpace(response.User.Organizations[0].ID)
	}
	if organizationID == "" {
		return CloudLoginResult{}, fmt.Errorf("cloud login response missing organization")
	}
	return CloudLoginResult{
		Token:          response.Token,
		OrganizationID: organizationID,
	}, nil
}

func (c *HTTPCloudClient) Register(ctx context.Context, login CloudLoginResult, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error) {
	payload := map[string]any{
		"name":            edgeName,
		"user_token":      login.Token,
		"organizationId":  login.OrganizationID,
		"organization_id": login.OrganizationID,
	}
	for key, value := range metadata {
		payload[key] = value
	}
	var response domain.ProvisionedEdge
	headers := map[string]string{"X-Organization-ID": login.OrganizationID}
	if err := c.postJSON(ctx, "/edges/register", login.Token, headers, payload, &response); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	return validateRegistration(response)
}

func (c *HTTPCloudClient) Pair(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error) {
	var response domain.ProvisionedEdge
	if err := c.postJSON(ctx, "/edges/pair", "", nil, map[string]any{"token": token, "edge_name": edgeName, "name": edgeName, "metadata": metadata}, &response); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	return validateRegistration(response)
}

func (c *HTTPCloudClient) Unpair(ctx context.Context, edge domain.ProvisionedEdge) error {
	return c.postJSON(ctx, "/edges/unpair-self", "", map[string]string{
		"x-edge-id":       edge.EdgeID,
		"x-edge-username": edge.MQTT.Username,
		"x-edge-password": edge.MQTT.Password,
	}, map[string]any{"edgeId": edge.EdgeID}, nil)
}

func (c *HTTPCloudClient) postJSON(ctx context.Context, path string, bearer string, extraHeaders map[string]string, payload any, target any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	for key, value := range extraHeaders {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("cloud request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return &CloudRequestError{
			StatusCode: res.StatusCode,
			Path:       path,
			Message:    cloudErrorMessage(bodyBytes, res.Status),
		}
	}
	if target != nil {
		if err := json.NewDecoder(res.Body).Decode(target); err != nil {
			return err
		}
	}
	return nil
}

func validateRegistration(edge domain.ProvisionedEdge) (domain.ProvisionedEdge, error) {
	if edge.EdgeID == "" || edge.MQTT.URL == "" || edge.MQTT.Username == "" || edge.MQTT.Password == "" {
		return domain.ProvisionedEdge{}, fmt.Errorf("invalid cloud registration response")
	}
	edge.Provisioned = true
	return edge, nil
}

func cloudErrorMessage(body []byte, fallback string) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fallback
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		for _, key := range []string{"message", "error", "detail"} {
			if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
				return value
			}
		}
	}
	return trimmed
}
