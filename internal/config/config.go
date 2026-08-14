package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
	"github.com/kelwSagashi/sparkedge-go/internal/connectivity"
	"gopkg.in/yaml.v3"
)

const (
	defaultCloudURL  = "http://localhost:3000"
	defaultMQTTURL   = "mqtt://localhost:1883"
	defaultDBFile    = "sparkedge.db"
	defaultJWTSecret = "dev-secret"
	defaultHTTPPort  = 3009
	configFileName   = "config.yml"
)

type File struct {
	Cloud        CloudSection        `yaml:"cloud"`
	DB           DBSection           `yaml:"db"`
	Auth         AuthSection         `yaml:"auth"`
	Server       ServerSection       `yaml:"server"`
	Update       UpdateSection       `yaml:"update"`
	Connectivity ConnectivitySection `yaml:"connectivity"`
	Retention    RetentionSection    `yaml:"retention"`
}

type Effective struct {
	Cloud        CloudSection        `json:"cloud"`
	DB           DBSection           `json:"db"`
	Auth         AuthView            `json:"auth"`
	Server       ServerView          `json:"server"`
	Update       UpdateView          `json:"update"`
	Connectivity ConnectivitySection `json:"connectivity"`
	Retention    RetentionSection    `json:"retention"`
	ConfigFile   string              `json:"config_file,omitempty"`
}

type CloudSection struct {
	URL       string `json:"url" yaml:"url"`
	MQTTURL   string `json:"mqtt_url" yaml:"mqtt_url"`
	SyncToken string `json:"sync_token" yaml:"sync_token"`
}

type DBSection struct {
	File string `json:"file" yaml:"file"`
}

type AuthSection struct {
	JWTSecret string `json:"jwt_secret" yaml:"jwt_secret"`
}

type AuthView struct {
	JWTSecret string `json:"jwt_secret"`
	IsDefault bool   `json:"is_default"`
}

type ServerSection struct {
	Port int `json:"port" yaml:"port"`
}

type ServerView struct {
	Port string `json:"port"`
}

type UpdateSection struct {
	Enabled         *bool  `json:"enabled" yaml:"enabled"`
	Provider        string `json:"provider" yaml:"provider"`
	Repo            string `json:"repo" yaml:"repo"`
	Channel         string `json:"channel" yaml:"channel"`
	AllowPrerelease *bool  `json:"allow_prerelease" yaml:"allow_prerelease"`
	ServiceName     string `json:"service_name" yaml:"service_name"`
	RestartCommand  string `json:"restart_command" yaml:"restart_command"`
}

type UpdateView struct {
	Enabled         bool   `json:"enabled"`
	Provider        string `json:"provider"`
	Repo            string `json:"repo"`
	Channel         string `json:"channel"`
	AllowPrerelease bool   `json:"allow_prerelease"`
	ServiceName     string `json:"service_name"`
	RestartCommand  string `json:"restart_command"`
}

type ConnectivitySection = connectivity.Policy

type RetentionSection struct {
	MQTTQueueMaxItems                 int `json:"mqtt_queue_max_items" yaml:"mqtt_queue_max_items"`
	MQTTQueueMaxAgeHours              int `json:"mqtt_queue_max_age_hours" yaml:"mqtt_queue_max_age_hours"`
	CloudSyncSentRetentionHours       int `json:"cloud_sync_sent_retention_hours" yaml:"cloud_sync_sent_retention_hours"`
	CloudSyncFailedRetentionHours     int `json:"cloud_sync_failed_retention_hours" yaml:"cloud_sync_failed_retention_hours"`
	CloudSyncKeepSentItems            int `json:"cloud_sync_keep_sent_items" yaml:"cloud_sync_keep_sent_items"`
	CloudSyncKeepFailedItems          int `json:"cloud_sync_keep_failed_items" yaml:"cloud_sync_keep_failed_items"`
	LocalFallbackSentRetentionHours   int `json:"local_fallback_sent_retention_hours" yaml:"local_fallback_sent_retention_hours"`
	LocalFallbackFailedRetentionHours int `json:"local_fallback_failed_retention_hours" yaml:"local_fallback_failed_retention_hours"`
	LocalFallbackKeepSentItems        int `json:"local_fallback_keep_sent_items" yaml:"local_fallback_keep_sent_items"`
	LocalFallbackKeepFailedItems      int `json:"local_fallback_keep_failed_items" yaml:"local_fallback_keep_failed_items"`
}

