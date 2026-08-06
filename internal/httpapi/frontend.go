package httpapi

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	frontendassets "github.com/kelwSagashi/sparkedge-go/frontend"
)

func (s *Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	requestPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if requestPath == "." {
		requestPath = ""
	}

	source, err := resolveFrontendSource()
	if err != nil {
		RespondError(w, NewHTTPError(http.StatusServiceUnavailable, err.Error()))
		return
	}

	source.serve(w, r, requestPath)
}

type frontendSource interface {
	serve(w http.ResponseWriter, r *http.Request, requestPath string)
}

type diskFrontendSource struct {
	distDir string
}

type embeddedFrontendSource struct {
	distFS fs.FS
}

func resolveFrontendSource() (frontendSource, error) {
	candidates := []string{}
	if env := strings.TrimSpace(os.Getenv("SPARKEDGE_FRONTEND_DIST")); env != "" {
		candidates = append(candidates, env)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "frontend", "dist"))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "frontend", "dist"),
			filepath.Join(exeDir, "..", "frontend", "dist"),
		)
	}
	for _, candidate := range candidates {
		if fileExists(filepath.Join(candidate, "index.html")) {
			return diskFrontendSource{distDir: candidate}, nil
		}
	}

	distFS, err := frontendassets.Dist()
	if err == nil {
		if _, statErr := fs.Stat(distFS, "index.html"); statErr == nil {
			return embeddedFrontendSource{distFS: distFS}, nil
		}
	}

	return nil, NewHTTPError(http.StatusServiceUnavailable, "frontend assets not found; build frontend/dist before compiling the production binary")
}

func (s diskFrontendSource) serve(w http.ResponseWriter, r *http.Request, requestPath string) {
	if requestPath == "" {
		http.ServeFile(w, r, filepath.Join(s.distDir, "index.html"))
		return
	}

	target := filepath.Join(s.distDir, filepath.FromSlash(requestPath))
	if fileExists(target) {
		http.ServeFile(w, r, target)
		return
	}

	if filepath.Ext(requestPath) != "" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filepath.Join(s.distDir, "index.html"))
}

func (s embeddedFrontendSource) serve(w http.ResponseWriter, r *http.Request, requestPath string) {
	if requestPath == "" {
		s.serveEmbeddedFile(w, r, "index.html")
		return
	}

	cleanPath := filepath.ToSlash(requestPath)
	if embeddedFileExists(s.distFS, cleanPath) {
		s.serveEmbeddedFile(w, r, cleanPath)
		return
	}

	if filepath.Ext(requestPath) != "" {
		http.NotFound(w, r)
		return
	}

	s.serveEmbeddedFile(w, r, "index.html")
}

func (s embeddedFrontendSource) serveEmbeddedFile(w http.ResponseWriter, r *http.Request, name string) {
	file, err := s.distFS.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := fs.Stat(s.distFS, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, name, info.ModTime(), file.(io.ReadSeeker))
}

func embeddedFileExists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
