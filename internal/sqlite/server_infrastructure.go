package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type ServerTypesRepository struct {
	db *gorm.DB
}

type AuthTypesRepository struct {
	db *gorm.DB
}

type CredentialsRepository struct {
	db *gorm.DB
}

type ServersRepository struct {
	db *gorm.DB
}

type ServerResourcesRepository struct {
	db *gorm.DB
}

type ResourceOperationsRepository struct {
	db *gorm.DB
}

type OperationTarget struct {
	Operation  domain.ResourceOperation
	Resource   domain.ServerResource
	Server     domain.Server
	Credential *domain.Credential
}

type UpsertServerTypeParams struct {
	ID          string
	Key         string
	Name        string
	Description string
}

type UpsertAuthTypeParams struct {
	ID           string
	Name         string
	Strategy     string
	Fields       []domain.AuthTypeField
	ServerTypeID string
}

type UpsertCredentialParams struct {
	ID         string
	Name       string
	AuthTypeID string
	Data       map[string]any
	OwnerID    string
	ProjectID  string
}

type UpsertServerParams struct {
	ID           string
	Name         string
	Type         string
	ServerTypeID string
	DriverKey    string
	CredentialID string
	Headers      map[string]any
	ProjectID    string
	CreatedBy    string
}

type UpsertServerResourceParams struct {
	ID       string
	ServerID string
	Name     string
	Type     string
	Config   map[string]any
}

type UpsertResourceOperationParams struct {
	ID           string
	ResourceID   string
	Name         string
	Type         string
	Config       map[string]any
	InputSchema  map[string]any
	OutputSchema map[string]any
}

func NewServerTypesRepository(db *gorm.DB) *ServerTypesRepository {
	return &ServerTypesRepository{db: db}
}

func NewAuthTypesRepository(db *gorm.DB) *AuthTypesRepository {
	return &AuthTypesRepository{db: db}
}

func NewCredentialsRepository(db *gorm.DB) *CredentialsRepository {
	return &CredentialsRepository{db: db}
}

func NewServersRepository(db *gorm.DB) *ServersRepository {
	return &ServersRepository{db: db}
}

func NewServerResourcesRepository(db *gorm.DB) *ServerResourcesRepository {
	return &ServerResourcesRepository{db: db}
}

func NewResourceOperationsRepository(db *gorm.DB) *ResourceOperationsRepository {
	return &ResourceOperationsRepository{db: db}
}

func (r *ServerTypesRepository) Upsert(ctx context.Context, params UpsertServerTypeParams) (domain.ServerType, error) {
	model := serverTypeModel{ID: params.ID, Key: params.Key, Name: params.Name, Description: params.Description}
	if model.ID == "" {
		model.ID = newID()
	}
	return saveServerType(ctx, r.db, model)
}

func (r *ServerTypesRepository) ListAll(ctx context.Context) ([]domain.ServerType, error) {
	var models []serverTypeModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ServerType, 0, len(models))
	for _, model := range models {
		result = append(result, serverTypeFromModel(model))
	}
	return result, nil
}

func (r *AuthTypesRepository) Upsert(ctx context.Context, params UpsertAuthTypeParams) (domain.AuthType, error) {
	model := authTypeModel{ID: params.ID, Name: params.Name, Strategy: params.Strategy, Fields: authTypeFieldsJSON(params.Fields), ServerTypeID: params.ServerTypeID}
	if model.ID == "" {
		model.ID = newID()
	}
	return saveAuthType(ctx, r.db, model)
}

func (r *AuthTypesRepository) ListAll(ctx context.Context) ([]domain.AuthType, error) {
	var models []authTypeModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.AuthType, 0, len(models))
	for _, model := range models {
		result = append(result, authTypeFromModel(model))
	}
	return result, nil
}

func (r *AuthTypesRepository) ListByServerType(ctx context.Context, serverTypeID string) ([]domain.AuthType, error) {
	var models []authTypeModel
	query := r.db.WithContext(ctx).Order("name ASC")
	if serverTypeID != "" {
		query = query.Where("server_type_id = ?", serverTypeID)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.AuthType, 0, len(models))
	for _, model := range models {
		result = append(result, authTypeFromModel(model))
	}
	return result, nil
}

