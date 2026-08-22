package scripts

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var (
	ErrInvalidScript      = errors.New("invalid script")
	ErrScriptFileNotFound = errors.New("script file not found")
	renamePath            = os.Rename
)

type Repository interface {
	Create(ctx context.Context, params sqlite.CreateScriptParams) (domain.DownloadedScript, error)
	Upsert(ctx context.Context, params sqlite.CreateScriptParams) (domain.DownloadedScript, error)
	FindByID(ctx context.Context, id string) (domain.DownloadedScript, error)
	ListAll(ctx context.Context) ([]domain.DownloadedScript, error)
	Update(ctx context.Context, id string, params sqlite.UpdateScriptParams) (domain.DownloadedScript, error)
	Delete(ctx context.Context, id string) error
	CreateHistoryEntry(ctx context.Context, params sqlite.CreateScriptHistoryParams) (domain.ScriptHistoryEntry, error)
	ListHistoryByScriptID(ctx context.Context, scriptID string) ([]domain.ScriptHistoryEntry, error)
	FindHistoryEntryByID(ctx context.Context, scriptID string, historyID string) (domain.ScriptHistoryEntry, error)
}

type Service struct {
	scripts Repository
	python  PythonRuntime
}

type PythonRuntime interface {
	CreateVenv(ctx context.Context, venvPath string) error
	InstallRequirements(ctx context.Context, venvPath string, requirementsPath string) error
	SchemaFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string) (map[string]any, error)
	RunFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string, input map[string]any) (domain.ScriptResult, error)
	ReadmeFile(ctx context.Context, scriptFolder string, mainFile string, venvPath string) (string, error)
}

type CreateRequest struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Author           string                `json:"author"`
	Version          string                `json:"version"`
	Source           domain.ScriptSource   `json:"source"`
	GitHubRepo       string                `json:"github_repo"`
	GitHubRef        string                `json:"github_ref"`
	LocalPath        string                `json:"local_path"`
	MainFile         string                `json:"main_file"`
	VenvPath         string                `json:"venv_path"`
	RequirementsFile string                `json:"requirements_file"`
	VenvReady        bool                  `json:"venv_ready"`
	Language         domain.ScriptLanguage `json:"language"`
	Tags             []string              `json:"tags"`
	SchemaConfig     map[string]any        `json:"schema_config"`
}

type UpdateRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	Author           *string                `json:"author"`
	Version          *string                `json:"version"`
	Source           *domain.ScriptSource   `json:"source"`
	GitHubRepo       *string                `json:"github_repo"`
	GitHubRef        *string                `json:"github_ref"`
	LocalPath        *string                `json:"local_path"`
	MainFile         *string                `json:"main_file"`
	VenvPath         *string                `json:"venv_path"`
	RequirementsFile *string                `json:"requirements_file"`
	VenvReady        *bool                  `json:"venv_ready"`
	Language         *domain.ScriptLanguage `json:"language"`
	Tags             *[]string              `json:"tags"`
	SchemaConfig     *map[string]any        `json:"schema_config"`
}

type InspectResult struct {
	TempFolder string   `json:"tempFolder"`
	PyFiles    []string `json:"pyFiles"`
	HasSparkit bool     `json:"hasSparkit"`
}

type FinalizeRequest struct {
	TempFolder  string         `json:"tempFolder"`
	MainFile    string         `json:"mainFile"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tags        []string       `json:"tags"`
	Author      string         `json:"author"`
	Version     string         `json:"version"`
	Schema      map[string]any `json:"-"`
}

type FinalizeResult struct {
	Script domain.DownloadedScript `json:"script"`
	Schema map[string]any          `json:"schema"`
}

type PlaygroundRequest struct {
	ScriptID   string         `json:"script_id"`
	SampleName string         `json:"sample_name"`
	Inputs     map[string]any `json:"inputs"`
}

type DraftFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ScriptFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type DraftRequest struct {
	Files       []DraftFile    `json:"files"`
	MainFile    string         `json:"mainFile"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tags        []string       `json:"tags"`
	Author      string         `json:"author"`
	Version     string         `json:"version"`
	Inputs      map[string]any `json:"inputs"`
}

const (
	ScriptHistoryActionInstalled       = "installed"
	ScriptHistoryActionBundleUpdated   = "bundle_updated"
	ScriptHistoryActionMetadataUpdated = "metadata_updated"
	ScriptHistoryActionRestored        = "restored"
)

func NewService(scripts Repository, python ...PythonRuntime) *Service {
	service := &Service{scripts: scripts}
	if len(python) > 0 {
		service.python = python[0]
	}
	return service
}

