package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type CloudClient interface {
	Login(ctx context.Context, email string, password string) (string, error)
	Register(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error)
	Pair(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error)
	Unpair(ctx context.Context, edgeID string) error
}

type HTTPCloudClient struct {
	BaseURL string
	Client  *http.Client
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

func (c *HTTPCloudClient) Login(ctx context.Context, email string, password string) (string, error) {
	var response struct {
		Token string `json:"token"`
	}
	if err := c.postJSON(ctx, "/auth/login", "", map[string]any{"email": email, "password": password}, &response); err != nil {
		return "", err
	}
	if response.Token == "" {
		return "", fmt.Errorf("cloud login response missing token")
	}
	return response.Token, nil
}

func (c *HTTPCloudClient) Register(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error) {
	payload := map[string]any{"name": edgeName, "user_token": token}
	for key, value := range metadata {
		payload[key] = value
	}
	var response domain.ProvisionedEdge
	if err := c.postJSON(ctx, "/edges/register", token, payload, &response); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	return validateRegistration(response)
}

func (c *HTTPCloudClient) Pair(ctx context.Context, token string, edgeName string, metadata map[string]any) (domain.ProvisionedEdge, error) {
	var response domain.ProvisionedEdge
	if err := c.postJSON(ctx, "/edges/pair", "", map[string]any{"token": token, "edge_name": edgeName, "name": edgeName, "metadata": metadata}, &response); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	return validateRegistration(response)
}

func (c *HTTPCloudClient) Unpair(ctx context.Context, edgeID string) error {
	return c.postJSON(ctx, "/edges/unpair", "", map[string]any{"edgeId": edgeID}, nil)
}

func (c *HTTPCloudClient) postJSON(ctx context.Context, path string, bearer string, payload any, target any) error {
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
	res, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("cloud request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var body map[string]any
		_ = json.NewDecoder(res.Body).Decode(&body)
		if message, ok := body["message"].(string); ok && message != "" {
			return fmt.Errorf("cloud request failed (%d): %s", res.StatusCode, message)
		}
		return fmt.Errorf("cloud request failed (%d): %s", res.StatusCode, res.Status)
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
