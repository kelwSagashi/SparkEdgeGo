package app

import (
	"context"
	"os"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/devices"
	"github.com/kelwSagashi/sparkedge-go/internal/executions"
	"github.com/kelwSagashi/sparkedge-go/internal/httpapi"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/python/sparkit"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/serverinfra"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type App struct {
	DB          *sqlite.Store
	Auth        *auth.Service
	Users       *users.Service
	Projects    *projects.Service
	Scripts     *scripts.Service
	Devices     *devices.Service
	Tags        *tags.Service
	Instances   *instances.Service
	Executions  *executions.Service
	MQTT        *mqtt.Client
	Providers   *providers.Registry
	Runtime     *runtime.Runner
	ServerInfra *serverinfra.Service
}

func New() *App {
	providerRegistry := providers.NewRegistry()
	sparkitExecutor := sparkit.NewExecutor()
	store := sqlite.NewStore()
	if err := store.Open(context.Background()); err != nil {
		panic(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	tagsService := tags.NewService(store.Tags, store.InstanceTags)

	return &App{
		DB:         store,
		Auth:       auth.NewService(store.Users, store.Projects, jwtSecret),
		Users:      users.NewService(store.Users),
		Projects:   projects.NewService(store.Projects, store.ProjectMembers),
		Scripts:    scripts.NewService(store.Scripts, sparkitExecutor),
		Devices:    devices.NewService(store.Devices),
		Tags:       tagsService,
		Instances:  instances.NewService(store.Instances, tagsService, store.Destinations, store.DataMappings),
		Executions: executions.NewService(store.Executions),
		MQTT:       mqtt.NewClient(),
		Providers:  providerRegistry,
		Runtime: runtime.NewRunner(runtime.Dependencies{
			Sparkit:            sparkitExecutor,
			Providers:          providerRegistry,
			ResourceOperations: store.ResourceOperations,
		}),
		ServerInfra: serverinfra.NewService(store),
	}
}

func (a *App) HTTPServer(addr string) *httpapi.Server {
	return httpapi.NewServer(addr, httpapi.Dependencies{
		DB:          a.DB,
		Auth:        a.Auth,
		Users:       a.Users,
		Projects:    a.Projects,
		Scripts:     a.Scripts,
		Devices:     a.Devices,
		Tags:        a.Tags,
		Instances:   a.Instances,
		Executions:  a.Executions,
		MQTT:        a.MQTT,
		Providers:   a.Providers,
		Runtime:     a.Runtime,
		ServerInfra: a.ServerInfra,
	})
}
