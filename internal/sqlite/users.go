package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type UsersRepository struct {
	db *gorm.DB
}

type CreateUserParams struct {
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	Role         domain.UserRole
	IsActive     bool
}

type UpdateUserParams struct {
	ID           string
	Email        *string
	FirstName    *string
	LastName     *string
	PasswordHash *string
	Role         *domain.UserRole
	IsActive     *bool
}

func NewUsersRepository(db *gorm.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Create(ctx context.Context, params CreateUserParams) (domain.User, error) {
	if params.Role == "" {
		params.Role = domain.RoleViewer
	}

	model := userModel{
		ID:           newID(),
		Email:        params.Email,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		PasswordHash: params.PasswordHash,
		Role:         string(params.Role),
		IsActive:     params.IsActive,
		APIKey:       newAPIKey(),
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.User{}, err
	}
	return userFromModel(model), nil
}

func (r *UsersRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *UsersRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *UsersRepository) FindByAPIKey(ctx context.Context, apiKey string) (domain.User, error) {
	return r.findOne(ctx, "api_key = ?", apiKey)
}

func (r *UsersRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	var models []userModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(models))
	for _, model := range models {
		users = append(users, userFromModel(model))
	}
	return users, nil
}

func (r *UsersRepository) Upsert(ctx context.Context, params UpdateUserParams) (domain.User, error) {
	if params.ID == "" {
		return r.createFromUpdate(ctx, params)
	}

	var model userModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.createFromUpdate(ctx, params)
	}
	if err != nil {
		return domain.User{}, err
	}

	applyUserUpdate(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.User{}, err
	}
	return userFromModel(model), nil
}

func (r *UsersRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&userModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UsersRepository) CreateAPIKey(ctx context.Context, userID string) (string, error) {
	apiKey := newAPIKey()
	result := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", userID).Update("api_key", apiKey)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", ErrNotFound
	}
	return apiKey, nil
}

func (r *UsersRepository) FindProjectUserByName(ctx context.Context, id string, projectName string) (domain.ProjectUser, error) {
	user, err := r.FindByID(ctx, id)
	if err != nil {
		return domain.ProjectUser{}, err
	}

	var project projectModel
	err = r.db.WithContext(ctx).Where("owner_id = ? AND name = ?", id, projectName).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ProjectUser{}, ErrNotFound
	}
	if err != nil {
		return domain.ProjectUser{}, err
	}

	return domain.ProjectUser{
		User:    user,
		Project: projectFromModel(project),
	}, nil
}

func (r *UsersRepository) findOne(ctx context.Context, query string, args ...any) (domain.User, error) {
	var model userModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return userFromModel(model), nil
}

func (r *UsersRepository) createFromUpdate(ctx context.Context, params UpdateUserParams) (domain.User, error) {
	if params.Role == nil {
		role := domain.RoleViewer
		params.Role = &role
	}
	if params.IsActive == nil {
		active := true
		params.IsActive = &active
	}

	model := userModel{
		ID:       params.ID,
		Role:     string(*params.Role),
		IsActive: *params.IsActive,
		APIKey:   newAPIKey(),
	}
	if model.ID == "" {
		model.ID = newID()
	}
	applyUserUpdate(&model, params)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.User{}, err
	}
	return userFromModel(model), nil
}

func applyUserUpdate(model *userModel, params UpdateUserParams) {
	if params.Email != nil {
		model.Email = *params.Email
	}
	if params.FirstName != nil {
		model.FirstName = *params.FirstName
	}
	if params.LastName != nil {
		model.LastName = *params.LastName
	}
	if params.PasswordHash != nil {
		model.PasswordHash = *params.PasswordHash
	}
	if params.Role != nil {
		model.Role = string(*params.Role)
	}
	if params.IsActive != nil {
		model.IsActive = *params.IsActive
	}
}

func userFromModel(model userModel) domain.User {
	return domain.User{
		ID:           model.ID,
		Email:        model.Email,
		FirstName:    model.FirstName,
		LastName:     model.LastName,
		PasswordHash: model.PasswordHash,
		Role:         domain.UserRole(model.Role),
		IsActive:     model.IsActive,
		APIKey:       model.APIKey,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
