package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
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
	Cloud  CloudSection  `yaml:"cloud"`
	DB     DBSection     `yaml:"db"`
	Auth   AuthSection   `yaml:"auth"`
	Server ServerSection `yaml:"server"`
	Update UpdateSection `yaml:"update"`
}

type Effective struct {
	Cloud      CloudSection `json:"cloud"`
	DB         DBSection    `json:"db"`
	Auth       AuthView     `json:"auth"`
	Server     ServerView   `json:"server"`
	Update     UpdateView   `json:"update"`
	ConfigFile string       `json:"config_file,omitempty"`
}

type CloudSection struct {
	URL     string `json:"url" yaml:"url"`
	MQTTURL string `json:"mqtt_url" yaml:"mqtt_url"`
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

type Update struct {
	Cloud  *CloudSectionUpdate  `json:"cloud"`
	DB     *DBSectionUpdate     `json:"db"`
	Auth   *AuthSectionUpdate   `json:"auth"`
	Server *ServerSectionUpdate `json:"server"`
}

type CloudSectionUpdate struct {
	URL     *string `json:"url"`
	MQTTURL *string `json:"mqtt_url"`
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

type Runtime struct {
	CloudURL   string
	MQTTURL    string
	DBFile     string
	JWTSecret  string
	HTTPPort   int
	ConfigFile string
	Update     UpdateRuntime
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
		CloudURL:   defaultCloudURL,
		MQTTURL:    defaultMQTTURL,
		DBFile:     defaultDBFile,
		JWTSecret:  defaultJWTSecret,
		HTTPPort:   defaultHTTPPort,
		ConfigFile: m.path,
		Update: UpdateRuntime{
			Enabled:         true,
			Provider:        "github",
			Repo:            "kelwSagashi/SparkEdgeGo",
			Channel:         "stable",
			AllowPrerelease: false,
		},
	}

	applyEnv(&runtimeCfg)
	applyFile(&runtimeCfg, fileCfg)

	effective := Effective{
		Cloud: CloudSection{
			URL:     runtimeCfg.CloudURL,
			MQTTURL: runtimeCfg.MQTTURL,
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
		ConfigFile: m.path,
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
	return nil
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
