package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/cloudsync"
	"github.com/kelwSagashi/sparkedge-go/internal/config"
	"github.com/kelwSagashi/sparkedge-go/internal/devices"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/edge"
	"github.com/kelwSagashi/sparkedge-go/internal/executions"
	"github.com/kelwSagashi/sparkedge-go/internal/instances"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/serverinfra"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/tags"
	"github.com/kelwSagashi/sparkedge-go/internal/updater"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type Dependencies struct {
	DB                  *sqlite.Store
	Auth                *auth.Service
	Users               *users.Service
	Projects            *projects.Service
	Scripts             *scripts.Service
	Devices             *devices.Service
	Edge                *edge.Service
	Tags                *tags.Service
	Instances           *instances.Service
	Executions          *executions.Service
	MQTT                *mqtt.Client
	Providers           *providers.Registry
	Runtime             *runtime.Runner
	TriggerInstance     func(context.Context, string, map[string]any, domain.TriggerType) (domain.InstanceExecution, runtime.TriggerResult, error)
	DispatchEvent       func(context.Context, string, map[string]any) (any, error)
	DispatchStateChange func(context.Context, map[string]any) (any, error)
	ServerInfra         *serverinfra.Service
	Updater             *updater.Service
	CloudSync           *cloudsync.Service
	Config              *config.Manager
	RuntimeCfg          config.Runtime
}

type Server struct {
	addr string
	deps Dependencies
}

