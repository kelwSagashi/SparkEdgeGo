package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type ScriptsRepository struct {
	db *gorm.DB
}

type CreateScriptHistoryParams struct {
	ID               string
	ScriptID         string
	Action           string
	Name             string
	Description      string
	Author           string
	Version          string
	MainFile         string
	RequirementsFile string
	Tags             []string
	SchemaConfig     map[string]any
}

type CreateScriptParams struct {
	ID               string
	Name             string
	Description      string
	Author           string
	Version          string
	Source           domain.ScriptSource
	GitHubRepo       string
	GitHubRef        string
	LocalPath        string
	MainFile         string
	VenvPath         string
	RequirementsFile string
	VenvReady        bool
	Language         domain.ScriptLanguage
	Tags             []string
	SchemaConfig     map[string]any
}

type UpdateScriptParams struct {
	Name             *string
	Description      *string
	Author           *string
	Version          *string
	Source           *domain.ScriptSource
	GitHubRepo       *string
	GitHubRef        *string
	LocalPath        *string
	MainFile         *string
	VenvPath         *string
	RequirementsFile *string
	VenvReady        *bool
	Language         *domain.ScriptLanguage
	Tags             *[]string
	SchemaConfig     *map[string]any
}

func NewScriptsRepository(db *gorm.DB) *ScriptsRepository {
	return &ScriptsRepository{db: db}
}

func (r *ScriptsRepository) Create(ctx context.Context, params CreateScriptParams) (domain.DownloadedScript, error) {
	model := scriptModelFromCreate(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.DownloadedScript{}, err
	}
	return scriptFromModel(model), nil
}

func (r *ScriptsRepository) Upsert(ctx context.Context, params CreateScriptParams) (domain.DownloadedScript, error) {
	if params.ID == "" {
		return r.Create(ctx, params)
	}

	var model downloadedScriptModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.Create(ctx, params)
	}
	if err != nil {
		return domain.DownloadedScript{}, err
	}

	applyScriptCreate(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.DownloadedScript{}, err
	}
	return scriptFromModel(model), nil
}

func (r *ScriptsRepository) FindByID(ctx context.Context, id string) (domain.DownloadedScript, error) {
	var model downloadedScriptModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DownloadedScript{}, ErrNotFound
		}
		return domain.DownloadedScript{}, err
	}
	return scriptFromModel(model), nil
}

func (r *ScriptsRepository) ListAll(ctx context.Context) ([]domain.DownloadedScript, error) {
	var models []downloadedScriptModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	scripts := make([]domain.DownloadedScript, 0, len(models))
	for _, model := range models {
		scripts = append(scripts, scriptFromModel(model))
	}
	return scripts, nil
}

func (r *ScriptsRepository) MarkVenvReady(ctx context.Context, id string, venvPath string) (domain.DownloadedScript, error) {
	ready := true
	return r.Update(ctx, id, UpdateScriptParams{VenvReady: &ready, VenvPath: &venvPath})
}

func (r *ScriptsRepository) Update(ctx context.Context, id string, params UpdateScriptParams) (domain.DownloadedScript, error) {
	var model downloadedScriptModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DownloadedScript{}, ErrNotFound
		}
		return domain.DownloadedScript{}, err
	}

	applyScriptUpdate(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.DownloadedScript{}, err
	}
	return scriptFromModel(model), nil
}

func (r *ScriptsRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&downloadedScriptModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ScriptsRepository) CreateHistoryEntry(ctx context.Context, params CreateScriptHistoryParams) (domain.ScriptHistoryEntry, error) {
	model := downloadedScriptHistoryModel{
		ID:               params.ID,
		ScriptID:         params.ScriptID,
		Action:           params.Action,
		Name:             params.Name,
		Description:      params.Description,
		Author:           params.Author,
		Version:          params.Version,
		MainFile:         params.MainFile,
		RequirementsFile: params.RequirementsFile,
		Tags:             stringSliceJSON(params.Tags),
		SchemaConfig:     mapJSON(params.SchemaConfig),
	}
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.ScriptHistoryEntry{}, err
	}
	return scriptHistoryFromModel(model), nil
}

