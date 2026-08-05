package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/devices"
	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func (s *Server) handleDevicesList(r *http.Request) (any, error) {
	items, err := s.deps.Devices.ListAll(r.Context())
	if err != nil {
		return nil, err
	}
	return map[string]any{"data": publicDevices(items), "error": nil}, nil
}

func (s *Server) handleDeviceGet(r *http.Request) (any, error) {
	device, err := s.deps.Devices.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return deviceError(err)
	}
	return map[string]any{"data": publicDevice(device), "error": nil}, nil
}

func (s *Server) handleDeviceCreate(r *http.Request) (any, error) {
	var req devices.UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	device, err := s.deps.Devices.Upsert(r.Context(), req)
	if err != nil {
		return deviceError(err)
	}
	return map[string]any{"data": publicDevice(device), "error": nil}, nil
}

func (s *Server) handleDeviceUpdate(r *http.Request) (any, error) {
	var req devices.UpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.ID = r.PathValue("id")

	device, err := s.deps.Devices.Upsert(r.Context(), req)
	if err != nil {
		return deviceError(err)
	}
	return map[string]any{"data": publicDevice(device), "error": nil}, nil
}

func (s *Server) handleDeviceDelete(r *http.Request) (any, error) {
	if err := s.deps.Devices.Delete(r.Context(), r.PathValue("id")); err != nil {
		return deviceError(err)
	}
	return map[string]any{"data": map[string]any{"deleted": true}, "error": nil}, nil
}

func deviceError(err error) (any, error) {
	if errors.Is(err, devices.ErrInvalidDevice) {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid device")
	}
	if errors.Is(err, sqlite.ErrNotFound) {
		return map[string]any{"data": nil, "error": "Device not found"}, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return map[string]any{"data": nil, "error": "Device already exists"}, nil
	}
	return nil, err
}

func publicDevices(items []domain.Device) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicDevice(item))
	}
	return result
}

func publicDevice(device domain.Device) map[string]any {
	return map[string]any{
		"id":                    device.ID,
		"device_id":             device.DeviceID,
		"name":                  device.Name,
		"brand":                 device.Brand,
		"serial_number":         device.SerialNumber,
		"connection_method":     device.ConnectionMethod,
		"ip_address":            device.IPAddress,
		"location":              device.Location,
		"description":           device.Description,
		"others":                device.Others,
		"resource_operation_id": device.ResourceOperationID,
		"created_at":            device.CreatedAt,
		"updated_at":            device.UpdatedAt,
	}
}
