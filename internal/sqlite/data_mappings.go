package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type DataMappingsRepository struct {
	db *gorm.DB
}

type UpsertDataMappingParams struct {
	ID                    string
	InstanceDestinationID string
	Mapping               map[string]any
	PayloadTemplate       map[string]any
	CustomFields          []domain.MappingCustomField
	TransformScript       string
}

func NewDataMappingsRepository(db *gorm.DB) *DataMappingsRepository {
	return &DataMappingsRepository{db: db}
}

func (r *DataMappingsRepository) Upsert(ctx context.Context, params UpsertDataMappingParams) (domain.DataMapping, error) {
	if params.ID == "" {
		return r.Create(ctx, params)
	}

	var model dataMappingModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.Create(ctx, params)
	}
	if err != nil {
		return domain.DataMapping{}, err
	}

	applyDataMappingParams(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.DataMapping{}, err
	}
	return dataMappingFromModel(model), nil
}

func (r *DataMappingsRepository) Create(ctx context.Context, params UpsertDataMappingParams) (domain.DataMapping, error) {
	model := dataMappingModelFromParams(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.DataMapping{}, err
	}
	return dataMappingFromModel(model), nil
}

func (r *DataMappingsRepository) GetByInstanceDestination(ctx context.Context, instanceDestinationID string) (domain.DataMapping, error) {
	var model dataMappingModel
	if err := r.db.WithContext(ctx).Where("instance_destination_id = ?", instanceDestinationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.DataMapping{}, ErrNotFound
		}
		return domain.DataMapping{}, err
	}
	return dataMappingFromModel(model), nil
}

func (r *DataMappingsRepository) DeleteByInstanceDestination(ctx context.Context, instanceDestinationID string) error {
	return r.db.WithContext(ctx).Where("instance_destination_id = ?", instanceDestinationID).Delete(&dataMappingModel{}).Error
}

func dataMappingModelFromParams(params UpsertDataMappingParams) dataMappingModel {
	model := dataMappingModel{ID: params.ID}
	applyDataMappingParams(&model, params)
	return model
}

func applyDataMappingParams(model *dataMappingModel, params UpsertDataMappingParams) {
	if params.Mapping == nil {
		params.Mapping = map[string]any{}
	}
	if params.PayloadTemplate == nil {
		params.PayloadTemplate = map[string]any{}
	}
	if params.CustomFields == nil {
		params.CustomFields = []domain.MappingCustomField{}
	}

	model.InstanceDestinationID = params.InstanceDestinationID
	model.Mapping = mapJSON(params.Mapping)
	model.PayloadTemplate = mapJSON(params.PayloadTemplate)
	model.CustomFields = mappingCustomFieldsJSON(params.CustomFields)
	model.TransformScript = params.TransformScript
}

func dataMappingFromModel(model dataMappingModel) domain.DataMapping {
	return domain.DataMapping{
		ID:                    model.ID,
		InstanceDestinationID: model.InstanceDestinationID,
		Mapping:               map[string]any(model.Mapping),
		PayloadTemplate:       map[string]any(model.PayloadTemplate),
		CustomFields:          []domain.MappingCustomField(model.CustomFields),
		TransformScript:       model.TransformScript,
		CreatedAt:             model.CreatedAt,
	}
}