func (r *ScriptsRepository) ListHistoryByScriptID(ctx context.Context, scriptID string) ([]domain.ScriptHistoryEntry, error) {
	var models []downloadedScriptHistoryModel
	if err := r.db.WithContext(ctx).
		Where("script_id = ?", scriptID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	history := make([]domain.ScriptHistoryEntry, 0, len(models))
	for _, model := range models {
		history = append(history, scriptHistoryFromModel(model))
	}
	return history, nil
}

func scriptModelFromCreate(params CreateScriptParams) downloadedScriptModel {
	model := downloadedScriptModel{ID: params.ID}
	applyScriptCreate(&model, params)
	return model
}

func applyScriptCreate(model *downloadedScriptModel, params CreateScriptParams) {
	if params.Version == "" {
		params.Version = "1.0.0"
	}
	if params.Source == "" {
		params.Source = domain.ScriptSourceLocal
	}
	if params.Language == "" {
		params.Language = domain.ScriptLanguagePython
	}
	if params.Tags == nil {
		params.Tags = []string{}
	}
	if params.SchemaConfig == nil {
		params.SchemaConfig = map[string]any{}
	}

	model.Name = params.Name
	model.Description = params.Description
	model.Author = params.Author
	model.Version = params.Version
	model.Source = string(params.Source)
	model.GitHubRepo = params.GitHubRepo
	model.GitHubRef = params.GitHubRef
	model.LocalPath = params.LocalPath
	model.MainFile = params.MainFile
	model.VenvPath = params.VenvPath
	model.RequirementsFile = params.RequirementsFile
	model.VenvReady = params.VenvReady
	model.Language = string(params.Language)
	model.Tags = stringSliceJSON(params.Tags)
	model.SchemaConfig = mapJSON(params.SchemaConfig)
}

func applyScriptUpdate(model *downloadedScriptModel, params UpdateScriptParams) {
	if params.Name != nil {
		model.Name = *params.Name
	}
	if params.Description != nil {
		model.Description = *params.Description
	}
	if params.Author != nil {
		model.Author = *params.Author
	}
	if params.Version != nil {
		model.Version = *params.Version
	}
	if params.Source != nil {
		model.Source = string(*params.Source)
	}
	if params.GitHubRepo != nil {
		model.GitHubRepo = *params.GitHubRepo
	}
	if params.GitHubRef != nil {
		model.GitHubRef = *params.GitHubRef
	}
	if params.LocalPath != nil {
		model.LocalPath = *params.LocalPath
	}
	if params.MainFile != nil {
		model.MainFile = *params.MainFile
	}
	if params.VenvPath != nil {
		model.VenvPath = *params.VenvPath
	}
	if params.RequirementsFile != nil {
		model.RequirementsFile = *params.RequirementsFile
	}
	if params.VenvReady != nil {
		model.VenvReady = *params.VenvReady
	}
	if params.Language != nil {
		model.Language = string(*params.Language)
	}
	if params.Tags != nil {
		model.Tags = stringSliceJSON(*params.Tags)
	}
	if params.SchemaConfig != nil {
		model.SchemaConfig = mapJSON(*params.SchemaConfig)
	}
}

func scriptFromModel(model downloadedScriptModel) domain.DownloadedScript {
	return domain.DownloadedScript{
		ID:               model.ID,
		Name:             model.Name,
		Description:      model.Description,
		Author:           model.Author,
		Version:          model.Version,
		Source:           domain.ScriptSource(model.Source),
		GitHubRepo:       model.GitHubRepo,
		GitHubRef:        model.GitHubRef,
		LocalPath:        model.LocalPath,
		MainFile:         model.MainFile,
		VenvPath:         model.VenvPath,
		RequirementsFile: model.RequirementsFile,
		VenvReady:        model.VenvReady,
		Language:         domain.ScriptLanguage(model.Language),
		Tags:             []string(model.Tags),
		SchemaConfig:     map[string]any(model.SchemaConfig),
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}

func scriptHistoryFromModel(model downloadedScriptHistoryModel) domain.ScriptHistoryEntry {
	return domain.ScriptHistoryEntry{
		ID:               model.ID,
		ScriptID:         model.ScriptID,
		Action:           model.Action,
		Name:             model.Name,
		Description:      model.Description,
		Author:           model.Author,
		Version:          model.Version,
		MainFile:         model.MainFile,
		RequirementsFile: model.RequirementsFile,
		Tags:             []string(model.Tags),
		SchemaConfig:     map[string]any(model.SchemaConfig),
		CreatedAt:        model.CreatedAt,
	}
}