type Update struct {
	Cloud        *CloudSectionUpdate        `json:"cloud"`
	DB           *DBSectionUpdate           `json:"db"`
	Auth         *AuthSectionUpdate         `json:"auth"`
	Server       *ServerSectionUpdate       `json:"server"`
	Update       *UpdateSectionUpdate       `json:"update"`
	Connectivity *ConnectivitySectionUpdate `json:"connectivity"`
	Retention    *RetentionSectionUpdate    `json:"retention"`
}

type CloudSectionUpdate struct {
	URL       *string `json:"url"`
	MQTTURL   *string `json:"mqtt_url"`
	SyncToken *string `json:"sync_token"`
}

type DBSectionUpdate struct {
	File *string `json:"file"`
}

type AuthSectionUpdate struct {
	JWTSecret *string `json:"jwt_secret"`
}

type ServerSectionUpdate struct {
	Port any `json:"port"`
}

type UpdateSectionUpdate struct {
	Enabled         *bool   `json:"enabled"`
	Provider        *string `json:"provider"`
	Repo            *string `json:"repo"`
	Channel         *string `json:"channel"`
	AllowPrerelease *bool   `json:"allow_prerelease"`
	ServiceName     *string `json:"service_name"`
	RestartCommand  *string `json:"restart_command"`
}

type ConnectivitySectionUpdate struct {
	IntermittentPendingAgeSeconds   *int `json:"intermittent_pending_age_seconds"`
	DegradedPendingAgeSeconds       *int `json:"degraded_pending_age_seconds"`
	IntermittentCloudSyncQueueDepth *int `json:"intermittent_cloud_sync_queue_depth"`
	DegradedCloudSyncQueueDepth     *int `json:"degraded_cloud_sync_queue_depth"`
	DegradedMQTTQueueDepth          *int `json:"degraded_mqtt_queue_depth"`
	HeartbeatHealthySeconds         *int `json:"heartbeat_healthy_seconds"`
	HeartbeatDegradedSeconds        *int `json:"heartbeat_degraded_seconds"`
	StatsHealthySeconds             *int `json:"stats_healthy_seconds"`
	StatsDegradedSeconds            *int `json:"stats_degraded_seconds"`
}

type RetentionSectionUpdate struct {
	MQTTQueueMaxItems                 *int `json:"mqtt_queue_max_items"`
	MQTTQueueMaxAgeHours              *int `json:"mqtt_queue_max_age_hours"`
	CloudSyncSentRetentionHours       *int `json:"cloud_sync_sent_retention_hours"`
	CloudSyncFailedRetentionHours     *int `json:"cloud_sync_failed_retention_hours"`
	CloudSyncKeepSentItems            *int `json:"cloud_sync_keep_sent_items"`
	CloudSyncKeepFailedItems          *int `json:"cloud_sync_keep_failed_items"`
	LocalFallbackSentRetentionHours   *int `json:"local_fallback_sent_retention_hours"`
	LocalFallbackFailedRetentionHours *int `json:"local_fallback_failed_retention_hours"`
	LocalFallbackKeepSentItems        *int `json:"local_fallback_keep_sent_items"`
	LocalFallbackKeepFailedItems      *int `json:"local_fallback_keep_failed_items"`
}

type Runtime struct {
	CloudURL       string
	MQTTURL        string
	CloudSyncToken string
	DBFile         string
	JWTSecret      string
	HTTPPort       int
	ConfigFile     string
	Update         UpdateRuntime
	Connectivity   connectivity.Policy
	Retention      RetentionSection
}

type UpdateRuntime struct {
	Enabled         bool
	Provider        string
	Repo            string
	Channel         string
	AllowPrerelease bool
	ServiceName     string
	RestartCommand  string
}

type Manager struct {
	path string
}

func NewManager(path string) *Manager {
	path = strings.TrimSpace(path)
	if path == "" {
		path = appfs.ResolveFromRoot(configFileName)
	}
	return &Manager{path: path}
}

