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
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var (
	ErrInvalidScript      = errors.New("invalid script")
	ErrScriptFileNotFound = errors.New("script file not found")
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

const (
	ScriptHistoryActionInstalled       = "installed"
	ScriptHistoryActionBundleUpdated   = "bundle_updated"
	ScriptHistoryActionMetadataUpdated = "metadata_updated"
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
	if err := s.recordHistory(ctx, script, ScriptHistoryActionInstalled); err != nil {
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
	if err := s.recordHistory(ctx, script, ScriptHistoryActionMetadataUpdated); err != nil {
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
		if err := os.Rename(finalFolder, backupFolder); err != nil {
			return FinalizeResult{}, err
		}
	}

	restoreBackup := true
	defer func() {
		if restoreBackup {
			_ = os.RemoveAll(finalFolder)
			if _, err := os.Stat(backupFolder); err == nil {
				_ = os.Rename(backupFolder, finalFolder)
			}
		}
	}()

	if err := os.Rename(replacementFolder, finalFolder); err != nil {
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
	if err := s.recordHistory(ctx, script, ScriptHistoryActionBundleUpdated); err != nil {
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
		home, err := os.UserHomeDir()
		if err != nil {
			return pathValue
		}
		return filepath.Join(home, pathValue)
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
	home, err := os.UserHomeDir()
	if err != nil {
		return pathValue
	}
	rel, err := filepath.Rel(home, pathValue)
	if err != nil || strings.HasPrefix(rel, "..") {
		return pathValue
	}
	return rel
}

func scriptsStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".spark_edge", "scripts"), nil
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
	if err := os.Rename(tempFolder, finalFolder); err != nil {
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

func (s *Service) recordHistory(ctx context.Context, script domain.DownloadedScript, action string) error {
	_, err := s.scripts.CreateHistoryEntry(ctx, sqlite.CreateScriptHistoryParams{
		ScriptID:         script.ID,
		Action:           action,
		Name:             script.Name,
		Description:      script.Description,
		Author:           script.Author,
		Version:          script.Version,
		MainFile:         script.MainFile,
		RequirementsFile: script.RequirementsFile,
		Tags:             script.Tags,
		SchemaConfig:     script.SchemaConfig,
	})
	return err
}

func timeNowUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}
