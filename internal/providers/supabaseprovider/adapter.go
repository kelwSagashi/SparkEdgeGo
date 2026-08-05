package supabaseprovider

import (
	"bytes"
	"context"
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

type Adapter struct {
	client      *http.Client
	baseURL     string
	apiKey      string
	headers     map[string]string
	resource    map[string]any
	operation   map[string]any
	credentials map[string]any
}

func Register(registry *providers.Registry) {
	if registry == nil {
		return
	}
	registry.Register(StrategyID, func(config providers.Config) (providers.Adapter, error) {
		return New(config)
	})
}

func New(config providers.Config) (*Adapter, error) {
	credentials := mapValue(config.Credentials, "data")
	baseURL := stringValue(credentials, "url")
	apiKey := stringValue(credentials, "apiKey", "api_key")
	if baseURL == "" || apiKey == "" {
		return nil, errors.New("supabase provider requires credentials.data.url and credentials.data.apiKey")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid supabase url: %w", err)
	}
	return &Adapter{
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      apiKey,
		headers:     stringMap(mapValue(config.Server, "headers")),
		resource:    mapValue(config.Resource, "config"),
		operation:   mapValue(config.Operation, "config"),
		credentials: credentials,
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	_, err := a.do(ctx, payload)
	return err
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	_, err := a.schema(ctx)
	return err
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	schema, err := a.schema(ctx)
	if err != nil {
		return nil, err
	}
	definitions, _ := schema["definitions"].(map[string]any)
	result := make([]providers.Resource, 0, len(definitions))
	for tableName, rawDefinition := range definitions {
		resource := providers.Resource{
			Name:   tableName,
			Type:   "table",
			Config: map[string]any{"table": tableName},
		}
		if definition, ok := rawDefinition.(map[string]any); ok {
			if properties, ok := definition["properties"].(map[string]any); ok {
				for columnName, rawColumn := range properties {
					fieldType := "string"
					if column, ok := rawColumn.(map[string]any); ok {
						fieldType = stringValue(column, "type")
						if fieldType == "" {
							fieldType = "string"
						}
					}
					resource.Fields = append(resource.Fields, providers.Field{Name: columnName, Type: fieldType})
				}
			}
		}
		result = append(result, resource)
	}
	return result, nil
}

func (a *Adapter) do(ctx context.Context, payload map[string]any) (map[string]any, error) {
	table := stringValue(a.resource, "table")
	if table == "" {
		return nil, errors.New("supabase provider requires resource.config.table")
	}
	methodName := stringValue(a.operation, "method")
	if methodName == "" {
		methodName = "insert"
	}
	method := http.MethodPost
	body := payload
	queryParams := map[string]string{}
	extraHeaders := map[string]string{}

	switch methodName {
	case "insert":
		method = http.MethodPost
	case "update":
		method = http.MethodPatch
	case "upsert":
		method = http.MethodPost
		extraHeaders["Prefer"] = "resolution=merge-duplicates"
	case "select":
		method = http.MethodGet
		body = nil
		for key, value := range payload {
			queryParams[key] = "eq." + fmt.Sprint(value)
		}
	default:
		return nil, fmt.Errorf("unsupported supabase method %q", methodName)
	}
	return a.request(ctx, table, method, body, queryParams, extraHeaders)
}

func (a *Adapter) schema(ctx context.Context) (map[string]any, error) {
	return a.request(ctx, "", http.MethodGet, nil, nil, nil)
}

func (a *Adapter) request(ctx context.Context, table string, method string, body map[string]any, params map[string]string, extraHeaders map[string]string) (map[string]any, error) {
	endpoint, err := url.Parse(a.baseURL + "/rest/v1/")
	if err != nil {
		return nil, err
	}
	if table != "" {
		endpoint = endpoint.JoinPath(table)
	}
	query := endpoint.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	endpoint.RawQuery = query.Encode()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", a.apiKey)
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Prefer", "return=minimal")
	for key, value := range a.headers {
		req.Header.Set(key, value)
	}
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

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
		return nil, fmt.Errorf("supabase error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
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