func (m *Manager) Path() string {
	return m.path
}

func (m *Manager) Load() (Effective, Runtime, error) {
	fileCfg, err := m.readFile()
	if err != nil {
		return Effective{}, Runtime{}, err
	}

	runtimeCfg := Runtime{
		CloudURL:       defaultCloudURL,
		MQTTURL:        defaultMQTTURL,
		CloudSyncToken: "",
		DBFile:         defaultDBFile,
		JWTSecret:      defaultJWTSecret,
		HTTPPort:       defaultHTTPPort,
		ConfigFile:     m.path,
		Update: UpdateRuntime{
			Enabled:         true,
			Provider:        "github",
			Repo:            "kelwSagashi/SparkEdgeGo",
			Channel:         "stable",
			AllowPrerelease: false,
		},
		Connectivity: connectivity.DefaultPolicy(),
		Retention:    defaultRetention(),
	}

	applyEnv(&runtimeCfg)
	applyFile(&runtimeCfg, fileCfg)

	effective := Effective{
		Cloud: CloudSection{
			URL:       runtimeCfg.CloudURL,
			MQTTURL:   runtimeCfg.MQTTURL,
			SyncToken: maskOptionalSecret(runtimeCfg.CloudSyncToken),
		},
		DB: DBSection{
			File: runtimeCfg.DBFile,
		},
		Auth: AuthView{
			JWTSecret: maskSecret(runtimeCfg.JWTSecret),
			IsDefault: runtimeCfg.JWTSecret == defaultJWTSecret,
		},
		Server: ServerView{
			Port: strconv.Itoa(runtimeCfg.HTTPPort),
		},
		Update: UpdateView{
			Enabled:         runtimeCfg.Update.Enabled,
			Provider:        runtimeCfg.Update.Provider,
			Repo:            runtimeCfg.Update.Repo,
			Channel:         runtimeCfg.Update.Channel,
			AllowPrerelease: runtimeCfg.Update.AllowPrerelease,
			ServiceName:     runtimeCfg.Update.ServiceName,
			RestartCommand:  runtimeCfg.Update.RestartCommand,
		},
		Connectivity: runtimeCfg.Connectivity.Normalize(),
		Retention:    normalizeRetention(runtimeCfg.Retention),
		ConfigFile:   m.path,
	}

	return effective, runtimeCfg, nil
}

func (m *Manager) Save(update Update) (Effective, error) {
	fileCfg, err := m.readFile()
	if err != nil {
		return Effective{}, err
	}
	if err := applyUpdate(&fileCfg, update); err != nil {
		return Effective{}, err
	}
	if err := m.writeFile(fileCfg); err != nil {
		return Effective{}, err
	}
	effective, _, err := m.Load()
	return effective, err
}

func (m *Manager) readFile() (File, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, err
	}
	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config file: %w", err)
	}
	return cfg, nil
}

func (m *Manager) writeFile(cfg File) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	payload, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config file: %w", err)
	}
	return os.WriteFile(m.path, payload, 0o644)
}

