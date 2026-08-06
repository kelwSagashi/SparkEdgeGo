package googleprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/googleauth"
)

var googleScopes = []string{
	"https://www.googleapis.com/auth/spreadsheets",
	"https://www.googleapis.com/auth/drive.file",
}

type SheetsAdapter struct {
	auth      googleauth.ServiceAccount
	resource  map[string]any
	operation map[string]any
	client    *http.Client
	tokenURL  string
	baseURL   string
}

type DriveAdapter struct {
	auth      googleauth.ServiceAccount
	resource  map[string]any
	operation map[string]any
	client    *http.Client
	tokenURL  string
	uploadURL string
	baseURL   string
}

func Register(registry *providers.Registry) {
	if registry == nil {
		return
	}
	registry.Register(SheetsStrategyID, func(config providers.Config) (providers.Adapter, error) {
		return NewSheets(config)
	})
	registry.Register(DriveStrategyID, func(config providers.Config) (providers.Adapter, error) {
		return NewDrive(config)
	})
}

func NewSheets(config providers.Config) (*SheetsAdapter, error) {
	account, err := accountFromConfig(config)
	if err != nil {
		return nil, err
	}
	return &SheetsAdapter{
		auth:      account,
		resource:  mapValue(config.Resource, "config"),
		operation: mapValue(config.Operation, "config"),
		client:    &http.Client{Timeout: 30 * time.Second},
		tokenURL:  googleauth.DefaultTokenURL,
		baseURL:   "https://sheets.googleapis.com/v4",
	}, nil
}

func NewDrive(config providers.Config) (*DriveAdapter, error) {
	account, err := accountFromConfig(config)
	if err != nil {
		return nil, err
	}
	return &DriveAdapter{
		auth:      account,
		resource:  mapValue(config.Resource, "config"),
		operation: mapValue(config.Operation, "config"),
		client:    &http.Client{Timeout: 30 * time.Second},
		tokenURL:  googleauth.DefaultTokenURL,
		uploadURL: "https://www.googleapis.com/upload/drive/v3",
		baseURL:   "https://www.googleapis.com/drive/v3",
	}, nil
}

func (a *SheetsAdapter) Send(ctx context.Context, payload map[string]any) error {
	spreadsheetID := stringValue(a.resource, "spreadsheetId", "spreadsheet_id")
	if spreadsheetID == "" {
		return errors.New("google sheets provider requires resource.config.spreadsheetId")
	}
	rangeName := stringValue(a.operation, "range")
	if rangeName == "" {
		rangeName = "Sheet1!A1"
	}
	action := stringValue(a.operation, "action")
	method := http.MethodPost
	suffix := ":append"
	if action == "update" {
		method = http.MethodPut
		suffix = ""
	}
	endpoint := fmt.Sprintf("%s/spreadsheets/%s/values/%s%s", a.baseURL, url.PathEscape(spreadsheetID), url.PathEscape(rangeName), suffix)
	query := url.Values{"valueInputOption": []string{"RAW"}}
	return a.request(ctx, method, endpoint+"?"+query.Encode(), map[string]any{"values": [][]any{objectValues(payload)}})
}

func (a *SheetsAdapter) Test(ctx context.Context, payload map[string]any) error {
	spreadsheetID := stringValue(a.resource, "spreadsheetId", "spreadsheet_id")
	if spreadsheetID == "" {
		return nil
	}
	endpoint := fmt.Sprintf("%s/spreadsheets/%s", a.baseURL, url.PathEscape(spreadsheetID))
	return a.request(ctx, http.MethodGet, endpoint, nil)
}

func (a *SheetsAdapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{{Name: "Google Spreadsheet", Type: "spreadsheet", Config: a.resource}}, ctx.Err()
}

func (a *SheetsAdapter) request(ctx context.Context, method string, endpoint string, body map[string]any) error {
	token, err := googleauth.Token(ctx, a.client, a.tokenURL, a.auth, googleScopes)
	if err != nil {
		return err
	}
	return jsonRequest(ctx, a.client, method, endpoint, token, body)
}

func (a *DriveAdapter) Send(ctx context.Context, payload map[string]any) error {
	token, err := googleauth.Token(ctx, a.client, a.tokenURL, a.auth, googleScopes)
	if err != nil {
		return err
	}
	fileName := stringValue(a.operation, "fileName", "file_name")
	if fileName == "" {
		fileName = fmt.Sprintf("export_%d.json", time.Now().UnixMilli())
	}
	return a.uploadJSON(ctx, token, stringValue(a.resource, "folderId", "folder_id"), fileName, payload)
}

func (a *DriveAdapter) Test(ctx context.Context, payload map[string]any) error {
	token, err := googleauth.Token(ctx, a.client, a.tokenURL, a.auth, googleScopes)
	if err != nil {
		return err
	}
	return jsonRequest(ctx, a.client, http.MethodGet, a.baseURL+"/files?pageSize=1", token, nil)
}

func (a *DriveAdapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{{Name: "Google Drive folder", Type: "folder", Config: a.resource}}, ctx.Err()
}

func (a *DriveAdapter) uploadJSON(ctx context.Context, token string, folderID string, fileName string, payload map[string]any) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadata := map[string]any{"name": fileName}
	if folderID != "" {
		metadata["parents"] = []string{folderID}
	}
	if err := writeJSONPart(writer, "metadata", "application/json; charset=UTF-8", metadata); err != nil {
		return err
	}
	if err := writeJSONPart(writer, "media", "application/json", payload); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	endpoint := a.uploadURL + "/files?uploadType=multipart"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("google drive error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return nil
}

func accountFromConfig(config providers.Config) (googleauth.ServiceAccount, error) {
	credentials := mapValue(config.Credentials, "data")
	serviceAccountJSON := stringValue(credentials, "serviceAccountJson", "service_account_json")
	if serviceAccountJSON == "" {
		return googleauth.ServiceAccount{}, errors.New("google provider requires credentials.data.serviceAccountJson")
	}
	return googleauth.FromJSON(serviceAccountJSON)
}

func jsonRequest(ctx context.Context, client *http.Client, method string, endpoint string, token string, body map[string]any) error {
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
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("google api error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return nil
}

func writeJSONPart(writer *multipart.Writer, name string, contentType string, value any) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, name))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	return json.NewEncoder(part).Encode(value)
}

func objectValues(payload map[string]any) []any {
	values := make([]any, 0, len(payload))
	for _, value := range payload {
		values = append(values, value)
	}
	return values
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
