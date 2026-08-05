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
	return s.scripts.Create(ctx, createParams(req))
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
	return s.scripts.Update(ctx, id, sqlite.UpdateScriptParams{
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

	fullPath := filepath.Join(resolveHomePath(script.LocalPath), filepath.FromSlash(filename))
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
	storageDir, err := scriptsStorageDir()
	if err != nil {
		return FinalizeResult{}, err
	}
	finalFolder := filepath.Join(storageDir, finalID)
	venvPath := filepath.Join(finalFolder, "venv")

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.RemoveAll(finalFolder); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.Rename(req.TempFolder, finalFolder); err != nil {
		return FinalizeResult{}, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(finalFolder)
		}
	}()

	if err := s.python.CreateVenv(ctx, venvPath); err != nil {
		return FinalizeResult{}, err
	}

	requirementsPath := findRequirementsFile(finalFolder)
	if requirementsPath != "" {
		if err := s.python.InstallRequirements(ctx, venvPath, requirementsPath); err != nil {
			return FinalizeResult{}, err
		}
	}

	schema, err := s.python.SchemaFile(ctx, finalFolder, req.MainFile, venvPath)
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

	cleanup = false
	return FinalizeResult{Script: script, Schema: schema}, nil
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

func newScriptID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("script-%d", timeNowUnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

func timeNowUnixNano() int64 {
	return time.Now().UTC().UnixNano()
}