func applyEnv(runtimeCfg *Runtime) {
	if value := strings.TrimSpace(os.Getenv("SPARK_CLOUD_URL")); value != "" {
		runtimeCfg.CloudURL = value
	}
	if value := strings.TrimSpace(os.Getenv("MQTT_URL")); value != "" {
		runtimeCfg.MQTTURL = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_CLOUD_SYNC_TOKEN")); value != "" {
		runtimeCfg.CloudSyncToken = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_INTERMITTENT_PENDING_AGE_SECONDS"); ok {
		runtimeCfg.Connectivity.IntermittentPendingAgeSeconds = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_DEGRADED_PENDING_AGE_SECONDS"); ok {
		runtimeCfg.Connectivity.DegradedPendingAgeSeconds = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_INTERMITTENT_CLOUD_SYNC_QUEUE_DEPTH"); ok {
		runtimeCfg.Connectivity.IntermittentCloudSyncQueueDepth = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_DEGRADED_CLOUD_SYNC_QUEUE_DEPTH"); ok {
		runtimeCfg.Connectivity.DegradedCloudSyncQueueDepth = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_DEGRADED_MQTT_QUEUE_DEPTH"); ok {
		runtimeCfg.Connectivity.DegradedMQTTQueueDepth = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_HEARTBEAT_HEALTHY_SECONDS"); ok {
		runtimeCfg.Connectivity.HeartbeatHealthySeconds = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_HEARTBEAT_DEGRADED_SECONDS"); ok {
		runtimeCfg.Connectivity.HeartbeatDegradedSeconds = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_STATS_HEALTHY_SECONDS"); ok {
		runtimeCfg.Connectivity.StatsHealthySeconds = value
	}
	if value, ok := envInt("SPARKEDGE_CONNECTIVITY_STATS_DEGRADED_SECONDS"); ok {
		runtimeCfg.Connectivity.StatsDegradedSeconds = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_MQTT_QUEUE_MAX_ITEMS"); ok {
		runtimeCfg.Retention.MQTTQueueMaxItems = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_MQTT_QUEUE_MAX_AGE_HOURS"); ok {
		runtimeCfg.Retention.MQTTQueueMaxAgeHours = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_CLOUD_SYNC_SENT_RETENTION_HOURS"); ok {
		runtimeCfg.Retention.CloudSyncSentRetentionHours = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_CLOUD_SYNC_FAILED_RETENTION_HOURS"); ok {
		runtimeCfg.Retention.CloudSyncFailedRetentionHours = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_CLOUD_SYNC_KEEP_SENT_ITEMS"); ok {
		runtimeCfg.Retention.CloudSyncKeepSentItems = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_CLOUD_SYNC_KEEP_FAILED_ITEMS"); ok {
		runtimeCfg.Retention.CloudSyncKeepFailedItems = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_LOCAL_FALLBACK_SENT_RETENTION_HOURS"); ok {
		runtimeCfg.Retention.LocalFallbackSentRetentionHours = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_LOCAL_FALLBACK_FAILED_RETENTION_HOURS"); ok {
		runtimeCfg.Retention.LocalFallbackFailedRetentionHours = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_LOCAL_FALLBACK_KEEP_SENT_ITEMS"); ok {
		runtimeCfg.Retention.LocalFallbackKeepSentItems = value
	}
	if value, ok := envInt("SPARKEDGE_RETENTION_LOCAL_FALLBACK_KEEP_FAILED_ITEMS"); ok {
		runtimeCfg.Retention.LocalFallbackKeepFailedItems = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_DB_PATH")); value != "" {
		runtimeCfg.DBFile = value
	}
	if value := strings.TrimSpace(os.Getenv("JWT_SECRET")); value != "" {
		runtimeCfg.JWTSecret = value
	}
	if port := readPortFromEnv(); port > 0 {
		runtimeCfg.HTTPPort = port
	}
	if value, ok := envBool("SPARKEDGE_UPDATE_ENABLED"); ok {
		runtimeCfg.Update.Enabled = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_PROVIDER")); value != "" {
		runtimeCfg.Update.Provider = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_REPO")); value != "" {
		runtimeCfg.Update.Repo = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_CHANNEL")); value != "" {
		runtimeCfg.Update.Channel = value
	}
	if value, ok := envBool("SPARKEDGE_UPDATE_ALLOW_PRERELEASE"); ok {
		runtimeCfg.Update.AllowPrerelease = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_SERVICE_NAME")); value != "" {
		runtimeCfg.Update.ServiceName = value
	}
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_RESTART_COMMAND")); value != "" {
		runtimeCfg.Update.RestartCommand = value
	}
}

