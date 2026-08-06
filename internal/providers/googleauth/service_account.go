package googleauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const DefaultTokenURL = "https://oauth2.googleapis.com/token"

type ServiceAccount struct {
	ProjectID    string `json:"project_id"`
	ClientEmail  string `json:"client_email"`
	PrivateKey   string `json:"private_key"`
	PrivateKeyID string `json:"private_key_id"`
}

func FromJSON(value string) (ServiceAccount, error) {
	var account ServiceAccount
	if err := json.Unmarshal([]byte(value), &account); err != nil {
		return ServiceAccount{}, err
	}
	account.PrivateKey = strings.ReplaceAll(account.PrivateKey, `\n`, "\n")
	if account.ClientEmail == "" || account.PrivateKey == "" {
		return ServiceAccount{}, errors.New("google service account email and private key are required")
	}
	return account, nil
}

func FromFirebaseFields(credentials map[string]any) (ServiceAccount, error) {
	account := ServiceAccount{
		ProjectID:   stringValue(credentials, "projectId", "project_id"),
		ClientEmail: stringValue(credentials, "clientEmail", "client_email"),
		PrivateKey:  strings.ReplaceAll(stringValue(credentials, "privateKey", "private_key"), `\n`, "\n"),
	}
	if account.ProjectID == "" || account.ClientEmail == "" || account.PrivateKey == "" {
		return ServiceAccount{}, errors.New("firebase service account projectId, clientEmail and privateKey are required")
	}
	return account, nil
}

func Token(ctx context.Context, client *http.Client, tokenURL string, account ServiceAccount, scopes []string) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if tokenURL == "" {
		tokenURL = DefaultTokenURL
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(account.PrivateKey))
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"iss":   account.ClientEmail,
		"scope": strings.Join(scopes, " "),
		"aud":   tokenURL,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if account.PrivateKeyID != "" {
		token.Header["kid"] = account.PrivateKeyID
	}
	assertion, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New(strings.TrimSpace(string(data)))
	}
	var decoded struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return "", err
	}
	if decoded.AccessToken == "" {
		return "", errors.New("google token response did not include access_token")
	}
	return decoded.AccessToken, nil
}

func stringValue(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := source[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
