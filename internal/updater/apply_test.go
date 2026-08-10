package updater

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyDownloadedPreservesConfigAndUpdatesFiles(t *testing.T) {
	appRoot := t.TempDir()
	t.Setenv("SPARKEDGE_HOME", appRoot)
	t.Setenv("SPARKEDGE_TARGET_LABEL", "linux-amd64")

	writeTestFile(t, filepath.Join(appRoot, "sparkedge"), "old-binary")
	writeTestFile(t, filepath.Join(appRoot, "README.md"), "old-readme")
	writeTestFile(t, filepath.Join(appRoot, "version.txt"), "Version: v0.1.0\nTarget: linux-amd64\n")
	writeTestFile(t, filepath.Join(appRoot, "config", ".env.example"), "old-env")
	writeTestFile(t, filepath.Join(appRoot, "config.yml"), "cloud:\n  url: https://current.example\n")
	writeTestFile(t, filepath.Join(appRoot, "sparkedge.db"), "db-state")
	writeTestFile(t, filepath.Join(appRoot, "webui", "dist", "index.html"), "old-ui")

	zipPath := filepath.Join(appRoot, "updates", "downloads", "v0.2.0", "sparkedge-v0.2.0-linux-amd64.zip")
	createTestPackageZip(t, zipPath, map[string]string{
		"sparkEdge/sparkedge":              "new-binary",
		"sparkEdge/README.md":              "new-readme",
		"sparkEdge/version.txt":            "Version: v0.2.0\nTarget: linux-amd64\n",
		"sparkEdge/config.yml":             "cloud:\n  url: https://package.example\n",
		"sparkEdge/config/.env.example":    "new-env",
		"sparkEdge/webui/dist/index.html":  "new-ui",
		"sparkEdge/webui/dist/assets/a.js": "console.log('ok')",
	})

	service := NewService(Config{}, fakeReleaseClient{})
	result, err := service.ApplyDownloaded(context.Background(), zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Applied || result.PreparedOnly {
		t.Fatalf("expected direct apply result, got %#v", result)
	}

	assertFileContains(t, filepath.Join(appRoot, "sparkedge"), "new-binary")
	assertFileContains(t, filepath.Join(appRoot, "README.md"), "new-readme")
	assertFileContains(t, filepath.Join(appRoot, "webui", "dist", "index.html"), "new-ui")
	assertFileContains(t, filepath.Join(appRoot, "config", ".env.example"), "new-env")
	assertFileContains(t, filepath.Join(appRoot, "config.yml"), "https://current.example")
	assertFileContains(t, filepath.Join(appRoot, "sparkedge.db"), "db-state")
	assertFileContains(t, filepath.Join(result.BackupPath, "sparkedge"), "old-binary")
	assertFileContains(t, filepath.Join(result.BackupPath, "config.yml"), "https://current.example")
}

func TestApplyDownloadedPreparesScriptForWindowsTargets(t *testing.T) {
	appRoot := t.TempDir()
	t.Setenv("SPARKEDGE_HOME", appRoot)
	t.Setenv("SPARKEDGE_TARGET_LABEL", "windows-amd64")

	writeTestFile(t, filepath.Join(appRoot, "sparkedge.exe"), "old-binary")
	writeTestFile(t, filepath.Join(appRoot, "README.md"), "old-readme")
	writeTestFile(t, filepath.Join(appRoot, "version.txt"), "Version: v0.1.0\nTarget: windows-amd64\n")
	writeTestFile(t, filepath.Join(appRoot, "config", ".env.example"), "old-env")
	writeTestFile(t, filepath.Join(appRoot, "config.yml"), "cloud:\n  url: https://current.example\n")
	writeTestFile(t, filepath.Join(appRoot, "webui", "dist", "index.html"), "old-ui")

	zipPath := filepath.Join(appRoot, "updates", "downloads", "v0.2.0", "sparkedge-v0.2.0-windows-amd64.zip")
	createTestPackageZip(t, zipPath, map[string]string{
		"sparkEdge/sparkedge.exe":          "new-binary",
		"sparkEdge/README.md":              "new-readme",
		"sparkEdge/version.txt":            "Version: v0.2.0\nTarget: windows-amd64\n",
		"sparkEdge/config.yml":             "cloud:\n  url: https://package.example\n",
		"sparkEdge/config/.env.example":    "new-env",
		"sparkEdge/webui/dist/index.html":  "new-ui",
		"sparkEdge/webui/dist/assets/a.js": "console.log('ok')",
	})

	service := NewService(Config{}, fakeReleaseClient{})
	result, err := service.ApplyDownloaded(context.Background(), zipPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.PreparedOnly || result.Applied {
		t.Fatalf("expected prepared-only result, got %#v", result)
	}
	if result.ScriptPath == "" {
		t.Fatalf("expected apply script path, got %#v", result)
	}
	assertFileContains(t, result.ScriptPath, "Applying SparkEdge update")
	assertFileContains(t, filepath.Join(appRoot, "sparkedge.exe"), "old-binary")
}

func createTestPackageZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		t.Fatalf("mkdir zip dir: %v", err)
	}
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, contents := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir test file dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), expected) {
		t.Fatalf("expected %s to contain %q, got %q", path, expected, string(data))
	}
}