func applyFile(runtimeCfg *Runtime, fileCfg File) {
	if value := strings.TrimSpace(fileCfg.Cloud.URL); value != "" {
		runtimeCfg.CloudURL = value
	}
	if value := strings.TrimSpace(fileCfg.Cloud.MQTTURL); value != "" {
		runtimeCfg.MQTTURL = value
	}
	if value := strings.TrimSpace(fileCfg.Cloud.SyncToken); value != "" {
		runtimeCfg.CloudSyncToken = value
	}
	if value := strings.TrimSpace(fileCfg.DB.File); value != "" {
		runtimeCfg.DBFile = value
	}
	if value := strings.TrimSpace(fileCfg.Auth.JWTSecret); value != "" {
		runtimeCfg.JWTSecret = value
	}
	if fileCfg.Server.Port > 0 {
		runtimeCfg.HTTPPort = fileCfg.Server.Port
	}
	if fileCfg.Update.Enabled != nil {
		runtimeCfg.Update.Enabled = *fileCfg.Update.Enabled
	}
	if value := strings.TrimSpace(fileCfg.Update.Provider); value != "" {
		runtimeCfg.Update.Provider = value
	}
	if value := strings.TrimSpace(fileCfg.Update.Repo); value != "" {
		runtimeCfg.Update.Repo = value
	}
	if value := strings.TrimSpace(fileCfg.Update.Channel); value != "" {
		runtimeCfg.Update.Channel = value
	}
	if fileCfg.Update.AllowPrerelease != nil {
		runtimeCfg.Update.AllowPrerelease = *fileCfg.Update.AllowPrerelease
	}
	if value := strings.TrimSpace(fileCfg.Update.ServiceName); value != "" {
		runtimeCfg.Update.ServiceName = value
	}
	if value := strings.TrimSpace(fileCfg.Update.RestartCommand); value != "" {
		runtimeCfg.Update.RestartCommand = value
	}
	if fileCfg.Connectivity.IntermittentPendingAgeSeconds > 0 {
		runtimeCfg.Connectivity.IntermittentPendingAgeSeconds = fileCfg.Connectivity.IntermittentPendingAgeSeconds
	}
	if fileCfg.Connectivity.DegradedPendingAgeSeconds > 0 {
		runtimeCfg.Connectivity.DegradedPendingAgeSeconds = fileCfg.Connectivity.DegradedPendingAgeSeconds
	}
	if fileCfg.Connectivity.IntermittentCloudSyncQueueDepth > 0 {
		runtimeCfg.Connectivity.IntermittentCloudSyncQueueDepth = fileCfg.Connectivity.IntermittentCloudSyncQueueDepth
	}
	if fileCfg.Connectivity.DegradedCloudSyncQueueDepth > 0 {
		runtimeCfg.Connectivity.DegradedCloudSyncQueueDepth = fileCfg.Connectivity.DegradedCloudSyncQueueDepth
	}
	if fileCfg.Connectivity.DegradedMQTTQueueDepth > 0 {
		runtimeCfg.Connectivity.DegradedMQTTQueueDepth = fileCfg.Connectivity.DegradedMQTTQueueDepth
	}
	if fileCfg.Connectivity.HeartbeatHealthySeconds > 0 {
		runtimeCfg.Connectivity.HeartbeatHealthySeconds = fileCfg.Connectivity.HeartbeatHealthySeconds
	}
	if fileCfg.Connectivity.HeartbeatDegradedSeconds > 0 {
		runtimeCfg.Connectivity.HeartbeatDegradedSeconds = fileCfg.Connectivity.HeartbeatDegradedSeconds
	}
	if fileCfg.Connectivity.StatsHealthySeconds > 0 {
		runtimeCfg.Connectivity.StatsHealthySeconds = fileCfg.Connectivity.StatsHealthySeconds
	}
	if fileCfg.Connectivity.StatsDegradedSeconds > 0 {
		runtimeCfg.Connectivity.StatsDegradedSeconds = fileCfg.Connectivity.StatsDegradedSeconds
	}
	runtimeCfg.Connectivity = runtimeCfg.Connectivity.Normalize()
	if fileCfg.Retention.MQTTQueueMaxItems > 0 {
		runtimeCfg.Retention.MQTTQueueMaxItems = fileCfg.Retention.MQTTQueueMaxItems
	}
	if fileCfg.Retention.MQTTQueueMaxAgeHours > 0 {
		runtimeCfg.Retention.MQTTQueueMaxAgeHours = fileCfg.Retention.MQTTQueueMaxAgeHours
	}
	if fileCfg.Retention.CloudSyncSentRetentionHours > 0 {
		runtimeCfg.Retention.CloudSyncSentRetentionHours = fileCfg.Retention.CloudSyncSentRetentionHours
	}
	if fileCfg.Retention.CloudSyncFailedRetentionHours > 0 {
		runtimeCfg.Retention.CloudSyncFailedRetentionHours = fileCfg.Retention.CloudSyncFailedRetentionHours
	}
	if fileCfg.Retention.CloudSyncKeepSentItems > 0 {
		runtimeCfg.Retention.CloudSyncKeepSentItems = fileCfg.Retention.CloudSyncKeepSentItems
	}
	if fileCfg.Retention.CloudSyncKeepFailedItems > 0 {
		runtimeCfg.Retention.CloudSyncKeepFailedItems = fileCfg.Retention.CloudSyncKeepFailedItems
	}
	if fileCfg.Retention.LocalFallbackSentRetentionHours > 0 {
		runtimeCfg.Retention.LocalFallbackSentRetentionHours = fileCfg.Retention.LocalFallbackSentRetentionHours
	}
	if fileCfg.Retention.LocalFallbackFailedRetentionHours > 0 {
		runtimeCfg.Retention.LocalFallbackFailedRetentionHours = fileCfg.Retention.LocalFallbackFailedRetentionHours
	}
	if fileCfg.Retention.LocalFallbackKeepSentItems > 0 {
		runtimeCfg.Retention.LocalFallbackKeepSentItems = fileCfg.Retention.LocalFallbackKeepSentItems
	}
	if fileCfg.Retention.LocalFallbackKeepFailedItems > 0 {
		runtimeCfg.Retention.LocalFallbackKeepFailedItems = fileCfg.Retention.LocalFallbackKeepFailedItems
	}
	runtimeCfg.Retention = normalizeRetention(runtimeCfg.Retention)
}