func (s *Service) ListAll(ctx context.Context) ([]domain.DownloadedScript, error) {
	return s.scripts.ListAll(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.DownloadedScript, error) {
	if strings.TrimSpace(id) == "" {
		return domain.DownloadedScript{}, ErrInvalidScript
	}
	return s.scripts.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.DownloadedScript, error) {
	if err := validateCreate(req); err != nil {
		return domain.DownloadedScript{}, err
	}
	script, err := s.scripts.Create(ctx, createParams(req))
	if err != nil {
		return domain.DownloadedScript{}, err
	}
	if err := s.recordHistory(ctx, domain.DownloadedScript{}, script, ScriptHistoryActionInstalled, "", resolveHomePath(script.LocalPath), nil); err != nil {
		return domain.DownloadedScript{}, err
	}
	return script, nil
}

func (s *Service) Upsert(ctx context.Context, req CreateRequest) (domain.DownloadedScript, error) {
	if err := validateCreate(req); err != nil {
		return domain.DownloadedScript{}, err
	}
	return s.scripts.Upsert(ctx, createParams(req))
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (domain.DownloadedScript, error) {
	if strings.TrimSpace(id) == "" {
		return domain.DownloadedScript{}, ErrInvalidScript
	}
	previous, err := s.FindByID(ctx, id)
	if err != nil {
		return domain.DownloadedScript{}, err
	}
	script, err := s.scripts.Update(ctx, id, sqlite.UpdateScriptParams{
		Name:             req.Name,
		Description:      req.Description,
		Author:           req.Author,
		Version:          req.Version,
		Source:           req.Source,
		GitHubRepo:       req.GitHubRepo,
		GitHubRef:        req.GitHubRef,
		LocalPath:        req.LocalPath,
		MainFile:         req.MainFile,
		VenvPath:         req.VenvPath,
		RequirementsFile: req.RequirementsFile,
		VenvReady:        req.VenvReady,
		Language:         req.Language,
		Tags:             req.Tags,
		SchemaConfig:     req.SchemaConfig,
	})
	if err != nil {
		return domain.DownloadedScript{}, err
	}
	if err := s.recordHistory(ctx, previous, script, ScriptHistoryActionMetadataUpdated, "", resolveHomePath(script.LocalPath), nil); err != nil {
		return domain.DownloadedScript{}, err
	}
	return script, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidScript
	}

	script, err := s.scripts.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if script.LocalPath != "" {
		_ = os.RemoveAll(resolveHomePath(script.LocalPath))
	}
	return s.scripts.Delete(ctx, id)
}

func (s *Service) FileContent(ctx context.Context, id string, filename string) (string, error) {
	if strings.TrimSpace(filename) == "" || strings.Contains(filename, "..") || filepath.IsAbs(filename) {
		return "", ErrInvalidScript
	}

	script, err := s.FindByID(ctx, id)
	if err != nil {
		return "", err
	}

	fullPath, err := resolveScriptFilePath(resolveHomePath(script.LocalPath), filename)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(fullPath)
	if os.IsNotExist(err) {
		return "", ErrScriptFileNotFound
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Service) Files(ctx context.Context, id string) ([]ScriptFile, error) {
	script, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	root := resolveHomePath(script.LocalPath)
	files := make([]ScriptFile, 0)
	if err := filepath.WalkDir(root, func(pathValue string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if pathValue == root {
			return nil
		}

		name := entry.Name()
		if entry.IsDir() {
			if shouldSkipScriptEditorDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipScriptEditorFile(name) {
			return nil
		}

		rel, err := filepath.Rel(root, pathValue)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(pathValue)
		if err != nil {
			return err
		}
		if !isEditableScriptContent(content) {
			return nil
		}
		files = append(files, ScriptFile{
			Path:    filepath.ToSlash(rel),
			Content: string(content),
		})
		return nil
	}); err != nil {
		return nil, err
	}

	slices.SortFunc(files, func(a ScriptFile, b ScriptFile) int {
		return strings.Compare(a.Path, b.Path)
	})
	return files, nil
}

func (s *Service) InspectZip(zipFilePath string) (InspectResult, error) {
	tempFolder, err := os.MkdirTemp("", "spark_edge_script_*")
	if err != nil {
		return InspectResult{}, err
	}

	result, err := extractZipToTemp(zipFilePath, tempFolder)
	if err != nil {
		_ = os.RemoveAll(tempFolder)
		return InspectResult{}, err
	}
	return result, nil
}

func (s *Service) FinalizeDraft(ctx context.Context, req DraftRequest) (FinalizeResult, error) {
	if s.python == nil {
		return FinalizeResult{}, errors.New("sparkit runtime unavailable")
	}
	tempFolder, err := writeDraftToTemp(req)
	if err != nil {
		return FinalizeResult{}, err
	}
	result, err := s.FinalizeUpload(ctx, FinalizeRequest{
		TempFolder:  tempFolder,
		MainFile:    req.MainFile,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Author:      req.Author,
		Version:     req.Version,
	})
	if err != nil {
		_ = os.RemoveAll(tempFolder)
		return FinalizeResult{}, err
	}
	return result, nil
}

func (s *Service) ReplaceDraft(ctx context.Context, scriptID string, req DraftRequest) (FinalizeResult, error) {
	if s.python == nil {
		return FinalizeResult{}, errors.New("sparkit runtime unavailable")
	}
	tempFolder, err := writeDraftToTemp(req)
	if err != nil {
		return FinalizeResult{}, err
	}
	result, err := s.ReplaceUpload(ctx, scriptID, FinalizeRequest{
		TempFolder:  tempFolder,
		MainFile:    req.MainFile,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Author:      req.Author,
		Version:     req.Version,
	})
	if err != nil {
		_ = os.RemoveAll(tempFolder)
		return FinalizeResult{}, err
	}
	return result, nil
}

func (s *Service) RunDraftPlayground(ctx context.Context, req DraftRequest) (domain.ScriptResult, error) {
	if s.python == nil {
		return domain.ScriptResult{}, errors.New("sparkit runtime unavailable")
	}
	if req.Inputs == nil {
		req.Inputs = map[string]any{}
	}

	tempFolder, err := writeDraftToTemp(req)
	if err != nil {
		return domain.ScriptResult{}, err
	}
	defer os.RemoveAll(tempFolder)

	venvPath, err := s.prepareDraftRuntime(ctx, tempFolder)
	if err != nil {
		return domain.ScriptResult{}, err
	}
	return s.python.RunFile(ctx, tempFolder, req.MainFile, venvPath, req.Inputs)
}

func (s *Service) GenerateDraftReadme(ctx context.Context, req DraftRequest) (string, error) {
	if s.python == nil {
		return "", errors.New("sparkit runtime unavailable")
	}

	tempFolder, err := writeDraftToTemp(req)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempFolder)

	venvPath, err := s.prepareDraftRuntime(ctx, tempFolder)
	if err != nil {
		return "", err
	}
	return s.python.ReadmeFile(ctx, tempFolder, req.MainFile, venvPath)
}

func (s *Service) FinalizeUpload(ctx context.Context, req FinalizeRequest) (FinalizeResult, error) {
	if s.python == nil {
		return FinalizeResult{}, errors.New("sparkit runtime unavailable")
	}
	if strings.TrimSpace(req.TempFolder) == "" || strings.TrimSpace(req.MainFile) == "" {
		return FinalizeResult{}, ErrInvalidScript
	}

	if req.Name == "" {
		req.Name = "Unnamed Script"
	}
	if req.Author == "" {
		req.Author = "unknown"
	}
	if req.Version == "" {
		req.Version = "1.0.0"
	}

	finalID := newScriptID()
	finalFolder, venvPath, _, schema, err := s.prepareScriptFolder(ctx, req.TempFolder, finalID, req.MainFile)
	if err != nil {
		return FinalizeResult{}, err
	}

	script, err := s.Create(ctx, CreateRequest{
		ID:           finalID,
		Name:         req.Name,
		Description:  req.Description,
		Author:       req.Author,
		Version:      req.Version,
		Source:       domain.ScriptSourceLocal,
		LocalPath:    toRelativePath(finalFolder),
		MainFile:     req.MainFile,
		VenvPath:     toRelativePath(venvPath),
		VenvReady:    true,
		Tags:         req.Tags,
		SchemaConfig: schema,
	})
	if err != nil {
		return FinalizeResult{}, err
	}

	return FinalizeResult{Script: script, Schema: schema}, nil
}

func (s *Service) ReplaceUpload(ctx context.Context, scriptID string, req FinalizeRequest) (FinalizeResult, error) {
	return s.replaceUpload(ctx, scriptID, req, ScriptHistoryActionBundleUpdated, nil)
}

func (s *Service) replaceUpload(ctx context.Context, scriptID string, req FinalizeRequest, historyAction string, explicitSummary []string) (FinalizeResult, error) {
	if s.python == nil {
		return FinalizeResult{}, errors.New("sparkit runtime unavailable")
	}
	if strings.TrimSpace(scriptID) == "" || strings.TrimSpace(req.TempFolder) == "" || strings.TrimSpace(req.MainFile) == "" {
		return FinalizeResult{}, ErrInvalidScript
	}

	existing, err := s.FindByID(ctx, scriptID)
	if err != nil {
		return FinalizeResult{}, err
	}

	if req.Name == "" {
		req.Name = existing.Name
	}
	if req.Description == "" {
		req.Description = existing.Description
	}
	if len(req.Tags) == 0 {
		req.Tags = existing.Tags
	}
	if req.Author == "" {
		req.Author = existing.Author
	}
	if req.Version == "" {
		req.Version = existing.Version
	}
	previousFolder := resolveHomePath(existing.LocalPath)

	replacementID := scriptID + "_replacement"
	replacementFolder, _, requirementsPath, schema, err := s.prepareScriptFolder(ctx, req.TempFolder, replacementID, req.MainFile)
	if err != nil {
		return FinalizeResult{}, err
	}
	defer os.RemoveAll(replacementFolder)

	finalFolder := resolveHomePath(existing.LocalPath)
	backupFolder := finalFolder + "_backup"
	_ = os.RemoveAll(backupFolder)

	if _, err := os.Stat(finalFolder); err == nil {
		if err := movePath(finalFolder, backupFolder); err != nil {
			return FinalizeResult{}, err
		}
	}

	restoreBackup := true
	defer func() {
		if restoreBackup {
			_ = os.RemoveAll(finalFolder)
			if _, err := os.Stat(backupFolder); err == nil {
				_ = movePath(backupFolder, finalFolder)
			}
		}
	}()

	if err := movePath(replacementFolder, finalFolder); err != nil {
		return FinalizeResult{}, err
	}

	finalVenv := filepath.Join(finalFolder, "venv")
	requirementsRelative := ""
	if requirementsPath != "" {
		if rel, err := filepath.Rel(replacementFolder, requirementsPath); err == nil {
			requirementsRelative = rel
		}
	}
	if requirementsRelative != "" {
		requirementsRelative = filepath.ToSlash(requirementsRelative)
	}
	mainFile := filepath.ToSlash(req.MainFile)
	localPath := toRelativePath(finalFolder)
	venvPath := toRelativePath(finalVenv)
	requirementsFile := requirementsRelative
	requirementsFilePtr := &requirementsFile
	venvReady := true

	name := req.Name
	description := req.Description
	author := req.Author
	version := req.Version
	tags := req.Tags
	mainFilePtr := &mainFile
	localPathPtr := &localPath
	venvPathPtr := &venvPath

	script, err := s.scripts.Update(ctx, scriptID, sqlite.UpdateScriptParams{
		Name:             &name,
		Description:      &description,
		Author:           &author,
		Version:          &version,
		LocalPath:        localPathPtr,
		MainFile:         mainFilePtr,
		VenvPath:         venvPathPtr,
		RequirementsFile: requirementsFilePtr,
		VenvReady:        &venvReady,
		Tags:             &tags,
		SchemaConfig:     &schema,
	})
	if err != nil {
		return FinalizeResult{}, err
	}
	if err := s.recordHistory(ctx, existing, script, historyAction, previousFolder, finalFolder, explicitSummary); err != nil {
		return FinalizeResult{}, err
	}
	_ = os.RemoveAll(backupFolder)
	restoreBackup = false
	return FinalizeResult{Script: script, Schema: schema}, nil
}

func (s *Service) History(ctx context.Context, scriptID string) ([]domain.ScriptHistoryEntry, error) {
	if strings.TrimSpace(scriptID) == "" {
		return nil, ErrInvalidScript
	}
	return s.scripts.ListHistoryByScriptID(ctx, scriptID)
}

func (s *Service) RestoreHistory(ctx context.Context, scriptID string, historyID string) (FinalizeResult, error) {
	if s.python == nil {
		return FinalizeResult{}, errors.New("sparkit runtime unavailable")
	}
	if strings.TrimSpace(scriptID) == "" || strings.TrimSpace(historyID) == "" {
		return FinalizeResult{}, ErrInvalidScript
	}

	entry, err := s.scripts.FindHistoryEntryByID(ctx, scriptID, historyID)
	if err != nil {
		return FinalizeResult{}, err
	}
	if strings.TrimSpace(entry.BundlePath) == "" {
		return FinalizeResult{}, ErrScriptFileNotFound
	}

	tempFolder, err := os.MkdirTemp("", "spark_edge_script_restore_*")
	if err != nil {
		return FinalizeResult{}, err
	}

	if _, err := extractZipToTemp(resolveHomePath(entry.BundlePath), tempFolder); err != nil {
		_ = os.RemoveAll(tempFolder)
		return FinalizeResult{}, err
	}

	result, err := s.replaceUpload(ctx, scriptID, FinalizeRequest{
		TempFolder:  tempFolder,
		MainFile:    entry.MainFile,
		Name:        entry.Name,
		Description: entry.Description,
		Tags:        entry.Tags,
		Author:      entry.Author,
		Version:     entry.Version,
	}, ScriptHistoryActionRestored, []string{
		fmt.Sprintf("Restaurado a partir do histórico %s", historyID),
		fmt.Sprintf("Versão restaurada: %s", entry.Version),
	})
	return result, err
}

func (s *Service) RunPlayground(ctx context.Context, req PlaygroundRequest) (domain.ScriptResult, error) {
	if s.python == nil {
		return domain.ScriptResult{}, errors.New("sparkit runtime unavailable")
	}
	if req.Inputs == nil {
		req.Inputs = map[string]any{}
	}
	if req.SampleName != "" {
		folder, err := s.sampleFolder(req.SampleName)
		if err != nil {
			return domain.ScriptResult{}, err
		}
		return s.python.RunFile(ctx, folder, "main.py", filepath.Join(folder, "venv"), req.Inputs)
	}
	if req.ScriptID == "" {
		return domain.ScriptResult{}, ErrInvalidScript
	}

	script, err := s.FindByID(ctx, req.ScriptID)
	if err != nil {
		return domain.ScriptResult{}, err
	}

	return s.python.RunFile(ctx, resolveHomePath(script.LocalPath), script.MainFile, resolveHomePath(script.VenvPath), req.Inputs)
}

func (s *Service) ListSamples() ([]string, error) {
	root, err := samplesRoot()
	if err != nil {
		return []string{}, nil
	}

	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}

func (s *Service) SampleSchema(ctx context.Context, sampleName string) (map[string]any, error) {
	if s.python == nil {
		return nil, errors.New("sparkit runtime unavailable")
	}

	folder, err := s.sampleFolder(sampleName)
	if err != nil {
		return nil, err
	}
	return s.python.SchemaFile(ctx, folder, "main.py", filepath.Join(folder, "venv"))
}

func validateCreate(req CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Author) == "" || strings.TrimSpace(req.LocalPath) == "" || strings.TrimSpace(req.MainFile) == "" {
		return ErrInvalidScript
	}
	return nil
}

func createParams(req CreateRequest) sqlite.CreateScriptParams {
	return sqlite.CreateScriptParams{
		ID:               req.ID,
		Name:             req.Name,
		Description:      req.Description,
		Author:           req.Author,
		Version:          req.Version,
		Source:           req.Source,
		GitHubRepo:       req.GitHubRepo,
		GitHubRef:        req.GitHubRef,
		LocalPath:        req.LocalPath,
		MainFile:         req.MainFile,
		VenvPath:         req.VenvPath,
		RequirementsFile: req.RequirementsFile,
		VenvReady:        req.VenvReady,
		Language:         req.Language,
		Tags:             req.Tags,
		SchemaConfig:     req.SchemaConfig,
	}
}

func resolveHomePath(pathValue string) string {
	if strings.HasPrefix(pathValue, ".spark_edge") {
		return appfs.ResolveFromRoot(pathValue)
	}
	if strings.HasPrefix(pathValue, "~/") || pathValue == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return pathValue
		}
		if pathValue == "~" {
			return home
		}
		return filepath.Join(home, strings.TrimPrefix(pathValue, "~/"))
	}
	return pathValue
}

