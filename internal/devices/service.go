package devices

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidDevice = errors.New("invalid device")

type Repository interface {
	Create(ctx context.Context, params sqlite.UpsertDeviceParams) (domain.Device, error)
	Upsert(ctx context.Context, params sqlite.UpsertDeviceParams) (domain.Device, error)
	FindByID(ctx context.Context, id string) (domain.Device, error)
	FindByDeviceID(ctx context.Context, deviceID string) (domain.Device, error)
	ListAll(ctx context.Context) ([]domain.Device, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	devices Repository
}

type UpsertRequest struct {
	ID                  string                        `json:"id"`
	DeviceID            string                        `json:"device_id"`
	Name                string                        `json:"name"`
	Brand               string                        `json:"brand"`
	SerialNumber        string                        `json:"serial_number"`
	ConnectionMethod    domain.DeviceConnectionMethod `json:"connection_method"`
	IPAddress           string                        `json:"ip_address"`
	Location            string                        `json:"location"`
	Description         string                        `json:"description"`
	Others              []domain.DeviceOtherField     `json:"others"`
	ResourceOperationID string                        `json:"resource_operation_id"`
}

func NewService(devices Repository) *Service {
	return &Service{devices: devices}
}

func (s *Service) ListAll(ctx context.Context) ([]domain.Device, error) {
	return s.devices.ListAll(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.Device, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Device{}, ErrInvalidDevice
	}
	return s.devices.FindByID(ctx, id)
}

func (s *Service) FindByDeviceID(ctx context.Context, deviceID string) (domain.Device, error) {
	if strings.TrimSpace(deviceID) == "" {
		return domain.Device{}, ErrInvalidDevice
	}
	return s.devices.FindByDeviceID(ctx, deviceID)
}

func (s *Service) Upsert(ctx context.Context, req UpsertRequest) (domain.Device, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Brand) == "" {
		return domain.Device{}, ErrInvalidDevice
	}
	if req.ConnectionMethod == "" {
		req.ConnectionMethod = domain.DeviceConnectionNone
	}
	if req.Others == nil {
		req.Others = []domain.DeviceOtherField{}
	}

	return s.devices.Upsert(ctx, sqlite.UpsertDeviceParams{
		ID:                  req.ID,
		DeviceID:            req.DeviceID,
		Name:                req.Name,
		Brand:               req.Brand,
		SerialNumber:        req.SerialNumber,
		ConnectionMethod:    req.ConnectionMethod,
		IPAddress:           req.IPAddress,
		Location:            req.Location,
		Description:         req.Description,
		Others:              req.Others,
		ResourceOperationID: req.ResourceOperationID,
	})
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidDevice
	}
	return s.devices.Delete(ctx, id)
}
