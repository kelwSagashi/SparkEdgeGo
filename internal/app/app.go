package app

import (
	"context"
	"os"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/httpapi"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/python/sparkit"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type App struct {
	DB        *sqlite.Store
	Auth      *auth.Service
	Users     *users.Service
	Projects  *projects.Service
	Scripts   *scripts.Service
	MQTT      *mqtt.Client
	Providers *providers.Registry
	Runtime   *runtime.Runner
}

func New() *App {
	providerRegistry := providers.NewRegistry()
	store := sqlite.NewStore()
	if err := store.Open(context.Background()); err != nil {
		panic(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	return &App{
		DB:        store,
		Auth:      auth.NewService(store.Users, store.Projects, jwtSecret),
		Users:     users.NewService(store.Users),
		Projects:  projects.NewService(store.Projects, store.ProjectMembers),
		Scripts:   scripts.NewService(store.Scripts),
		MQTT:      mqtt.NewClient(),
		Providers: providerRegistry,
		Runtime: runtime.NewRunner(runtime.Dependencies{
			Sparkit:   sparkit.NewExecutor(),
			Providers: providerRegistry,
		}),
	}
}

func (a *App) HTTPServer(addr string) *httpapi.Server {
	return httpapi.NewServer(addr, httpapi.Dependencies{
		DB:        a.DB,
		Auth:      a.Auth,
		Users:     a.Users,
		Projects:  a.Projects,
		Scripts:   a.Scripts,
		MQTT:      a.MQTT,
		Providers: a.Providers,
		Runtime:   a.Runtime,
	})
}
