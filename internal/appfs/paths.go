package appfs

import (
	"os"
	"path/filepath"
	"strings"
)

func AppRoot() string {
	if env := strings.TrimSpace(os.Getenv("SPARKEDGE_HOME")); env != "" {
		return env
	}

	if cwd, err := os.Getwd(); err == nil && looksLikeAppRoot(cwd) {
		return cwd
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if looksLikeAppRoot(exeDir) {
			return exeDir
		}
		return exeDir
	}

	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return "."
}

func ResolveFromRoot(parts ...string) string {
	all := append([]string{AppRoot()}, parts...)
	return filepath.Join(all...)
}

func looksLikeAppRoot(path string) bool {
	return fileExists(filepath.Join(path, "go.mod")) ||
		dirExists(filepath.Join(path, "webui")) ||
		dirExists(filepath.Join(path, "banco")) ||
		dirExists(filepath.Join(path, "config"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
