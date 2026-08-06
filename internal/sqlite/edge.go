package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type EdgeRepository struct {
	db *gorm.DB
}

type UpsertEdgeConfigParams struct {
	EdgeName       *string
	Lat            *string
	Lng            *string
	LocationSource *string
	Tags           *[]string
	OS             *string
	OSVersion      *string
	EdgeVersion    *string
	Hardware       *string
	Environment    *string
	Description    *string
}

func NewEdgeRepository(db *gorm.DB) *EdgeRepository {
	return &EdgeRepository{db: db}
}

func (r *EdgeRepository) GetIdentity(ctx context.Context) (domain.EdgeIdentity, error) {
	var model edgeIdentityModel
	if err := r.db.WithContext(ctx).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.EdgeIdentity{}, ErrNotFound
		}
		return domain.EdgeIdentity{}, err
	}
	return edgeIdentityFromModel(model), nil
}

func (r *EdgeRepository) UpsertIdentity(ctx context.Context, identity domain.EdgeIdentity) (domain.EdgeIdentity, error) {
	var existing edgeIdentityModel
	err := r.db.WithContext(ctx).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.EdgeIdentity{}, err
	}
	model := edgeIdentityModel{
		ID:          existing.ID,
		EdgeID:      identity.EdgeID,
		EdgeName:    identity.EdgeName,
		Provisioned: identity.Provisioned,
		CreatedAt:   existing.CreatedAt,
	}
	if model.ID == "" {
		model.ID = newID()
	}
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().UTC()
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return domain.EdgeIdentity{}, err
		}
		return edgeIdentityFromModel(model), nil
	}
	if err := r.db.WithContext(ctx).Model(&edgeIdentityModel{}).Where("id = ?", existing.ID).Updates(model).Error; err != nil {
		return domain.EdgeIdentity{}, err
	}
	return r.GetIdentity(ctx)
}

func (r *EdgeRepository) ClearIdentity(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&edgeIdentityModel{}).Error
}

func (r *EdgeRepository) GetMqttCredentials(ctx context.Context) (domain.EdgeCredentials, error) {
	var model edgeCredentialModel
	if err := r.db.WithContext(ctx).Where("type = ?", "mqtt").First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.EdgeCredentials{}, ErrNotFound
		}
		return domain.EdgeCredentials{}, err
	}
	return edgeCredentialsFromModel(model), nil
}

func (r *EdgeRepository) UpsertMqttCredentials(ctx context.Context, credentials domain.EdgeCredentials) (domain.EdgeCredentials, error) {
	var existing edgeCredentialModel
	err := r.db.WithContext(ctx).Where("type = ?", "mqtt").First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.EdgeCredentials{}, err
	}
	model := edgeCredentialModel{
		ID:        existing.ID,
		Type:      "mqtt",
		BrokerURL: credentials.BrokerURL,
		Username:  credentials.Username,
		Password:  credentials.Password,
		UpdatedAt: time.Now().UTC(),
	}
	if model.ID == "" {
		model.ID = newID()
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return domain.EdgeCredentials{}, err
		}
		return edgeCredentialsFromModel(model), nil
	}
	if err := r.db.WithContext(ctx).Model(&edgeCredentialModel{}).Where("id = ?", existing.ID).Updates(model).Error; err != nil {
		return domain.EdgeCredentials{}, err
	}
	return r.GetMqttCredentials(ctx)
}

func (r *EdgeRepository) ClearMqttCredentials(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("type = ?", "mqtt").Delete(&edgeCredentialModel{}).Error
}

func (r *EdgeRepository) GetEdgeConfig(ctx context.Context) (domain.EdgeConfig, error) {
	var model edgeConfigModel
	if err := r.db.WithContext(ctx).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.EdgeConfig{}, ErrNotFound
		}
		return domain.EdgeConfig{}, err
	}
	return edgeConfigFromModel(model), nil
}

func (r *EdgeRepository) UpsertEdgeConfig(ctx context.Context, params UpsertEdgeConfigParams) (domain.EdgeConfig, error) {
	var existing edgeConfigModel
	err := r.db.WithContext(ctx).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.EdgeConfig{}, err
	}
	model := existing
	if model.ID == "" {
		model.ID = newID()
	}
	if params.EdgeName != nil {
		model.EdgeName = *params.EdgeName
	}
	if params.Lat != nil {
		model.Lat = *params.Lat
	}
	if params.Lng != nil {
		model.Lng = *params.Lng
	}
	if params.LocationSource != nil {
		model.LocationSource = *params.LocationSource
	}
	if params.Tags != nil {
		model.Tags = stringSliceJSON(*params.Tags)
	}
	if params.OS != nil {
		model.OS = *params.OS
	}
	if params.OSVersion != nil {
		model.OSVersion = *params.OSVersion
	}
	if params.EdgeVersion != nil {
		model.EdgeVersion = *params.EdgeVersion
	}
	if params.Hardware != nil {
		model.Hardware = *params.Hardware
	}
	if params.Environment != nil {
		model.Environment = *params.Environment
	}
	if params.Description != nil {
		model.Description = *params.Description
	}
	if model.LocationSource == "" {
		model.LocationSource = "manual"
	}
	if model.Environment == "" {
		model.Environment = "production"
	}
	model.UpdatedAt = time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return domain.EdgeConfig{}, err
		}
		return edgeConfigFromModel(model), nil
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.EdgeConfig{}, err
	}
	return edgeConfigFromModel(model), nil
}

func (r *EdgeRepository) ClearEdgeConfig(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&edgeConfigModel{}).Error
}

func edgeIdentityFromModel(model edgeIdentityModel) domain.EdgeIdentity {
	return domain.EdgeIdentity{ID: model.ID, EdgeID: model.EdgeID, EdgeName: model.EdgeName, Provisioned: model.Provisioned, CreatedAt: model.CreatedAt}
}

func edgeCredentialsFromModel(model edgeCredentialModel) domain.EdgeCredentials {
	return domain.EdgeCredentials{ID: model.ID, Type: model.Type, BrokerURL: model.BrokerURL, Username: model.Username, Password: model.Password, UpdatedAt: model.UpdatedAt}
}

func edgeConfigFromModel(model edgeConfigModel) domain.EdgeConfig {
	return domain.EdgeConfig{ID: model.ID, EdgeName: model.EdgeName, Lat: model.Lat, Lng: model.Lng, LocationSource: model.LocationSource, Tags: []string(model.Tags), OS: model.OS, OSVersion: model.OSVersion, EdgeVersion: model.EdgeVersion, Hardware: model.Hardware, Environment: model.Environment, Description: model.Description, UpdatedAt: model.UpdatedAt}
}