func toRelativePath(pathValue string) string {
	appRoot := appfs.AppRoot()
	rel, err := filepath.Rel(appRoot, pathValue)
	if err != nil || strings.HasPrefix(rel, "..") {
		return pathValue
	}
	return rel
}

func scriptsStorageDir() (string, error) {
	return appfs.ResolveFromRoot(".spark_edge", "scripts"), nil
}

func samplesRoot() (string, error) {
	if configured := os.Getenv("SPARKEDGE_SAMPLES_DIR"); configured != "" {
		return configured, nil
	}

	candidates := []string{
		filepath.Join("extensions", "samples"),
		filepath.Join("..", "extensions", "samples"),
		filepath.Join("..", "SparkEdge", "extensions", "samples"),
	}
	if executable, err := os.Executable(); err == nil {
		base := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(base, "extensions", "samples"),
			filepath.Join(base, "..", "extensions", "samples"),
		)
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", os.ErrNotExist
}

func (s *Service) sampleFolder(sampleName string) (string, error) {
	if strings.TrimSpace(sampleName) == "" || strings.Contains(sampleName, "..") || strings.ContainsAny(sampleName, `/\`) {
		return "", ErrInvalidScript
	}

	root, err := samplesRoot()
	if err != nil {
		return "", ErrScriptFileNotFound
	}

	folder := filepath.Join(root, sampleName)
	info, err := os.Stat(folder)
	if os.IsNotExist(err) {
		return "", ErrScriptFileNotFound
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrInvalidScript
	}
	return folder, nil
}

func extractZipToTemp(zipFilePath string, tempFolder string) (InspectResult, error) {
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return InspectResult{}, err
	}
	defer reader.Close()

	result := InspectResult{TempFolder: tempFolder}
	for _, file := range reader.File {
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		if cleanName == "." || strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return InspectResult{}, fmt.Errorf("unsafe zip path: %s", file.Name)
		}

		targetPath := filepath.Join(tempFolder, cleanName)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return InspectResult{}, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return InspectResult{}, err
		}
		if err := extractZipFile(file, targetPath); err != nil {
			return InspectResult{}, err
		}

		lowerName := strings.ToLower(filepath.Base(cleanName))
		if strings.HasSuffix(lowerName, ".py") {
			result.PyFiles = append(result.PyFiles, filepath.ToSlash(cleanName))
		}
		if lowerName == "requirements.txt" {
			content, err := os.ReadFile(targetPath)
			if err != nil {
				return InspectResult{}, err
			}
			result.HasSparkit = strings.Contains(strings.ToLower(string(content)), "sparkit")
		}
	}

	slices.Sort(result.PyFiles)
	return result, nil
}

func writeDraftToTemp(req DraftRequest) (string, error) {
	if strings.TrimSpace(req.MainFile) == "" {
		return "", ErrInvalidScript
	}

	tempFolder, err := os.MkdirTemp("", "spark_edge_script_draft_*")
	if err != nil {
		return "", err
	}

	hasRequirements := false
	hasReadme := false
	hasMain := false
	for _, file := range req.Files {
		cleanPath, err := cleanDraftPath(file.Path)
		if err != nil {
			_ = os.RemoveAll(tempFolder)
			return "", err
		}
		lowerBase := strings.ToLower(filepath.Base(cleanPath))
		if lowerBase == "requirements.txt" {
			hasRequirements = true
			file.Content = ensureSparkitRequirement(file.Content)
		}
		if lowerBase == "readme.md" {
			hasReadme = true
		}
		if filepath.ToSlash(cleanPath) == filepath.ToSlash(filepath.Clean(filepath.FromSlash(req.MainFile))) {
			hasMain = true
		}
		if err := writeDraftFile(tempFolder, cleanPath, file.Content); err != nil {
			_ = os.RemoveAll(tempFolder)
			return "", err
		}
	}

	mainPath, err := cleanDraftPath(req.MainFile)
	if err != nil {
		_ = os.RemoveAll(tempFolder)
		return "", err
	}
	if !hasMain {
		if _, err := os.Stat(filepath.Join(tempFolder, mainPath)); err != nil {
			_ = os.RemoveAll(tempFolder)
			return "", ErrScriptFileNotFound
		}
	}
	if !hasRequirements {
		if err := writeDraftFile(tempFolder, "requirements.txt", "sparkit\n"); err != nil {
			_ = os.RemoveAll(tempFolder)
			return "", err
		}
	}
	if !hasReadme {
		if err := writeDraftFile(tempFolder, "README.md", "# "+firstNonEmpty(req.Name, "SparkEdge Script")+"\n"); err != nil {
			_ = os.RemoveAll(tempFolder)
			return "", err
		}
	}

	return tempFolder, nil
}

func (s *Service) prepareDraftRuntime(ctx context.Context, tempFolder string) (string, error) {
	venvPath := filepath.Join(tempFolder, "venv")
	if err := s.python.CreateVenv(ctx, venvPath); err != nil {
		return "", err
	}
	if requirementsPath := findRequirementsFile(tempFolder); requirementsPath != "" {
		if err := s.python.InstallRequirements(ctx, venvPath, requirementsPath); err != nil {
			return "", err
		}
	}
	return venvPath, nil
}

func cleanDraftPath(pathValue string) (string, error) {
	cleanPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(pathValue)))
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return "", ErrInvalidScript
	}
	return cleanPath, nil
}

func writeDraftFile(root string, relativePath string, content string) error {
	targetPath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, []byte(content), 0o644)
}

func shouldSkipScriptEditorDir(name string) bool {
	lowerName := strings.ToLower(name)
	return lowerName == "venv" ||
		lowerName == ".venv" ||
		lowerName == "__pycache__" ||
		lowerName == ".git" ||
		lowerName == ".pytest_cache" ||
		lowerName == ".mypy_cache"
}

func shouldSkipScriptEditorFile(name string) bool {
	lowerName := strings.ToLower(name)
	if strings.HasSuffix(lowerName, ".pyc") ||
		strings.HasSuffix(lowerName, ".pyo") ||
		strings.HasSuffix(lowerName, ".zip") ||
		strings.HasSuffix(lowerName, ".tar") ||
		strings.HasSuffix(lowerName, ".gz") ||
		strings.HasSuffix(lowerName, ".7z") {
		return true
	}
	return false
}

func isEditableScriptContent(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	if len(content) > 512*1024 {
		return false
	}
	for _, value := range content {
		if value == 0 {
			return false
		}
	}
	return true
}

func ensureSparkitRequirement(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if trimmed == "sparkit" || strings.HasPrefix(trimmed, "sparkit==") || strings.HasPrefix(trimmed, "sparkit>=") {
			return content
		}
	}
	trimmedContent := strings.TrimRight(content, "\r\n")
	if trimmedContent == "" {
		return "sparkit\n"
	}
	return trimmedContent + "\nsparkit\n"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func resolveScriptFilePath(root string, filename string) (string, error) {
	directPath := filepath.Join(root, filepath.FromSlash(filename))
	if _, err := os.Stat(directPath); err == nil {
		return directPath, nil
	}

	if strings.ContainsAny(filename, `/\`) {
		return "", ErrScriptFileNotFound
	}

	var resolved string
	_ = filepath.WalkDir(root, func(pathValue string, entry os.DirEntry, err error) error {
		if err != nil || resolved != "" {
			return nil
		}
		if entry.IsDir() {
			if strings.EqualFold(entry.Name(), "venv") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(entry.Name(), filename) {
			resolved = pathValue
		}
		return nil
	})
	if resolved == "" {
		return "", ErrScriptFileNotFound
	}
	return resolved, nil
}

func extractZipFile(file *zip.File, targetPath string) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

func findRequirementsFile(root string) string {
	var result string
	_ = filepath.WalkDir(root, func(pathValue string, entry os.DirEntry, err error) error {
		if err != nil || result != "" {
			return nil
		}
		if entry.IsDir() && entry.Name() == "venv" {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.EqualFold(entry.Name(), "requirements.txt") {
			result = pathValue
		}
		return nil
	})
	return result
}

func listBundleFiles(root string) ([]string, error) {
	if root == "" {
		return nil, nil
	}
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	files := make([]string, 0)
	err = filepath.WalkDir(root, func(pathValue string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if strings.EqualFold(entry.Name(), "venv") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, pathValue)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}

func filesDiffSummary(previousFiles []string, nextFiles []string) []string {
	prevSet := make(map[string]struct{}, len(previousFiles))
	nextSet := make(map[string]struct{}, len(nextFiles))
	for _, item := range previousFiles {
		prevSet[item] = struct{}{}
	}
	for _, item := range nextFiles {
		nextSet[item] = struct{}{}
	}

	added := make([]string, 0)
	removed := make([]string, 0)
	for _, item := range nextFiles {
		if _, ok := prevSet[item]; !ok {
			added = append(added, item)
		}
	}
	for _, item := range previousFiles {
		if _, ok := nextSet[item]; !ok {
			removed = append(removed, item)
		}
	}

	summary := make([]string, 0)
	if len(added) > 0 {
		summary = append(summary, summarizeFiles("Arquivos adicionados", added))
	}
	if len(removed) > 0 {
		summary = append(summary, summarizeFiles("Arquivos removidos", removed))
	}
	if len(summary) == 0 && len(nextFiles) > 0 {
		summary = append(summary, fmt.Sprintf("Bundle mantido com %d arquivo(s)", len(nextFiles)))
	}
	return summary
}

func summarizeFiles(prefix string, files []string) string {
	if len(files) == 0 {
		return prefix
	}
	limit := minInt(len(files), 3)
	snippet := strings.Join(files[:limit], ", ")
	if len(files) > limit {
		return fmt.Sprintf("%s (%d): %s...", prefix, len(files), snippet)
	}
	return fmt.Sprintf("%s (%d): %s", prefix, len(files), snippet)
}

func metadataDiffSummary(previous domain.DownloadedScript, current domain.DownloadedScript) []string {
	summary := make([]string, 0)
	if previous.Name != current.Name {
		summary = append(summary, fmt.Sprintf("Nome: %q -> %q", previous.Name, current.Name))
	}
	if previous.Version != current.Version {
		summary = append(summary, fmt.Sprintf("Versão: %q -> %q", previous.Version, current.Version))
	}
	if previous.Author != current.Author {
		summary = append(summary, fmt.Sprintf("Autor: %q -> %q", previous.Author, current.Author))
	}
	if previous.Description != current.Description {
		summary = append(summary, "Descrição atualizada")
	}
	if previous.MainFile != current.MainFile {
		summary = append(summary, fmt.Sprintf("Entrypoint: %q -> %q", previous.MainFile, current.MainFile))
	}
	if previous.RequirementsFile != current.RequirementsFile {
		summary = append(summary, fmt.Sprintf("Requirements: %q -> %q", previous.RequirementsFile, current.RequirementsFile))
	}
	if !stringSlicesEqual(previous.Tags, current.Tags) {
		summary = append(summary, fmt.Sprintf("Tags: [%s] -> [%s]", strings.Join(previous.Tags, ", "), strings.Join(current.Tags, ", ")))
	}
	return summary
}

func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func historyStorageDir() (string, error) {
	return appfs.ResolveFromRoot(".spark_edge", "script_history"), nil
}

func createBundleArchive(scriptID string, historyID string, scriptFolder string) (string, error) {
	if strings.TrimSpace(scriptFolder) == "" {
		return "", nil
	}
	info, err := os.Stat(scriptFolder)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", nil
	}

	historyRoot, err := historyStorageDir()
	if err != nil {
		return "", err
	}
	targetDir := filepath.Join(historyRoot, scriptID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	archivePath := filepath.Join(targetDir, historyID+".zip")
	if err := zipDirectory(scriptFolder, archivePath); err != nil {
		return "", err
	}
	return toRelativePath(archivePath), nil
}

func zipDirectory(root string, archivePath string) error {
	target, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer target.Close()

	writer := zip.NewWriter(target)
	defer writer.Close()

	return filepath.WalkDir(root, func(pathValue string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if strings.EqualFold(entry.Name(), "venv") {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, pathValue)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate

		fileWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		source, err := os.Open(pathValue)
		if err != nil {
			return err
		}
		_, err = io.Copy(fileWriter, source)
		closeErr := source.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
}

func (s *Service) prepareScriptFolder(ctx context.Context, tempFolder string, folderID string, mainFile string) (string, string, string, map[string]any, error) {
	storageDir, err := scriptsStorageDir()
	if err != nil {
		return "", "", "", nil, err
	}
	finalFolder := filepath.Join(storageDir, folderID)
	venvPath := filepath.Join(finalFolder, "venv")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return "", "", "", nil, err
	}
	if err := os.RemoveAll(finalFolder); err != nil {
		return "", "", "", nil, err
	}
	if err := movePath(tempFolder, finalFolder); err != nil {
		return "", "", "", nil, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(finalFolder)
		}
	}()

	if err := s.python.CreateVenv(ctx, venvPath); err != nil {
		return "", "", "", nil, err
	}

	requirementsPath := findRequirementsFile(finalFolder)
	if requirementsPath != "" {
		if err := s.python.InstallRequirements(ctx, venvPath, requirementsPath); err != nil {
			return "", "", "", nil, err
		}
	}

	schema, err := s.python.SchemaFile(ctx, finalFolder, mainFile, venvPath)
	if err != nil {
		return "", "", "", nil, err
	}

	cleanup = false
	return finalFolder, venvPath, requirementsPath, schema, nil
}

func newScriptID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("script-%d", timeNowUnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

func (s *Service) recordHistory(ctx context.Context, previous domain.DownloadedScript, current domain.DownloadedScript, action string, previousFolder string, currentFolder string, explicitSummary []string) error {
	historyID := newScriptID()
	summary := make([]string, 0)

	if explicitSummary != nil {
		summary = append(summary, explicitSummary...)
	} else {
		summary = append(summary, metadataDiffSummary(previous, current)...)
		previousFiles, err := listBundleFiles(previousFolder)
		if err != nil {
			return err
		}
		nextFiles, err := listBundleFiles(currentFolder)
		if err != nil {
			return err
		}
		switch action {
		case ScriptHistoryActionInstalled:
			if len(nextFiles) > 0 {
				summary = append(summary, fmt.Sprintf("Bundle instalado com %d arquivo(s)", len(nextFiles)))
			}
		case ScriptHistoryActionBundleUpdated:
			summary = append(summary, filesDiffSummary(previousFiles, nextFiles)...)
		}
	}

	if len(summary) == 0 {
		summary = append(summary, "Sem diferenças resumidas detectadas")
	}

	archivePath, err := createBundleArchive(current.ID, historyID, resolveHomePath(current.LocalPath))
	if err != nil {
		return err
	}

	_, err = s.scripts.CreateHistoryEntry(ctx, sqlite.CreateScriptHistoryParams{
		ID:               historyID,
		ScriptID:         current.ID,
		Action:           action,
		Name:             current.Name,
		Description:      current.Description,
		Author:           current.Author,
		Version:          current.Version,
		MainFile:         current.MainFile,
		RequirementsFile: current.RequirementsFile,
		Tags:             current.Tags,
		SchemaConfig:     current.SchemaConfig,
		ChangeSummary:    summary,
		BundlePath:       archivePath,
	})
	return err
}

func movePath(source string, target string) error {
	if err := renamePath(source, target); err == nil {
		return nil
	} else if !isCrossDeviceError(err) {
		return err
	}

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := copyDirectory(source, target); err != nil {
			return err
		}
	} else {
		if err := copyFile(source, target, info.Mode()); err != nil {
			return err
		}
	}

	return os.RemoveAll(source)
}

func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV) || strings.Contains(strings.ToLower(err.Error()), "cross-device link")
}

func copyDirectory(source string, target string) error {
	return filepath.Walk(source, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}

		if relativePath == "." {
			return os.MkdirAll(target, info.Mode())
		}

		destination := filepath.Join(target, relativePath)
		if info.IsDir() {
			return os.MkdirAll(destination, info.Mode())
		}
		return copyFile(current, destination, info.Mode())
	})
}

func copyFile(source string, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func timeNowUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