func (r *CredentialsRepository) Upsert(ctx context.Context, params UpsertCredentialParams) (domain.Credential, error) {
	model := credentialModel{ID: params.ID, Name: params.Name, AuthTypeID: params.AuthTypeID, Data: mapJSON(params.Data), OwnerID: params.OwnerID, ProjectID: params.ProjectID}
	if model.ID == "" {
		model.ID = newID()
	}
	if model.Data == nil {
		model.Data = mapJSON{}
	}
	return saveCredential(ctx, r.db, model)
}

func (r *CredentialsRepository) FindByID(ctx context.Context, id string) (domain.Credential, error) {
	var model credentialModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Credential{}, ErrNotFound
		}
		return domain.Credential{}, err
	}
	return credentialFromModel(model), nil
}

func (r *CredentialsRepository) ListByOwner(ctx context.Context, ownerID string) ([]domain.Credential, error) {
	var models []credentialModel
	query := r.db.WithContext(ctx).Order("name ASC")
	if ownerID != "" {
		query = query.Where("owner_id = ? OR owner_id = ''", ownerID)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Credential, 0, len(models))
	for _, model := range models {
		result = append(result, credentialFromModel(model))
	}
	return result, nil
}

func (r *CredentialsRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&credentialModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ServersRepository) Upsert(ctx context.Context, params UpsertServerParams) (domain.Server, error) {
	model := serverModel{ID: params.ID, Name: params.Name, Type: params.Type, ServerTypeID: params.ServerTypeID, DriverKey: params.DriverKey, CredentialID: params.CredentialID, Headers: mapJSON(params.Headers), ProjectID: params.ProjectID, CreatedBy: params.CreatedBy}
	if model.ID == "" {
		model.ID = newID()
	}
	if model.Headers == nil {
		model.Headers = mapJSON{}
	}
	return saveServer(ctx, r.db, model)
}

func (r *ServersRepository) FindByID(ctx context.Context, id string) (domain.Server, error) {
	var model serverModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Server{}, ErrNotFound
		}
		return domain.Server{}, err
	}
	return serverFromModel(model), nil
}

func (r *ServersRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Server, error) {
	var models []serverModel
	query := r.db.WithContext(ctx).Order("name ASC")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Server, 0, len(models))
	for _, model := range models {
		result = append(result, serverFromModel(model))
	}
	return result, nil
}

func (r *ServersRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&serverModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ServerResourcesRepository) Upsert(ctx context.Context, params UpsertServerResourceParams) (domain.ServerResource, error) {
	model := serverResourceModel{ID: params.ID, ServerID: params.ServerID, Name: params.Name, Type: params.Type, Config: mapJSON(params.Config)}
	if model.ID == "" {
		model.ID = newID()
	}
	if model.Config == nil {
		model.Config = mapJSON{}
	}
	return saveServerResource(ctx, r.db, model)
}

func (r *ServerResourcesRepository) FindByID(ctx context.Context, id string) (domain.ServerResource, error) {
	var model serverResourceModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ServerResource{}, ErrNotFound
		}
		return domain.ServerResource{}, err
	}
	return serverResourceFromModel(model), nil
}

func (r *ServerResourcesRepository) ListByServer(ctx context.Context, serverID string) ([]domain.ServerResource, error) {
	var models []serverResourceModel
	if err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ServerResource, 0, len(models))
	for _, model := range models {
		result = append(result, serverResourceFromModel(model))
	}
	return result, nil
}

func (r *ResourceOperationsRepository) Upsert(ctx context.Context, params UpsertResourceOperationParams) (domain.ResourceOperation, error) {
	model := resourceOperationModel{ID: params.ID, ResourceID: params.ResourceID, Name: params.Name, Type: params.Type, Config: mapJSON(params.Config), InputSchema: mapJSON(params.InputSchema), OutputSchema: mapJSON(params.OutputSchema)}
	if model.ID == "" {
		model.ID = newID()
	}
	return saveResourceOperation(ctx, r.db, model)
}

func (r *ResourceOperationsRepository) FindByID(ctx context.Context, id string) (domain.ResourceOperation, error) {
	var model resourceOperationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ResourceOperation{}, ErrNotFound
		}
		return domain.ResourceOperation{}, err
	}
	return resourceOperationFromModel(model), nil
}

