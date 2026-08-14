package jsonfileprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

type Adapter struct {
	basePath  string
	resource  map[string]any
	operation map[string]any
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
	basePath := strings.TrimSpace(stringValue(credentials, "basePath", "base_path"))
	if basePath == "" {
		basePath = "./data/exports"
	}

	return &Adapter{
		basePath:  basePath,
		resource:  mapValue(config.Resource, "config"),
		operation: mapValue(config.Operation, "config"),
	}, nil
}

func (a *Adapter) Send(ctx context.Context, payload map[string]any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	targetFile, err := a.targetFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
		return err
	}

	format := strings.ToLower(defaultString(stringValue(a.operation, "format"), "ndjson"))
	switch format {
	case "json", "json_array":
		return a.writeJSONArray(targetFile, payload)
	default:
		return a.writeNDJSON(targetFile, payload)
	}
}

func (a *Adapter) Test(ctx context.Context, payload map[string]any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	targetFile, err := a.targetFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(targetFile, os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func (a *Adapter) Discover(ctx context.Context) ([]providers.Resource, error) {
	return []providers.Resource{
		{
			Name:   "Arquivo JSON",
			Type:   "file",
			Config: a.resource,
			Fields: []providers.Field{
				{Name: "fileName", Type: "string"},
			},
		},
	}, ctx.Err()
}

func (a *Adapter) writeNDJSON(targetFile string, payload map[string]any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	flags := os.O_CREATE | os.O_WRONLY
	if strings.EqualFold(stringValue(a.operation, "mode"), "overwrite") {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(targetFile, flags, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return nil
}

func (a *Adapter) writeJSONArray(targetFile string, payload map[string]any) error {
	mode := strings.ToLower(defaultString(stringValue(a.operation, "mode"), "append"))
	items := make([]map[string]any, 0, 1)

	if mode == "append" {
		if data, err := os.ReadFile(targetFile); err == nil && len(data) > 0 {
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("invalid existing json array file: %w", err)
			}
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	items = append(items, payload)
	encoded, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(targetFile, encoded, 0o644)
}

func (a *Adapter) targetFile() (string, error) {
	fileName := strings.TrimSpace(stringValue(a.resource, "fileName", "file_name", "path"))
	if fileName == "" {
		return "", errors.New("json file provider requires resource.config.fileName")
	}

	cleaned := filepath.Clean(fileName)
	if filepath.IsAbs(cleaned) {
		return "", errors.New("json file provider does not accept absolute file paths in resources")
	}

	return filepath.Join(a.basePath, cleaned), nil
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

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
