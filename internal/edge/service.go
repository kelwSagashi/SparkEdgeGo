package edge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var (
	ErrInvalidOnboarding = errors.New("invalid edge onboarding")
	ErrNotProvisioned    = errors.New("edge not provisioned")
)

type Repository interface {
	GetIdentity(context.Context) (domain.EdgeIdentity, error)
	UpsertIdentity(context.Context, domain.EdgeIdentity) (domain.EdgeIdentity, error)
	ClearIdentity(context.Context) error
	GetMqttCredentials(context.Context) (domain.EdgeCredentials, error)
	UpsertMqttCredentials(context.Context, domain.EdgeCredentials) (domain.EdgeCredentials, error)
	ClearMqttCredentials(context.Context) error
	GetEdgeConfig(context.Context) (domain.EdgeConfig, error)
	UpsertEdgeConfig(context.Context, sqlite.UpsertEdgeConfigParams) (domain.EdgeConfig, error)
	ClearEdgeConfig(context.Context) error
}

type Service struct {
	repo  Repository
	cloud CloudClient
	mqtt  *mqtt.Client
}

type OnboardingRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Lat         string   `json:"lat"`
	Lng         string   `json:"lng"`
	Tags        []string `json:"tags"`
}

type PairRequest struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type ConnectRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewService(repo Repository, cloud CloudClient, mqttClient *mqtt.Client) *Service {
	return &Service{repo: repo, cloud: cloud, mqtt: mqttClient}
}

func (s *Service) GetOnboarding(ctx context.Context) (domain.EdgeConfig, bool, error) {
	config, err := s.repo.GetEdgeConfig(ctx)
	if errors.Is(err, sqlite.ErrNotFound) {
		return domain.EdgeConfig{}, false, nil
	}
	if err != nil {
		return domain.EdgeConfig{}, false, err
	}
	return config, strings.TrimSpace(config.EdgeName) != "", nil
}

func (s *Service) SaveOnboarding(ctx context.Context, req OnboardingRequest) (domain.EdgeConfig, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.EdgeConfig{}, ErrInvalidOnboarding
	}
	locationSource := ""
	if strings.TrimSpace(req.Lat) != "" && strings.TrimSpace(req.Lng) != "" {
		locationSource = "manual"
	}
	return s.repo.UpsertEdgeConfig(ctx, sqlite.UpsertEdgeConfigParams{
		EdgeName:       &name,
		Description:    nullablePtr(req.Description),
		Lat:            nullablePtr(req.Lat),
		Lng:            nullablePtr(req.Lng),
		Tags:           &req.Tags,
		LocationSource: &locationSource,
	})
}

func (s *Service) UpsertConfigMap(ctx context.Context, values map[string]any) (domain.EdgeConfig, error) {
	params := sqlite.UpsertEdgeConfigParams{}
	if value, ok := stringValue(values, "edge_name", "name"); ok {
		params.EdgeName = &value
	}
	if value, ok := stringValue(values, "description"); ok {
		params.Description = &value
	}
	if value, ok := stringValue(values, "lat"); ok {
		params.Lat = &value
	}
	if value, ok := stringValue(values, "lng"); ok {
		params.Lng = &value
	}
	if value, ok := stringValue(values, "location_source"); ok {
		params.LocationSource = &value
	}
	if value, ok := stringValue(values, "os"); ok {
		params.OS = &value
	}
	if value, ok := stringValue(values, "os_version"); ok {
		params.OSVersion = &value
	}
	if value, ok := stringValue(values, "edge_version"); ok {
		params.EdgeVersion = &value
	}
	if value, ok := stringValue(values, "hardware"); ok {
		params.Hardware = &value
	}
	if value, ok := stringValue(values, "environment"); ok {
		params.Environment = &value
	}
	if tags, ok := tagsValue(values["tags"]); ok {
		params.Tags = &tags
	}
	return s.repo.UpsertEdgeConfig(ctx, params)
}

func (s *Service) Load(ctx context.Context) (domain.ProvisionedEdge, error) {
	identity, err := s.repo.GetIdentity(ctx)
	if errors.Is(err, sqlite.ErrNotFound) || !identity.Provisioned || identity.EdgeID == "" {
		return domain.ProvisionedEdge{}, ErrNotProvisioned
	}
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	credentials, err := s.repo.GetMqttCredentials(ctx)
	if errors.Is(err, sqlite.ErrNotFound) || credentials.BrokerURL == "" || credentials.Username == "" || credentials.Password == "" {
		return domain.ProvisionedEdge{}, ErrNotProvisioned
	}
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	return domain.ProvisionedEdge{
		EdgeID:   identity.EdgeID,
		EdgeName: identity.EdgeName,
		MQTT: domain.EdgeMQTTConfig{
			URL:      envOrDefault("MQTT_URL", credentials.BrokerURL),
			Username: credentials.Username,
			Password: credentials.Password,
		},
		Provisioned: true,
	}, nil
}

func (s *Service) IsProvisioned(ctx context.Context) bool {
	_, err := s.Load(ctx)
	return err == nil
}

func (s *Service) SaveProvisioned(ctx context.Context, edge domain.ProvisionedEdge) error {
	if _, err := s.repo.UpsertIdentity(ctx, domain.EdgeIdentity{EdgeID: edge.EdgeID, EdgeName: edge.EdgeName, Provisioned: edge.Provisioned}); err != nil {
		return err
	}
	_, err := s.repo.UpsertMqttCredentials(ctx, domain.EdgeCredentials{Type: "mqtt", BrokerURL: edge.MQTT.URL, Username: edge.MQTT.Username, Password: edge.MQTT.Password})
	return err
}

