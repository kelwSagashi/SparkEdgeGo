package app

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/cloudsync"
	"github.com/kelwSagashi/sparkedge-go/internal/config"
	"github.com/kelwSagashi/sparkedge-go/internal/devices"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/edge"
	"github.com/kelwSagashi/sparkedge-go/internal/executions"
	"github.com/kelwSagashi/sparkedge-go/internal/httpapi"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/firebaseprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/googleprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/httpprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/jsonfileprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/mongoprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/mqttprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/providers/supabaseprovider"
	"github.com/kelwSagashi/sparkedge-go/internal/python/sparkit"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/serverinfra"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
	"github.com/kelwSagashi/sparkedge-go/internal/updater"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type App struct {
	DB          *sqlite.Store
	Auth        *auth.Service
	Users       *users.Service
	Projects    *projects.Service
	Scripts     *scripts.Service
	Devices     *devices.Service
	Edge        *edge.Service
	Tags        *tags.Service
	Instances   *instances.Service
	Executions  *executions.Service
	MQTT        *mqtt.Client
	Providers   *providers.Registry
	Runtime     *runtime.Runner
	ServerInfra *serverinfra.Service
	Updater     *updater.Service
	CloudSync   *cloudsync.Service
	Config      *config.Manager
	RuntimeCfg  config.Runtime

	workflowMu       sync.Mutex
	workflowInflight map[string]bool
}

func New(cfg *config.Manager) *App {
	_, runtimeCfg, err := cfg.Load()
	if err != nil {
		panic(err)
	}

	providerRegistry := providers.NewRegistry()
	firebaseprovider.Register(providerRegistry)
	googleprovider.Register(providerRegistry)
	httpprovider.Register(providerRegistry)
	jsonfileprovider.Register(providerRegistry)
	mongoprovider.Register(providerRegistry)
	mqttprovider.Register(providerRegistry)
	supabaseprovider.Register(providerRegistry)
	sparkitExecutor := sparkit.NewExecutor()
	store := sqlite.NewStore()
	store.Path = runtimeCfg.DBFile
	sqlite.ConfigureRetention(runtimeCfg.Retention)
	if err := store.Open(context.Background()); err != nil {
		panic(err)
	}
	serverInfraService := serverinfra.NewService(store)
	serverTypes := []domain.ServerType{
		firebaseprovider.ServerType(),
		httpprovider.ServerType(),
		jsonfileprovider.ServerType(),
		mongoprovider.ServerType(),
		mqttprovider.ServerType(),
		supabaseprovider.ServerType(),
	}
	serverTypes = append(serverTypes, googleprovider.ServerTypes()...)
	authTypes := append(firebaseprovider.AuthTypes(), googleprovider.AuthTypes()...)
	authTypes = append(authTypes, httpprovider.AuthTypes()...)
	authTypes = append(authTypes, jsonfileprovider.AuthTypes()...)
	authTypes = append(authTypes, mongoprovider.AuthTypes()...)
	authTypes = append(authTypes, mqttprovider.AuthTypes()...)
	authTypes = append(authTypes, supabaseprovider.AuthTypes()...)
	if err := serverInfraService.SeedCatalog(context.Background(), serverTypes, authTypes); err != nil {
		panic(err)
	}
	mqttClient := mqtt.NewClient()
	mqttClient.UseStores(store.MqttCommands, store.MqttQueue)
	edgeService := edge.NewService(store.Edge, edge.NewHTTPCloudClient(runtimeCfg.CloudURL), mqttClient)
	tagsService := tags.NewService(store.Tags, store.InstanceTags)
	cloudSyncService := cloudsync.NewService(store.CloudSync, cloudsync.Config{
		BaseURL:   runtimeCfg.CloudURL,
		EdgeID:    "",
		SyncToken: runtimeCfg.CloudSyncToken,
		Enabled:   true,
	})

	application := &App{
		DB:         store,
		Auth:       auth.NewService(store.Users, store.Projects, runtimeCfg.JWTSecret),
		Users:      users.NewService(store.Users),
		Projects:   projects.NewService(store.Projects, store.ProjectMembers),
		Scripts:    scripts.NewService(store.Scripts, sparkitExecutor),
		Devices:    devices.NewService(store.Devices),
		Edge:       edgeService,
		Tags:       tagsService,
		Instances:  instances.NewService(store.Instances, tagsService, store.Destinations, store.DataMappings, store.CircuitBreakers),
		Executions: executions.NewService(store.Executions),
		MQTT:       mqttClient,
		Providers:  providerRegistry,
		Runtime: runtime.NewRunner(runtime.Dependencies{
			Sparkit:            sparkitExecutor,
			Providers:          providerRegistry,
			ResourceOperations: store.ResourceOperations,
			Fallback:           store.LocalFallback,
			Destinations:       store.Destinations,
			CircuitBreakers:    store.CircuitBreakers,
			Devices:            store.Devices,
			EdgeConfig:         store.Edge,
		}),
		ServerInfra: serverInfraService,
		Updater: updater.NewService(updater.Config{
			Enabled:         runtimeCfg.Update.Enabled,
			Provider:        runtimeCfg.Update.Provider,
			Repo:            runtimeCfg.Update.Repo,
			Channel:         runtimeCfg.Update.Channel,
			AllowPrerelease: runtimeCfg.Update.AllowPrerelease,
			ServiceName:     runtimeCfg.Update.ServiceName,
			RestartCommand:  runtimeCfg.Update.RestartCommand,
		}, &updater.GitHubClient{Token: strings.TrimSpace(os.Getenv("SPARKEDGE_UPDATE_GITHUB_TOKEN"))}),
		CloudSync:  cloudSyncService,
		Config:     cfg,
		RuntimeCfg: runtimeCfg,
		workflowInflight: map[string]bool{},
	}
	application.MQTT.SetStatsProvider(application.collectOperationalSnapshot)
	application.registerMqttCommandHandlers()
	return application
}

func (a *App) HTTPServer(addr string) *httpapi.Server {
	return httpapi.NewServer(addr, httpapi.Dependencies{
		DB:         a.DB,
		Auth:       a.Auth,
		Users:      a.Users,
		Projects:   a.Projects,
		Scripts:    a.Scripts,
		Devices:    a.Devices,
		Edge:       a.Edge,
		Tags:       a.Tags,
		Instances:  a.Instances,
		Executions: a.Executions,
		MQTT:       a.MQTT,
		Providers:  a.Providers,
		Runtime:    a.Runtime,
		TriggerInstance: func(ctx context.Context, instanceID string, input map[string]any, triggerType domain.TriggerType) (domain.InstanceExecution, runtime.TriggerResult, error) {
			return a.triggerInstance(ctx, instanceID, input, triggerType)
		},
		DispatchEvent: func(ctx context.Context, eventName string, payload map[string]any) (any, error) {
			return a.DispatchEvent(ctx, eventName, payload)
		},
		DispatchStateChange: func(ctx context.Context, payload map[string]any) (any, error) {
			return a.DispatchStateChange(ctx, payload)
		},
		ServerInfra: a.ServerInfra,
		Updater:     a.Updater,
		CloudSync:   a.CloudSync,
		Config:      a.Config,
		RuntimeCfg:  a.RuntimeCfg,
	})
}
