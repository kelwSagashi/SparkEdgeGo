package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/devices"
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
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type Dependencies struct {
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
}

type Server struct {
	addr string
	deps Dependencies
}

func NewServer(addr string, deps Dependencies) *Server {
	return &Server{addr: addr, deps: deps}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	server := &http.Server{
		Addr:              s.addr,
		Handler:           s.withMiddlewares(mux),
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
	mux.HandleFunc("POST /api/spark-cloud/auth/login", Adapt(s.handleSparkCloudLogin))
	mux.HandleFunc("POST /api/spark-cloud/edges/register", Adapt(s.handleSparkCloudEdgeRegister))

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
	mux.HandleFunc("GET /api/instances/active", Adapt(s.handleInstancesActiveList))
	mux.HandleFunc("GET /api/instances/project/{project_id}", Adapt(s.handleInstancesByProjectList))
	mux.HandleFunc("POST /api/instances", Adapt(s.handleInstanceCreate))
	mux.HandleFunc("GET /api/instances/{id}", Adapt(s.handleInstanceGet))
	mux.HandleFunc("PUT /api/instances/{id}", Adapt(s.handleInstanceUpdate))
	mux.HandleFunc("DELETE /api/instances/{id}", Adapt(s.handleInstanceDelete))
	mux.HandleFunc("POST /api/instances/{id}/trigger", Adapt(s.handleInstanceTrigger))
	mux.HandleFunc("GET /api/instances/{id}/executions", Adapt(s.handleInstanceExecutionsList))

	mux.HandleFunc("GET /api/instance-advanced", Adapt(s.handleInstancesList))
	mux.HandleFunc("GET /api/instance-advanced/project/{projectId}", Adapt(s.handleInstanceAdvancedByProjectList))
	mux.HandleFunc("POST /api/instance-advanced", Adapt(s.handleInstanceCreate))
	mux.HandleFunc("GET /api/instance-advanced/{id}", Adapt(s.handleInstanceGet))
	mux.HandleFunc("PUT /api/instance-advanced/{id}", Adapt(s.handleInstanceUpdate))
	mux.HandleFunc("DELETE /api/instance-advanced/{id}", Adapt(s.handleInstanceDelete))
	mux.HandleFunc("POST /api/instance-advanced/{id}/trigger", Adapt(s.handleInstanceTrigger))
	mux.HandleFunc("GET /api/instance-advanced/{id}/executions", Adapt(s.handleInstanceExecutionsList))
	mux.HandleFunc("GET /api/instance-advanced/{id}/executions/{executionId}", Adapt(s.handleInstanceAdvancedExecutionGet))
	mux.HandleFunc("GET /api/instance-advanced/{id}/destinations", Adapt(s.handleInstanceAdvancedDestinationsList))
	mux.HandleFunc("POST /api/instance-advanced/{id}/destinations", Adapt(s.handleInstanceAdvancedDestinationAdd))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/destinations/{destinationId}", Adapt(s.handleInstanceAdvancedDestinationUpdate))
	mux.HandleFunc("DELETE /api/instance-advanced/{id}/destinations/{destinationId}", Adapt(s.handleInstanceAdvancedDestinationDelete))
	mux.HandleFunc("GET /api/instance-advanced/{id}/available-fields", Adapt(s.handleInstanceAdvancedAvailableFields))
	mux.HandleFunc("POST /api/instance-advanced/{id}/mappings/test", Adapt(s.handleInstanceAdvancedMappingTest))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/destinations/{destinationId}/mapping", Adapt(s.handleInstanceAdvancedMappingSet))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/active", Adapt(s.handleInstanceAdvancedActiveUpdate))
	mux.HandleFunc("GET /api/instance-advanced/{id}/status", Adapt(s.handleInstanceAdvancedStatusGet))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/trigger-config", Adapt(s.handleInstanceAdvancedTriggerConfigUpdate))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/script-params", Adapt(s.handleInstanceAdvancedScriptParamsUpdate))
	mux.HandleFunc("PUT /api/instance-advanced/{id}/fallback-config", Adapt(s.handleInstanceAdvancedFallbackConfigUpdate))

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
	mux.HandleFunc("POST /api/scripts/playground/run", Adapt(s.handleScriptPlaygroundRun))
	mux.HandleFunc("GET /api/scripts/samples/list", Adapt(s.handleScriptSamplesList))
	mux.HandleFunc("GET /api/scripts/samples/{name}/schema", Adapt(s.handleScriptSampleSchema))
	mux.HandleFunc("GET /api/scripts/{id}/contents/{filename}", Adapt(s.handleScriptFileContent))
	mux.HandleFunc("GET /api/scripts/{id}", Adapt(s.handleScriptGet))
	mux.HandleFunc("PUT /api/scripts/{id}", Adapt(s.handleScriptUpdate))
	mux.HandleFunc("DELETE /api/scripts/{id}", Adapt(s.handleScriptDelete))

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
