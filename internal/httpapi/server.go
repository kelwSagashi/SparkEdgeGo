package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/auth"
	"github.com/kelwSagashi/sparkedge-go/internal/mqtt"
	"github.com/kelwSagashi/sparkedge-go/internal/projects"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
	"github.com/kelwSagashi/sparkedge-go/internal/runtime"
	"github.com/kelwSagashi/sparkedge-go/internal/scripts"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"github.com/kelwSagashi/sparkedge-go/internal/users"
)

type Dependencies struct {
	DB        *sqlite.Store
	Auth      *auth.Service
	Users     *users.Service
	Projects  *projects.Service
	Scripts   *scripts.Service
	MQTT      *mqtt.Client
	Providers *providers.Registry
	Runtime   *runtime.Runner
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
	mux.HandleFunc("GET /api/auth/me", Adapt(s.handleMe))

	mux.HandleFunc("GET /api/users", Adapt(s.handleUsersList))
	mux.HandleFunc("POST /api/users", Adapt(s.handleUserCreate))
	mux.HandleFunc("GET /api/users/project/{id}/{project}", Adapt(s.handleUserProjectGet))
	mux.HandleFunc("GET /api/users/{id}", Adapt(s.handleUserGet))
	mux.HandleFunc("PUT /api/users/{id}", Adapt(s.handleUserUpdate))
	mux.HandleFunc("DELETE /api/users/{id}", Adapt(s.handleUserDelete))
	mux.HandleFunc("GET /api/users/{id}/api-key", Adapt(s.handleUserCreateAPIKey))

	mux.HandleFunc("GET /api/scripts", Adapt(s.handleScriptsList))
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
}

func (s *Server) withMiddlewares(next http.Handler) http.Handler {
	return recoverMiddleware(corsMiddleware(authMiddleware(next, s.deps.Auth)))
}
