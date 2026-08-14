package appmeta

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
)

type VersionInfo struct {
	Version string
	Target  string
}

func LoadVersionInfo() VersionInfo {
	info := VersionInfo{
		Version: fallbackVersion(),
		Target:  fallbackTarget(),
	}

	path := appfs.ResolveFromRoot("version.txt")
	file, err := os.Open(path)
	if err != nil {
		return info
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "Version:"):
			if value := strings.TrimSpace(strings.TrimPrefix(line, "Version:")); value != "" {
				info.Version = value
			}
		case strings.HasPrefix(line, "Target:"):
			if value := strings.TrimSpace(strings.TrimPrefix(line, "Target:")); value != "" {
				info.Target = value
			}
		}
	}

	return info
}

func fallbackVersion() string {
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_VERSION")); value != "" {
		return value
	}
	return "dev"
}

func fallbackTarget() string {
	if value := strings.TrimSpace(os.Getenv("SPARKEDGE_TARGET_LABEL")); value != "" {
		return value
	}

	if runtime.GOARCH == "arm" {
		if value := strings.TrimSpace(os.Getenv("GOARM")); value != "" {
			return runtime.GOOS + "-armv" + value
		}
		return runtime.GOOS + "-arm"
	}

	return runtime.GOOS + "-" + runtime.GOARCH
}

func VersionFilePath() string {
	return filepath.Clean(appfs.ResolveFromRoot("version.txt"))
}
