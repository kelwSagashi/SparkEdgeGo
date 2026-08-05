package scripts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidScript = errors.New("invalid script")

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

func NewService(scripts Repository) *Service {
	return &Service{scripts: scripts}
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