func (r *ResourceOperationsRepository) ListByResource(ctx context.Context, resourceID string) ([]domain.ResourceOperation, error) {
	var models []resourceOperationModel
	if err := r.db.WithContext(ctx).Where("resource_id = ?", resourceID).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.ResourceOperation, 0, len(models))
	for _, model := range models {
		result = append(result, resourceOperationFromModel(model))
	}
	return result, nil
}

func (r *ResourceOperationsRepository) ResolveTarget(ctx context.Context, id string) (OperationTarget, error) {
	operation, err := r.FindByID(ctx, id)
	if err != nil {
		return OperationTarget{}, err
	}
	resources := NewServerResourcesRepository(r.db)
	resource, err := resources.FindByID(ctx, operation.ResourceID)
	if err != nil {
		return OperationTarget{}, err
	}
	servers := NewServersRepository(r.db)
	server, err := servers.FindByID(ctx, resource.ServerID)
	if err != nil {
		return OperationTarget{}, err
	}
	target := OperationTarget{Operation: operation, Resource: resource, Server: server}
	if server.CredentialID != "" {
		credentials := NewCredentialsRepository(r.db)
		credential, err := credentials.FindByID(ctx, server.CredentialID)
		if err != nil {
			return OperationTarget{}, err
		}
		target.Credential = &credential
	}
	return target, nil
}

func saveServerType(ctx context.Context, db *gorm.DB, model serverTypeModel) (domain.ServerType, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.ServerType{}, err
	}
	return serverTypeFromModel(model), nil
}

func saveAuthType(ctx context.Context, db *gorm.DB, model authTypeModel) (domain.AuthType, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.AuthType{}, err
	}
	return authTypeFromModel(model), nil
}

func saveCredential(ctx context.Context, db *gorm.DB, model credentialModel) (domain.Credential, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Credential{}, err
	}
	return credentialFromModel(model), nil
}

func saveServer(ctx context.Context, db *gorm.DB, model serverModel) (domain.Server, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Server{}, err
	}
	return serverFromModel(model), nil
}

func saveServerResource(ctx context.Context, db *gorm.DB, model serverResourceModel) (domain.ServerResource, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.ServerResource{}, err
	}
	return serverResourceFromModel(model), nil
}

func saveResourceOperation(ctx context.Context, db *gorm.DB, model resourceOperationModel) (domain.ResourceOperation, error) {
	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.ResourceOperation{}, err
	}
	return resourceOperationFromModel(model), nil
}

func serverTypeFromModel(model serverTypeModel) domain.ServerType {
	return domain.ServerType{ID: model.ID, Key: model.Key, Name: model.Name, Description: model.Description}
}

func authTypeFromModel(model authTypeModel) domain.AuthType {
	return domain.AuthType{ID: model.ID, Name: model.Name, Strategy: model.Strategy, Fields: []domain.AuthTypeField(model.Fields), ServerTypeID: model.ServerTypeID}
}

func credentialFromModel(model credentialModel) domain.Credential {
	return domain.Credential{ID: model.ID, Name: model.Name, AuthTypeID: model.AuthTypeID, Data: map[string]any(model.Data), OwnerID: model.OwnerID, ProjectID: model.ProjectID, CreatedAt: model.CreatedAt}
}

func serverFromModel(model serverModel) domain.Server {
	return domain.Server{ID: model.ID, Name: model.Name, Type: model.Type, ServerTypeID: model.ServerTypeID, DriverKey: model.DriverKey, CredentialID: model.CredentialID, Headers: map[string]any(model.Headers), ProjectID: model.ProjectID, CreatedBy: model.CreatedBy, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

func serverResourceFromModel(model serverResourceModel) domain.ServerResource {
	return domain.ServerResource{ID: model.ID, ServerID: model.ServerID, Name: model.Name, Type: model.Type, Config: map[string]any(model.Config), CreatedAt: model.CreatedAt}
}

func resourceOperationFromModel(model resourceOperationModel) domain.ResourceOperation {
	return domain.ResourceOperation{ID: model.ID, ResourceID: model.ResourceID, Name: model.Name, Type: model.Type, Config: map[string]any(model.Config), InputSchema: map[string]any(model.InputSchema), OutputSchema: map[string]any(model.OutputSchema), CreatedAt: model.CreatedAt}
}