func NewServer(addr string, deps Dependencies) *Server {
	return &Server{addr: addr, deps: deps}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	apiMux := http.NewServeMux()
	s.registerRoutes(apiMux)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || (!strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/api")) {
			s.handleFrontend(w, r)
			return
		}
		s.withMiddlewares(apiMux).ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              s.addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", Adapt(func(_ *http.Request) (any, error) {
		return map[string]any{
			"status": "OK",
			"app":    "sparkedge-go",
		}, nil
	}))

	mux.HandleFunc("POST /api/auth/register", Adapt(s.handleRegister))
	mux.HandleFunc("POST /api/auth/login", Adapt(s.handleLogin))
	mux.HandleFunc("POST /api/auth/logout", Adapt(s.handleLogout))
	mux.HandleFunc("GET /api/auth/me", Adapt(s.handleMe))
	mux.HandleFunc("POST /api/auth/generate-new-api-key/{userId}", Adapt(s.handleAuthGenerateAPIKey))

	mux.HandleFunc("GET /api/cli/onboarding", Adapt(s.handleCliOnboardingGet))
	mux.HandleFunc("POST /api/cli/onboarding", Adapt(s.handleCliOnboardingSave))
	mux.HandleFunc("GET /api/cli/status", Adapt(s.handleCliStatus))
	mux.HandleFunc("POST /api/cli/pair", Adapt(s.handleCliPair))
	mux.HandleFunc("POST /api/cli/connect", Adapt(s.handleCliConnect))
	mux.HandleFunc("POST /api/cli/disconnect", Adapt(s.handleCliDisconnect))
	mux.HandleFunc("POST /api/cli/reconnect", Adapt(s.handleCliReconnect))
	mux.HandleFunc("POST /api/cli/remove", Adapt(s.handleCliRemove))
	mux.HandleFunc("GET /api/cli/mqtt-config", Adapt(s.handleCliMQTTConfigGet))
	mux.HandleFunc("GET /api/cli/config", Adapt(s.handleCliConfigGet))
	mux.HandleFunc("PUT /api/cli/config", Adapt(s.handleCliConfigUpdate))
	mux.HandleFunc("GET /api/cli/operational-summary", Adapt(s.handleCliOperationalSummary))
	mux.HandleFunc("GET /api/update/check", Adapt(s.handleUpdateCheck))
	mux.HandleFunc("GET /api/update/status", Adapt(s.handleUpdateStatus))
	mux.HandleFunc("POST /api/update/download", Adapt(s.handleUpdateDownload))
	mux.HandleFunc("POST /api/update/apply", Adapt(s.handleUpdateApply))
	mux.HandleFunc("POST /api/update/rollback", Adapt(s.handleUpdateRollback))
	mux.HandleFunc("POST /api/update/restart", Adapt(s.handleUpdateRestart))
	mux.HandleFunc("POST /api/spark-cloud/auth/login", Adapt(s.handleSparkCloudLogin))
	mux.HandleFunc("POST /api/spark-cloud/edges/register", Adapt(s.handleSparkCloudEdgeRegister))
	mux.HandleFunc("GET /api/cloud-sync", Adapt(s.handleCloudSyncList))
	mux.HandleFunc("GET /api/cloud-sync/stats", Adapt(s.handleCloudSyncStats))
	mux.HandleFunc("POST /api/cloud-sync/flush", Adapt(s.handleCloudSyncFlush))
	mux.HandleFunc("POST /api/cloud-sync/{id}/retry", Adapt(s.handleCloudSyncRetry))
	mux.HandleFunc("DELETE /api/cloud-sync/{id}", Adapt(s.handleCloudSyncDelete))
	mux.HandleFunc("POST /api/events/dispatch", Adapt(s.handleEventDispatch))
	mux.HandleFunc("POST /api/state/dispatch", Adapt(s.handleStateDispatch))

	mux.HandleFunc("GET /api/users", Adapt(s.handleUsersList))
	mux.HandleFunc("POST /api/users", Adapt(s.handleUserCreate))
	mux.HandleFunc("GET /api/users/project/{id}/{project}", Adapt(s.handleUserProjectGet))
	mux.HandleFunc("GET /api/users/{id}", Adapt(s.handleUserGet))
	mux.HandleFunc("PUT /api/users/{id}", Adapt(s.handleUserUpdate))
	mux.HandleFunc("DELETE /api/users/{id}", Adapt(s.handleUserDelete))
	mux.HandleFunc("GET /api/users/{id}/api-key", Adapt(s.handleUserCreateAPIKey))

	mux.HandleFunc("GET /api/devices", Adapt(s.handleDevicesList))
	mux.HandleFunc("POST /api/devices", Adapt(s.handleDeviceCreate))
	mux.HandleFunc("GET /api/devices/{id}", Adapt(s.handleDeviceGet))
	mux.HandleFunc("PUT /api/devices/{id}", Adapt(s.handleDeviceUpdate))
	mux.HandleFunc("DELETE /api/devices/{id}", Adapt(s.handleDeviceDelete))

	mux.HandleFunc("GET /api/tags", Adapt(s.handleTagsList))
	mux.HandleFunc("GET /api/tags/search", Adapt(s.handleTagsSearch))
	mux.HandleFunc("POST /api/tags", Adapt(s.handleTagCreate))
	mux.HandleFunc("DELETE /api/tags/{id}", Adapt(s.handleTagDelete))

	mux.HandleFunc("GET /api/instances", Adapt(s.handleInstancesList))
	mux.HandleFunc("POST /api/instances", Adapt(s.handleInstanceCreate))
	mux.HandleFunc("GET /api/instances/", s.routeInstances)
	mux.HandleFunc("PUT /api/instances/", s.routeInstances)
	mux.HandleFunc("DELETE /api/instances/", s.routeInstances)
	mux.HandleFunc("POST /api/instances/", s.routeInstances)

	mux.HandleFunc("GET /api/instance-advanced", Adapt(s.handleInstancesList))
	mux.HandleFunc("POST /api/instance-advanced", Adapt(s.handleInstanceCreate))
	mux.HandleFunc("GET /api/instance-advanced/", s.routeInstanceAdvanced)
	mux.HandleFunc("PUT /api/instance-advanced/", s.routeInstanceAdvanced)
	mux.HandleFunc("DELETE /api/instance-advanced/", s.routeInstanceAdvanced)
	mux.HandleFunc("POST /api/instance-advanced/", s.routeInstanceAdvanced)

	mux.HandleFunc("GET /api/executions", Adapt(s.handleExecutionsList))
	mux.HandleFunc("GET /api/executions/instance/{instance_id}", Adapt(s.handleExecutionsByInstanceList))
	mux.HandleFunc("GET /api/executions/{id}", Adapt(s.handleExecutionGet))
	mux.HandleFunc("GET /api/fallback/stats", Adapt(s.handleFallbackStats))
	mux.HandleFunc("GET /api/fallback", Adapt(s.handleFallbackList))
	mux.HandleFunc("POST /api/fallback/flush", Adapt(s.handleFallbackFlush))
	mux.HandleFunc("POST /api/fallback/{id}/retry", Adapt(s.handleFallbackRetry))
	mux.HandleFunc("DELETE /api/fallback/{id}", Adapt(s.handleFallbackDelete))
	mux.HandleFunc("POST /api/webhook/{instanceId}", Adapt(s.handleWebhookReceive))

	mux.HandleFunc("GET /api/scripts", Adapt(s.handleScriptsList))
	mux.HandleFunc("GET /api/scripts/downloaded", Adapt(s.handleScriptsList))
	mux.HandleFunc("POST /api/scripts", Adapt(s.handleScriptCreate))
	mux.HandleFunc("POST /api/scripts/upload/inspect", Adapt(s.handleScriptUploadInspect))
	mux.HandleFunc("POST /api/scripts/upload/finalize", Adapt(s.handleScriptUploadFinalize))
	mux.HandleFunc("POST /api/scripts/draft/finalize", Adapt(s.handleScriptDraftFinalize))
	mux.HandleFunc("POST /api/scripts/draft/playground/run", Adapt(s.handleScriptDraftPlaygroundRun))
	mux.HandleFunc("POST /api/scripts/draft/readme", Adapt(s.handleScriptDraftReadme))
	mux.HandleFunc("POST /api/scripts/playground/run", Adapt(s.handleScriptPlaygroundRun))
	mux.HandleFunc("GET /api/scripts/", s.routeScripts)
	mux.HandleFunc("POST /api/scripts/", s.routeScripts)
	mux.HandleFunc("PUT /api/scripts/", s.routeScripts)
	mux.HandleFunc("DELETE /api/scripts/", s.routeScripts)

	mux.HandleFunc("GET /api/projects", Adapt(s.handleProjectsList))
	mux.HandleFunc("POST /api/projects", Adapt(s.handleProjectCreate))
	mux.HandleFunc("GET /api/projects/{id}", Adapt(s.handleProjectGet))
	mux.HandleFunc("PUT /api/projects/{id}", Adapt(s.handleProjectUpdate))
	mux.HandleFunc("DELETE /api/projects/{id}", Adapt(s.handleProjectDelete))
	mux.HandleFunc("GET /api/projects/{id}/members", Adapt(s.handleProjectMembersList))
	mux.HandleFunc("POST /api/projects/{id}/members", Adapt(s.handleProjectMemberAdd))

	mux.HandleFunc("GET /api/server-types", Adapt(s.handleServerTypesList))
	mux.HandleFunc("POST /api/server-types", Adapt(s.handleServerTypeCreate))
	mux.HandleFunc("GET /api/server-types/{id}", Adapt(s.handleServerTypeGet))
	mux.HandleFunc("PUT /api/server-types/{id}", Adapt(s.handleServerTypeUpdate))
	mux.HandleFunc("DELETE /api/server-types/{id}", Adapt(s.handleServerTypeDelete))
	mux.HandleFunc("GET /api/auth-types", Adapt(s.handleAuthTypesList))
	mux.HandleFunc("GET /api/credentials/config/meta", Adapt(s.handleAuthTypesList))
	mux.HandleFunc("GET /api/adapters/metadata", Adapt(s.handleAdaptersMetadata))
	mux.HandleFunc("POST /api/adapters/{id}/discover", Adapt(s.handleAdapterDiscover))
	mux.HandleFunc("GET /api/credentials", Adapt(s.handleCredentialsList))
	mux.HandleFunc("POST /api/credentials", Adapt(s.handleCredentialCreate))
	mux.HandleFunc("POST /api/credentials/test", Adapt(s.handleCredentialTest))
	mux.HandleFunc("GET /api/credentials/{id}", Adapt(s.handleCredentialGet))
	mux.HandleFunc("PUT /api/credentials/{id}", Adapt(s.handleCredentialUpdate))
	mux.HandleFunc("DELETE /api/credentials/{id}", Adapt(s.handleCredentialDelete))
	mux.HandleFunc("GET /api/servers", Adapt(s.handleServersList))
	mux.HandleFunc("POST /api/servers", Adapt(s.handleServerCreate))
	mux.HandleFunc("POST /api/servers/execute", Adapt(s.handleServerExecute))
	mux.HandleFunc("POST /api/servers/register", Adapt(s.handleServerRegister))
	mux.HandleFunc("GET /api/servers/{id}", Adapt(s.handleServerGet))
	mux.HandleFunc("PUT /api/servers/{id}", Adapt(s.handleServerUpdate))
	mux.HandleFunc("DELETE /api/servers/{id}", Adapt(s.handleServerDelete))
	mux.HandleFunc("GET /api/servers/{id}/resources", Adapt(s.handleServerResourcesList))
	mux.HandleFunc("GET /api/servers/{id}/endpoints", Adapt(s.handleServerResourcesList))
	mux.HandleFunc("GET /api/resources/{id}/operations", Adapt(s.handleResourceOperationsList))
}

func (s *Server) withMiddlewares(next http.Handler) http.Handler {
	return recoverMiddleware(corsMiddleware(authMiddleware(next, s.deps.Auth)))
}
