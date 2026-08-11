package scripts

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestScriptsServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Scripts)
	script, err := service.Create(ctx, CreateRequest{
		Name:      "Temperature Collector",
		Author:    "SparkEdge",
		LocalPath: filepath.Join(t.TempDir(), "script"),
		MainFile:  "main.py",
		Tags:      []string{"sensor", "temperature"},
		SchemaConfig: map[string]any{
			"inputs": []any{},
			"outputs": []any{
				map[string]any{"name": "temperature", "type": "number"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if script.Source != domain.ScriptSourceLocal {
		t.Fatalf("expected local source, got %s", script.Source)
	}
	if script.Language != domain.ScriptLanguagePython {
		t.Fatalf("expected python language, got %s", script.Language)
	}

	history, err := service.History(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Action != ScriptHistoryActionInstalled {
		t.Fatalf("expected initial install history, got %#v", history)
	}

	found, err := service.FindByID(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Tags) != 2 || found.SchemaConfig["outputs"] == nil {
		t.Fatalf("expected tags and schema_config to round-trip, got %#v", found)
	}

	version := "2.0.0"
	updated, err := service.Update(ctx, script.ID, UpdateRequest{Version: &version})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != version {
		t.Fatalf("expected version %q, got %q", version, updated.Version)
	}

	history, err = service.History(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Action != ScriptHistoryActionMetadataUpdated {
		t.Fatalf("expected metadata update history, got %#v", history)
	}

	items, err := service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one script, got %d", len(items))
	}

	if err := service.Delete(ctx, script.ID); err != nil {
		t.Fatal(err)
	}
	items, err = service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no scripts after delete, got %d", len(items))
	}
}

func TestInspectZipFindsPythonFilesAndSparkitRequirement(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "script.zip")
	createTestZip(t, zipPath, map[string]string{
		"main.py":          "print('ok')",
		"requirements.txt": "requests\nsparkit\n",
	})

	service := NewService(nil)
	result, err := service.InspectZip(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(result.TempFolder)

	if !result.HasSparkit {
		t.Fatal("expected sparkit requirement to be detected")
	}
	if len(result.PyFiles) != 1 || result.PyFiles[0] != "main.py" {
		t.Fatalf("expected main.py in inspect result, got %#v", result.PyFiles)
	}
}

func TestFileContentReadsScriptFile(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	scriptDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(scriptDir, "main.py"), []byte("print('hello')"), 0o644); err != nil {
		t.Fatal(err)
	}

	service := NewService(store.Scripts)
	script, err := service.Create(ctx, CreateRequest{
		Name:      "Reader",
		Author:    "SparkEdge",
		LocalPath: scriptDir,
		MainFile:  "main.py",
	})
	if err != nil {
		t.Fatal(err)
	}

	content, err := service.FileContent(ctx, script.ID, "main.py")
	if err != nil {
		t.Fatal(err)
	}
	if content != "print('hello')" {
		t.Fatalf("unexpected file content %q", content)
	}
}

func TestFileContentFindsReadmeInsideNestedFolder(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	scriptDir := t.TempDir()
	nestedDir := filepath.Join(scriptDir, "bundle")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "README.md"), []byte("# Nested Readme"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "main.py"), []byte("print('nested')"), 0o644); err != nil {
		t.Fatal(err)
	}

	service := NewService(store.Scripts)
	script, err := service.Create(ctx, CreateRequest{
		Name:      "Nested",
		Author:    "SparkEdge",
		LocalPath: scriptDir,
		MainFile:  "bundle/main.py",
	})
	if err != nil {
		t.Fatal(err)
	}

	content, err := service.FileContent(ctx, script.ID, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if content != "# Nested Readme" {
		t.Fatalf("unexpected nested README content %q", content)
	}

	mainCode, err := service.FileContent(ctx, script.ID, "bundle/main.py")
	if err != nil {
		t.Fatal(err)
	}
	if mainCode != "print('nested')" {
		t.Fatalf("unexpected nested main content %q", mainCode)
	}
}

func TestReplaceUploadPreservesScriptIDAndRefreshesBundle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	homeDir := t.TempDir()
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("HOME", homeDir)

	runtime := &fakePythonRuntime{
		schema: map[string]any{
			"inputs":  []any{map[string]any{"name": "asset_id", "type": "string"}},
			"outputs": []any{map[string]any{"name": "temperature", "type": "number"}},
		},
	}
	service := NewService(store.Scripts, runtime)

	originalFolder := filepath.Join(homeDir, ".spark_edge", "scripts", "existing-script")
	if err := os.MkdirAll(originalFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(originalFolder, "main.py"), []byte("print('old')"), 0o644); err != nil {
		t.Fatal(err)
	}

	script, err := service.Create(ctx, CreateRequest{
		ID:           "existing-script",
		Name:         "Original Script",
		Author:       "SparkEdge",
		Description:  "before",
		Version:      "1.0.0",
		LocalPath:    toRelativePath(originalFolder),
		MainFile:     "main.py",
		VenvPath:     toRelativePath(filepath.Join(originalFolder, "venv")),
		VenvReady:    true,
		SchemaConfig: map[string]any{"inputs": []any{}},
	})
	if err != nil {
		t.Fatal(err)
	}

	uploadFolder := t.TempDir()
	nested := filepath.Join(uploadFolder, "bundle")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "main.py"), []byte("print('new')"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "requirements.txt"), []byte("sparkit\nrequests\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := service.ReplaceUpload(ctx, script.ID, FinalizeRequest{
		TempFolder:  uploadFolder,
		MainFile:    "bundle/main.py",
		Name:        "Updated Script",
		Description: "after",
		Author:      "SparkEdge Team",
		Version:     "2.0.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Script.ID != script.ID {
		t.Fatalf("expected script ID %q, got %q", script.ID, result.Script.ID)
	}
	if result.Script.MainFile != "bundle/main.py" {
		t.Fatalf("expected updated main file, got %q", result.Script.MainFile)
	}
	if result.Script.Version != "2.0.0" {
		t.Fatalf("expected updated version, got %q", result.Script.Version)
	}
	if result.Script.SchemaConfig["outputs"] == nil {
		t.Fatalf("expected refreshed schema config, got %#v", result.Script.SchemaConfig)
	}
	if runtime.lastMainFile != "bundle/main.py" {
		t.Fatalf("expected schema extraction from new entrypoint, got %q", runtime.lastMainFile)
	}

	content, err := service.FileContent(ctx, script.ID, "bundle/main.py")
	if err != nil {
		t.Fatal(err)
	}
	if content != "print('new')" {
		t.Fatalf("expected updated code, got %q", content)
	}

	history, err := service.History(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("expected two history entries, got %#v", history)
	}
	if history[0].Action != ScriptHistoryActionBundleUpdated {
		t.Fatalf("expected latest history action %q, got %#v", ScriptHistoryActionBundleUpdated, history[0])
	}
	if history[1].Action != ScriptHistoryActionInstalled {
		t.Fatalf("expected oldest history action %q, got %#v", ScriptHistoryActionInstalled, history[1])
	}
	if len(history[0].ChangeSummary) == 0 {
		t.Fatalf("expected change summary for updated bundle, got %#v", history[0])
	}
	if history[1].BundlePath == "" {
		t.Fatalf("expected initial bundle snapshot path, got %#v", history[1])
	}

	restored, err := service.RestoreHistory(ctx, script.ID, history[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Script.MainFile != "main.py" {
		t.Fatalf("expected restored main file %q, got %q", "main.py", restored.Script.MainFile)
	}
	restoredContent, err := service.FileContent(ctx, script.ID, "main.py")
	if err != nil {
		t.Fatal(err)
	}
	if restoredContent != "print('old')" {
		t.Fatalf("expected restored code, got %q", restoredContent)
	}

	history, err = service.History(ctx, script.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 {
		t.Fatalf("expected three history entries after restore, got %#v", history)
	}
	if history[0].Action != ScriptHistoryActionRestored {
		t.Fatalf("expected latest history action %q, got %#v", ScriptHistoryActionRestored, history[0])
	}
}

func TestListSamplesAndRunSamplePlayground(t *testing.T) {
	ctx := context.Background()
	samplesDir := t.TempDir()
	sampleDir := filepath.Join(samplesDir, "demo")
	if err := os.MkdirAll(sampleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sampleDir, "main.py"), []byte("print('demo')"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SPARKEDGE_SAMPLES_DIR", samplesDir)

	runtime := &fakePythonRuntime{
		schema: map[string]any{"inputs": []any{}},
		result: domain.ScriptResult{ExitCode: 0, Data: map[string]any{
			"stdout": map[string]any{"ok": true},
		}},
	}
	service := NewService(nil, runtime)

	samples, err := service.ListSamples()
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 || samples[0] != "demo" {
		t.Fatalf("expected demo sample, got %#v", samples)
	}

	schema, err := service.SampleSchema(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if schema["inputs"] == nil {
		t.Fatalf("expected sample schema, got %#v", schema)
	}

	result, err := service.RunPlayground(ctx, PlaygroundRequest{SampleName: "demo", Inputs: map[string]any{"ip": "127.0.0.1"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 || runtime.lastMainFile != "main.py" {
		t.Fatalf("unexpected playground result %#v with main file %q", result, runtime.lastMainFile)
	}
}

func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()

	handle, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()

	writer := zip.NewWriter(handle)
	defer writer.Close()

	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}

type fakePythonRuntime struct {
	schema       map[string]any
	result       domain.ScriptResult
	lastFolder   string
	lastMainFile string
	lastVenvPath string
}

func (f *fakePythonRuntime) CreateVenv(_ context.Context, _ string) error {
	return nil
}

func (f *fakePythonRuntime) InstallRequirements(_ context.Context, _ string, _ string) error {
	return nil
}

func (f *fakePythonRuntime) SchemaFile(_ context.Context, scriptFolder string, mainFile string, venvPath string) (map[string]any, error) {
	f.lastFolder = scriptFolder
	f.lastMainFile = mainFile
	f.lastVenvPath = venvPath
	return f.schema, nil
}

func (f *fakePythonRuntime) RunFile(_ context.Context, scriptFolder string, mainFile string, venvPath string, _ map[string]any) (domain.ScriptResult, error) {
	f.lastFolder = scriptFolder
	f.lastMainFile = mainFile
	f.lastVenvPath = venvPath
	return f.result, nil
}