func applyUpdate(fileCfg *File, update Update) error {
	if update.Cloud != nil {
		if update.Cloud.URL != nil {
			value := strings.TrimSpace(*update.Cloud.URL)
			if value == "" {
				return errors.New("cloud.url nao pode ser vazio")
			}
			fileCfg.Cloud.URL = value
		}
		if update.Cloud.MQTTURL != nil {
			value := strings.TrimSpace(*update.Cloud.MQTTURL)
			if value == "" {
				return errors.New("cloud.mqtt_url nao pode ser vazio")
			}
			fileCfg.Cloud.MQTTURL = value
		}
		if update.Cloud.SyncToken != nil {
			fileCfg.Cloud.SyncToken = strings.TrimSpace(*update.Cloud.SyncToken)
		}
	}
	if update.DB != nil && update.DB.File != nil {
		value := strings.TrimSpace(*update.DB.File)
		if value == "" {
			return errors.New("db.file nao pode ser vazio")
		}
		fileCfg.DB.File = value
	}
	if update.Auth != nil && update.Auth.JWTSecret != nil {
		value := strings.TrimSpace(*update.Auth.JWTSecret)
		if value == "" {
			return errors.New("auth.jwt_secret nao pode ser vazio")
		}
		if len(value) < 8 {
			return errors.New("jwt_secret deve ter pelo menos 8 caracteres")
		}
		fileCfg.Auth.JWTSecret = value
	}
	if update.Server != nil && update.Server.Port != nil {
		port, err := normalizePort(update.Server.Port)
		if err != nil {
			return err
		}
		fileCfg.Server.Port = port
	}
	if update.Update != nil {
		if update.Update.Enabled != nil {
			fileCfg.Update.Enabled = update.Update.Enabled
		}
		if update.Update.Provider != nil {
			value := strings.TrimSpace(*update.Update.Provider)
			if value == "" {
				return errors.New("update.provider nao pode ser vazio")
			}
			fileCfg.Update.Provider = value
		}
		if update.Update.Repo != nil {
			value := strings.TrimSpace(*update.Update.Repo)
			if value == "" {
				return errors.New("update.repo nao pode ser vazio")
			}
			fileCfg.Update.Repo = value
		}
		if update.Update.Channel != nil {
			value := normalizeUpdateChannel(*update.Update.Channel)
			fileCfg.Update.Channel = value
		}
		if update.Update.AllowPrerelease != nil {
			fileCfg.Update.AllowPrerelease = update.Update.AllowPrerelease
		}
		if update.Update.ServiceName != nil {
			fileCfg.Update.ServiceName = strings.TrimSpace(*update.Update.ServiceName)
		}
		if update.Update.RestartCommand != nil {
			fileCfg.Update.RestartCommand = strings.TrimSpace(*update.Update.RestartCommand)
		}
	}
	if update.Connectivity != nil {
		if update.Connectivity.IntermittentPendingAgeSeconds != nil {
			fileCfg.Connectivity.IntermittentPendingAgeSeconds = max(1, *update.Connectivity.IntermittentPendingAgeSeconds)
		}
		if update.Connectivity.DegradedPendingAgeSeconds != nil {
			fileCfg.Connectivity.DegradedPendingAgeSeconds = max(1, *update.Connectivity.DegradedPendingAgeSeconds)
		}
		if update.Connectivity.IntermittentCloudSyncQueueDepth != nil {
			fileCfg.Connectivity.IntermittentCloudSyncQueueDepth = max(1, *update.Connectivity.IntermittentCloudSyncQueueDepth)
		}
		if update.Connectivity.DegradedCloudSyncQueueDepth != nil {
			fileCfg.Connectivity.DegradedCloudSyncQueueDepth = max(1, *update.Connectivity.DegradedCloudSyncQueueDepth)
		}
		if update.Connectivity.DegradedMQTTQueueDepth != nil {
			fileCfg.Connectivity.DegradedMQTTQueueDepth = max(1, *update.Connectivity.DegradedMQTTQueueDepth)
		}
		if update.Connectivity.HeartbeatHealthySeconds != nil {
			fileCfg.Connectivity.HeartbeatHealthySeconds = max(5, *update.Connectivity.HeartbeatHealthySeconds)
		}
		if update.Connectivity.HeartbeatDegradedSeconds != nil {
			fileCfg.Connectivity.HeartbeatDegradedSeconds = max(5, *update.Connectivity.HeartbeatDegradedSeconds)
		}
		if update.Connectivity.StatsHealthySeconds != nil {
			fileCfg.Connectivity.StatsHealthySeconds = max(10, *update.Connectivity.StatsHealthySeconds)
		}
		if update.Connectivity.StatsDegradedSeconds != nil {
			fileCfg.Connectivity.StatsDegradedSeconds = max(10, *update.Connectivity.StatsDegradedSeconds)
		}
	}
	if update.Retention != nil {
		if update.Retention.MQTTQueueMaxItems != nil {
			fileCfg.Retention.MQTTQueueMaxItems = max(100, *update.Retention.MQTTQueueMaxItems)
		}
		if update.Retention.MQTTQueueMaxAgeHours != nil {
			fileCfg.Retention.MQTTQueueMaxAgeHours = max(1, *update.Retention.MQTTQueueMaxAgeHours)
		}
		if update.Retention.CloudSyncSentRetentionHours != nil {
			fileCfg.Retention.CloudSyncSentRetentionHours = max(1, *update.Retention.CloudSyncSentRetentionHours)
		}
		if update.Retention.CloudSyncFailedRetentionHours != nil {
			fileCfg.Retention.CloudSyncFailedRetentionHours = max(1, *update.Retention.CloudSyncFailedRetentionHours)
		}
		if update.Retention.CloudSyncKeepSentItems != nil {
			fileCfg.Retention.CloudSyncKeepSentItems = max(100, *update.Retention.CloudSyncKeepSentItems)
		}
		if update.Retention.CloudSyncKeepFailedItems != nil {
			fileCfg.Retention.CloudSyncKeepFailedItems = max(100, *update.Retention.CloudSyncKeepFailedItems)
		}
		if update.Retention.LocalFallbackSentRetentionHours != nil {
			fileCfg.Retention.LocalFallbackSentRetentionHours = max(1, *update.Retention.LocalFallbackSentRetentionHours)
		}
		if update.Retention.LocalFallbackFailedRetentionHours != nil {
			fileCfg.Retention.LocalFallbackFailedRetentionHours = max(1, *update.Retention.LocalFallbackFailedRetentionHours)
		}
		if update.Retention.LocalFallbackKeepSentItems != nil {
			fileCfg.Retention.LocalFallbackKeepSentItems = max(100, *update.Retention.LocalFallbackKeepSentItems)
		}
		if update.Retention.LocalFallbackKeepFailedItems != nil {
			fileCfg.Retention.LocalFallbackKeepFailedItems = max(100, *update.Retention.LocalFallbackKeepFailedItems)
		}
	}
	return nil
}

