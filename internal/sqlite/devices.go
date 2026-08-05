package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type DevicesRepository struct {
	db *gorm.DB
}

type UpsertDeviceParams struct {
	ID                  string
	DeviceID            string
	Name                string
	Brand               string
	SerialNumber        string
	ConnectionMethod    domain.DeviceConnectionMethod
	IPAddress           string
	Location            string
	Description         string
	Others              []domain.DeviceOtherField
	ResourceOperationID string
}

func NewDevicesRepository(db *gorm.DB) *DevicesRepository {
	return &DevicesRepository{db: db}
}

func (r *DevicesRepository) Create(ctx context.Context, params UpsertDeviceParams) (domain.Device, error) {
	model := deviceModelFromParams(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if model.DeviceID == "" {
		model.DeviceID = newID()
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Device{}, err
	}
	return deviceFromModel(model), nil
}

func (r *DevicesRepository) Upsert(ctx context.Context, params UpsertDeviceParams) (domain.Device, error) {
	if params.ID == "" {
		return r.Create(ctx, params)
	}

	var model deviceModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.Create(ctx, params)
	}
	if err != nil {
		return domain.Device{}, err
	}

	applyDeviceParams(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Device{}, err
	}
	return deviceFromModel(model), nil
}

func (r *DevicesRepository) FindByID(ctx context.Context, id string) (domain.Device, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *DevicesRepository) FindByDeviceID(ctx context.Context, deviceID string) (domain.Device, error) {
	return r.findOne(ctx, "device_id = ?", deviceID)
}

func (r *DevicesRepository) ListAll(ctx context.Context) ([]domain.Device, error) {
	var models []deviceModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	devices := make([]domain.Device, 0, len(models))
	for _, model := range models {
		devices = append(devices, deviceFromModel(model))
	}
	return devices, nil
}

func (r *DevicesRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&deviceModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *DevicesRepository) findOne(ctx context.Context, query string, args ...any) (domain.Device, error) {
	var model deviceModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Device{}, ErrNotFound
		}
		return domain.Device{}, err
	}
	return deviceFromModel(model), nil
}

func deviceModelFromParams(params UpsertDeviceParams) deviceModel {
	model := deviceModel{ID: params.ID}
	applyDeviceParams(&model, params)
	return model
}

func applyDeviceParams(model *deviceModel, params UpsertDeviceParams) {
	if params.ConnectionMethod == "" {
		params.ConnectionMethod = domain.DeviceConnectionNone
	}

	model.DeviceID = params.DeviceID
	model.Name = params.Name
	model.Brand = params.Brand
	model.SerialNumber = params.SerialNumber
	model.ConnectionMethod = string(params.ConnectionMethod)
	model.IPAddress = params.IPAddress
	model.Location = params.Location
	model.Description = params.Description
	model.Others = deviceOtherFieldsJSON(params.Others)
	model.ResourceOperationID = params.ResourceOperationID
}

func deviceFromModel(model deviceModel) domain.Device {
	return domain.Device{
		ID:                  model.ID,
		DeviceID:            model.DeviceID,
		Name:                model.Name,
		Brand:               model.Brand,
		SerialNumber:        model.SerialNumber,
		ConnectionMethod:    domain.DeviceConnectionMethod(model.ConnectionMethod),
		IPAddress:           model.IPAddress,
		Location:            model.Location,
		Description:         model.Description,
		Others:              []domain.DeviceOtherField(model.Others),
		ResourceOperationID: model.ResourceOperationID,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
