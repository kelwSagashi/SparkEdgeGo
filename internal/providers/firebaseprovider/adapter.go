package firebaseprovider

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
	"github.com/kelwSagashi/sparkedge-go/internal/providers/googleauth"
)

const firestoreScope = "https://www.googleapis.com/auth/datastore"

type Adapter struct {
	projectID string
	auth      googleauth.ServiceAccount
	resource  map[string]any
	operation map[string]any
	client    *http.Client
	baseURL   string
	tokenURL  string
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
	auth, err := googleauth.FromFirebaseFields(credentials)
	if err != nil {
		return nil, err
	}
	return &Adapter{
		projectID: auth.ProjectID,
		auth:      auth,
		resource:  mapValue(config.Resource, "config"),
		operation: mapValue(config.Operation, "config"),
		client:    &http.Client{Timeout: 30 * time.Second},
		baseURL:   "https://firestore.googleapis.com/v1",
		tokenURL:  googleauth.DefaultTokenURL,
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	collection := stringValue(a.resource, "collection")
	if collection == "" {
		return errors.New("firebase provider requires resource.config.collection")
	}
	op := stringValue(a.operation, "operation")
	if op == "" {
		op = "add"
	}
	docID := stringValue(a.operation, "docId", "doc_id")
	if docID == "" {
		docID = stringValue(payload, "id")
	}

	endpoint := fmt.Sprintf("%s/projects/%s/databases/(default)/documents/%s", a.baseURL, url.PathEscape(a.projectID), url.PathEscape(collection))
	method := http.MethodPost
	if op == "set" && docID != "" {
		method = http.MethodPatch
		endpoint += "/" + url.PathEscape(docID)
	}
	body := map[string]any{"fields": firestoreFields(payload)}
	return a.request(ctx, method, endpoint, body)
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	endpoint := fmt.Sprintf("%s/projects/%s/databases/(default)/documents", a.baseURL, url.PathEscape(a.projectID))
	return a.request(ctx, http.MethodGet, endpoint, nil)
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{{Name: "Firestore collection", Type: "collection", Config: a.resource}}, ctx.Err()
}

func (a *Adapter) request(ctx context.Context, method string, endpoint string, body map[string]any) error {
	token, err := googleauth.Token(ctx, a.client, a.tokenURL, a.auth, []string{firestoreScope})
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("firebase error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return nil
}

func firestoreFields(payload map[string]any) map[string]any {
	fields := map[string]any{}
	for key, value := range payload {
		fields[key] = firestoreValue(value)
	}
	return fields
}

func firestoreValue(value any) map[string]any {
	switch typed := value.(type) {
	case bool:
		return map[string]any{"booleanValue": typed}
	case int:
		return map[string]any{"integerValue": fmt.Sprint(typed)}
	case int64:
		return map[string]any{"integerValue": fmt.Sprint(typed)}
	case float64:
		return map[string]any{"doubleValue": typed}
	case map[string]any:
		return map[string]any{"mapValue": map[string]any{"fields": firestoreFields(typed)}}
	case []any:
		values := make([]any, 0, len(typed))
		for _, item := range typed {
			values = append(values, firestoreValue(item))
		}
		return map[string]any{"arrayValue": map[string]any{"values": values}}
	case nil:
		return map[string]any{"nullValue": nil}
	default:
		return map[string]any{"stringValue": fmt.Sprint(typed)}
	}
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