func normalizeUpdateChannel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "beta":
		return "beta"
	default:
		return "stable"
	}
}

func defaultRetention() RetentionSection {
	return RetentionSection{
		MQTTQueueMaxItems:                 1000,
		MQTTQueueMaxAgeHours:              14 * 24,
		CloudSyncSentRetentionHours:       7 * 24,
		CloudSyncFailedRetentionHours:     30 * 24,
		CloudSyncKeepSentItems:            1000,
		CloudSyncKeepFailedItems:          1000,
		LocalFallbackSentRetentionHours:   7 * 24,
		LocalFallbackFailedRetentionHours: 30 * 24,
		LocalFallbackKeepSentItems:        1000,
		LocalFallbackKeepFailedItems:      1000,
	}
}

func normalizeRetention(section RetentionSection) RetentionSection {
	defaults := defaultRetention()
	if section.MQTTQueueMaxItems <= 0 {
		section.MQTTQueueMaxItems = defaults.MQTTQueueMaxItems
	}
	if section.MQTTQueueMaxAgeHours <= 0 {
		section.MQTTQueueMaxAgeHours = defaults.MQTTQueueMaxAgeHours
	}
	if section.CloudSyncSentRetentionHours <= 0 {
		section.CloudSyncSentRetentionHours = defaults.CloudSyncSentRetentionHours
	}
	if section.CloudSyncFailedRetentionHours <= 0 {
		section.CloudSyncFailedRetentionHours = defaults.CloudSyncFailedRetentionHours
	}
	if section.CloudSyncKeepSentItems <= 0 {
		section.CloudSyncKeepSentItems = defaults.CloudSyncKeepSentItems
	}
	if section.CloudSyncKeepFailedItems <= 0 {
		section.CloudSyncKeepFailedItems = defaults.CloudSyncKeepFailedItems
	}
	if section.LocalFallbackSentRetentionHours <= 0 {
		section.LocalFallbackSentRetentionHours = defaults.LocalFallbackSentRetentionHours
	}
	if section.LocalFallbackFailedRetentionHours <= 0 {
		section.LocalFallbackFailedRetentionHours = defaults.LocalFallbackFailedRetentionHours
	}
	if section.LocalFallbackKeepSentItems <= 0 {
		section.LocalFallbackKeepSentItems = defaults.LocalFallbackKeepSentItems
	}
	if section.LocalFallbackKeepFailedItems <= 0 {
		section.LocalFallbackKeepFailedItems = defaults.LocalFallbackKeepFailedItems
	}
	return section
}

func readPortFromEnv() int {
	value := strings.TrimSpace(os.Getenv("SPARKEDGE_HTTP_ADDR"))
	if value == "" {
		return 0
	}
	port, err := normalizePort(value)
	if err != nil {
		return 0
	}
	return port
}

func normalizePort(raw any) (int, error) {
	switch typed := raw.(type) {
	case string:
		value := strings.TrimSpace(strings.TrimPrefix(typed, ":"))
		if value == "" {
			return 0, errors.New("porta invalida")
		}
		port, err := strconv.Atoi(value)
		if err != nil {
			return 0, errors.New("porta invalida")
		}
		if port < 1 || port > 65535 {
			return 0, errors.New("porta invalida")
		}
		return port, nil
	case float64:
		return normalizePort(strconv.Itoa(int(typed)))
	case int:
		return normalizePort(strconv.Itoa(typed))
	default:
		return 0, errors.New("porta invalida")
	}
}

func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", min(len(secret)-4, 20))
}

func envBool(key string) (bool, bool) {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func envInt(key string) (int, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func maskOptionalSecret(secret string) string {
	if strings.TrimSpace(secret) == "" {
		return ""
	}
	return maskSecret(secret)
}