func (s *Service) Status(ctx context.Context) (map[string]any, error) {
	edge, err := s.Load(ctx)
	if errors.Is(err, ErrNotProvisioned) {
		return map[string]any{"connected": false, "edge_id": nil, "edge_name": nil, "mqtt": map[string]any{"connected": s.mqtt != nil && s.mqtt.IsConnected()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"connected": edge.Provisioned, "edge_id": edge.EdgeID, "edge_name": edge.EdgeName, "mqtt": map[string]any{"connected": s.mqtt != nil && s.mqtt.IsConnected()}}, nil
}

func (s *Service) Pair(ctx context.Context, req PairRequest) (domain.ProvisionedEdge, error) {
	if strings.TrimSpace(req.Token) == "" {
		return domain.ProvisionedEdge{}, ErrInvalidOnboarding
	}
	config, _, _ := s.GetOnboarding(ctx)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = config.EdgeName
	}
	registration, err := s.cloud.Pair(ctx, req.Token, name, metadataFromConfig(config))
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	if err := s.SaveProvisioned(ctx, registration); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	_ = s.Reconnect(ctx)
	return registration, nil
}

func (s *Service) Connect(ctx context.Context, req ConnectRequest) (domain.ProvisionedEdge, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return domain.ProvisionedEdge{}, ErrInvalidOnboarding
	}
	config, complete, err := s.GetOnboarding(ctx)
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	if !complete {
		return domain.ProvisionedEdge{}, ErrInvalidOnboarding
	}
	token, err := s.cloud.Login(ctx, req.Email, req.Password)
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	registration, err := s.cloud.Register(ctx, token, config.EdgeName, metadataFromConfig(config))
	if err != nil {
		return domain.ProvisionedEdge{}, err
	}
	if err := s.SaveProvisioned(ctx, registration); err != nil {
		return domain.ProvisionedEdge{}, err
	}
	_ = s.Reconnect(ctx)
	return registration, nil
}

func (s *Service) Disconnect(ctx context.Context) error {
	if s.mqtt != nil && s.mqtt.IsConnected() {
		return s.mqtt.Disconnect(ctx)
	}
	return nil
}

func (s *Service) Reconnect(ctx context.Context) error {
	edge, err := s.Load(ctx)
	if err != nil {
		return err
	}
	if s.mqtt == nil {
		return nil
	}
	_ = s.mqtt.Disconnect(ctx)
	if err := s.mqtt.Connect(ctx, mqtt.Config{EdgeID: edge.EdgeID, BrokerURL: edge.MQTT.URL, Username: edge.MQTT.Username, Password: edge.MQTT.Password}); err != nil {
		return err
	}
	config, _, _ := s.GetOnboarding(ctx)
	_ = s.mqtt.PublishMeta(ctx, metadataEnvelope(edge.EdgeID, edge.EdgeName, config))
	_ = s.mqtt.PublishStats(ctx)
	s.mqtt.StartHeartbeat(0)
	s.mqtt.StartStats(0)
	_ = s.mqtt.RetryQueue(ctx, 5)
	return nil
}

func (s *Service) Remove(ctx context.Context) error {
	edge, err := s.Load(ctx)
	if err != nil && !errors.Is(err, ErrNotProvisioned) {
		return err
	}
	if err == nil && s.cloud != nil {
		_ = s.cloud.Unpair(ctx, edge.EdgeID)
	}
	_ = s.Disconnect(ctx)
	if err := s.repo.ClearIdentity(ctx); err != nil {
		return err
	}
	if err := s.repo.ClearMqttCredentials(ctx); err != nil {
		return err
	}
	return s.repo.ClearEdgeConfig(ctx)
}

func metadataFromConfig(config domain.EdgeConfig) map[string]any {
	metadata := map[string]any{
		"lat":          emptyAsNil(config.Lat),
		"lng":          emptyAsNil(config.Lng),
		"tags":         config.Tags,
		"description":  emptyAsNil(config.Description),
		"os":           envOrDefault("SPARKEDGE_OS", runtime.GOOS),
		"os_version":   envOrDefault("SPARKEDGE_OS_VERSION", runtime.GOOS+"/"+runtime.GOARCH),
		"edge_version": envOrDefault("SPARKEDGE_VERSION", "go-dev"),
		"hardware":     envOrDefault("SPARKEDGE_HARDWARE", runtime.GOARCH),
		"environment":  envOrDefault("SPARKEDGE_ENV", "production"),
	}
	return metadata
}

func metadataEnvelope(edgeID string, edgeName string, config domain.EdgeConfig) map[string]any {
	metadata := metadataFromConfig(config)
	metadata["edge_id"] = edgeID
	metadata["edge_name"] = edgeName
	metadata["location_source"] = emptyAsNil(config.LocationSource)
	return metadata
}

func nullablePtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func emptyAsNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func stringValue(values map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if raw, ok := values[key]; ok {
			switch value := raw.(type) {
			case string:
				return strings.TrimSpace(value), true
			case nil:
				return "", true
			default:
				return strings.TrimSpace(fmt.Sprint(value)), true
			}
		}
	}
	return "", false
}

func tagsValue(raw any) ([]string, bool) {
	switch value := raw.(type) {
	case []string:
		return value, true
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(text))
			}
		}
		return result, true
	default:
		return nil, false
	}
}
