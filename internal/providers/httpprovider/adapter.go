package httpprovider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

const (
	StrategyNoAuth = "no_auth"
	StrategyAPIKey = "api_key"
	StrategyBasic  = "basic_auth"
	StrategyBearer = "bearer_token"
)

type Adapter struct {
	client      *http.Client
	strategy    string
	baseURL     string
	headers     map[string]string
	credentials map[string]any
	operation   map[string]any
	resource    map[string]any
}

func Register(registry *providers.Registry) {
	if registry == nil {
		return
	}
	for _, strategy := range []string{StrategyNoAuth, StrategyAPIKey, StrategyBasic, StrategyBearer, "http"} {
		strategy := strategy
		registry.Register(strategy, func(config providers.Config) (providers.Adapter, error) {
			if strategy == "http" {
				strategy = inferStrategy(config)
			}
			return New(strategy, config)
		})
	}
}

func New(strategy string, config providers.Config) (*Adapter, error) {
	resourceConfig := mapValue(config.Resource, "config")
	baseURL := stringValue(resourceConfig, "baseUrl", "base_url")
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("http provider requires resource.config.baseUrl")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid http baseUrl: %w", err)
	}

	return &Adapter{
		client:      &http.Client{Timeout: 30 * time.Second},
		strategy:    strategy,
		baseURL:     baseURL,
		headers:     stringMap(mapValue(config.Server, "headers")),
		credentials: mapValue(config.Credentials, "data"),
		operation:   mapValue(config.Operation, "config"),
		resource:    resourceConfig,
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	_, err := a.do(ctx, payload, "POST")
	return err
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	_, err := a.do(ctx, payload, "GET")
	return err
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{
		{
			Name:   "HTTP endpoint",
			Type:   "http",
			Config: a.resource,
			Fields: []providers.Field{
				{Name: "baseUrl", Type: "text"},
			},
		},
	}, ctx.Err()
}

func (a *Adapter) do(ctx context.Context, payload map[string]any, fallbackMethod string) (map[string]any, error) {
	method := strings.ToUpper(stringValue(a.operation, "method"))
	if method == "" {
		method = fallbackMethod
	}
	path := stringValue(a.operation, "path")
	if path == "" {
		path = "/"
	}
	endpoint, err := url.Parse(a.baseURL)
	if err != nil {
		return nil, err
	}
	relative, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	endpoint = endpoint.ResolveReference(relative)

	var body io.Reader
	if method != http.MethodGet && payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range a.headers {
		req.Header.Set(key, value)
	}
	a.applyAuth(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return map[string]any{}, nil
	}
	return decoded, nil
}

func (a *Adapter) applyAuth(req *http.Request) {
	switch a.strategy {
	case StrategyBearer:
		if token := stringValue(a.credentials, "token"); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case StrategyBasic:
		username := stringValue(a.credentials, "username")
		password := stringValue(a.credentials, "password")
		if username != "" || password != "" {
			value := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			req.Header.Set("Authorization", "Basic "+value)
		}
	case StrategyAPIKey:
		key := stringValue(a.credentials, "key")
		value := stringValue(a.credentials, "value")
		location := stringValue(a.credentials, "in")
		if key == "" {
			return
		}
		if location == "query" {
			query := req.URL.Query()
			query.Set(key, value)
			req.URL.RawQuery = query.Encode()
			return
		}
		req.Header.Set(key, value)
	}
}

func inferStrategy(config providers.Config) string {
	if authType := stringValue(config.Credentials, "auth_type_id"); authType != "" {
		return authType
	}
	return StrategyNoAuth
}

func mapValue(source map[string]any, key string) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	if value, ok := source[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func stringMap(source map[string]any) map[string]string {
	result := map[string]string{}
	for key, value := range source {
		if text, ok := value.(string); ok {
			result[key] = text
		}
	}
	return result
}

func stringValue(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := source[key]; ok {
			switch typed := value.(type) {
			case string:
				return strings.TrimSpace(typed)
			case fmt.Stringer:
				return strings.TrimSpace(typed.String())
			}
		}
	}
	return ""
}
