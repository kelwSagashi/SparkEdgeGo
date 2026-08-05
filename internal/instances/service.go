package instances

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidInstance = errors.New("invalid instance")

type Repository interface {
	Create(ctx context.Context, params sqlite.UpsertInstanceParams) (domain.Instance, error)
	Upsert(ctx context.Context, params sqlite.UpsertInstanceParams) (domain.Instance, error)
	FindByID(ctx context.Context, id string) (domain.Instance, error)
	ListAll(ctx context.Context) ([]domain.Instance, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Instance, error)
	ListActive(ctx context.Context) ([]domain.Instance, error)
	Update(ctx context.Context, id string, params sqlite.UpdateInstanceParams) (domain.Instance, error)
	UpdateStatus(ctx context.Context, id string, status domain.InstanceStatus) (domain.Instance, error)
	Delete(ctx context.Context, id string) error
}

type TagsService interface {
	FindOrCreateByNames(ctx context.Context, names []string, projectID string) ([]string, error)
	SyncTags(ctx context.Context, instanceID string, tagIDs []string) ([]domain.Tag, error)
}

type Service struct {
	instances Repository
	tags      TagsService
}

type Payload struct {
	ID                           string                  `json:"id"`
	Name                         string                  `json:"name"`
	Description                  string                  `json:"description"`
	Tags                         []string                `json:"tags"`
	Status                       domain.InstanceStatus   `json:"status"`
	Active                       *bool                   `json:"active"`
	ProjectID                    string                  `json:"project_id"`
	ProjectIDCamel               string                  `json:"projectId"`
	DeviceID                     string                  `json:"device_id"`
	DeviceIDCamel                string                  `json:"deviceId"`
	ScriptID                     string                  `json:"script_id"`
	ScriptIDCamel                string                  `json:"scriptId"`
	IncludeDeviceData            *bool                   `json:"include_device_data"`
	IncludeDeviceDataCamel       *bool                   `json:"includeDeviceData"`
	ScriptInputs                 map[string]any          `json:"script_inputs"`
	ScriptInputsCamel            map[string]any          `json:"scriptInputs"`
	ScriptParameters             any                     `json:"script_parameters"`
	ScriptParametersCamel        any                     `json:"scriptParameters"`
	TriggerType                  domain.TriggerType      `json:"trigger_type"`
	TriggerTypeCamel             domain.TriggerType      `json:"triggerType"`
	TriggerConfig                map[string]any          `json:"trigger_config"`
	TriggerConfigCamel           map[string]any          `json:"triggerConfig"`
	FallbackEnabled              *bool                   `json:"fallback_enabled"`
	FallbackStrategy             domain.FallbackStrategy `json:"fallback_strategy"`
	FallbackRetryIntervalSeconds int                     `json:"fallback_retry_interval_seconds"`
	FallbackConfig               map[string]any          `json:"fallback_config"`
	FallbackConfigCamel          map[string]any          `json:"fallbackConfig"`
	OnErrorAction                domain.OnErrorAction    `json:"on_error_action"`
	OnErrorConfig                map[string]any          `json:"on_error_config"`
	ErrorConfig                  map[string]any          `json:"error_config"`
	ErrorConfigCamel             map[string]any          `json:"errorConfig"`
	CreatedBy                    string                  `json:"created_by"`
	Destinations                 []any                   `json:"destinations"`
}

type WithDestinations struct {
	Instance     domain.Instance `json:"instance"`
	Destinations []any           `json:"destinations"`
}

func NewService(instances Repository, tags TagsService) *Service {
	return &Service{instances: instances, tags: tags}
}

func (s *Service) ListAll(ctx context.Context) ([]domain.Instance, error) {
	return s.instances.ListAll(ctx)
}

func (s *Service) ListActive(ctx context.Context) ([]domain.Instance, error) {
	return s.instances.ListActive(ctx)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]domain.Instance, error) {
	if strings.TrimSpace(projectID) == "" {
		return []domain.Instance{}, ErrInvalidInstance
	}
	return s.instances.ListByProject(ctx, projectID)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.Instance, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Instance{}, ErrInvalidInstance
	}
	return s.instances.FindByID(ctx, id)
}

func (s *Service) GetWithDestinations(ctx context.Context, id string) (WithDestinations, error) {
	instance, err := s.FindByID(ctx, id)
	if err != nil {
		return WithDestinations{}, err
	}
	return WithDestinations{Instance: instance, Destinations: []any{}}, nil
}

func (s *Service) Create(ctx context.Context, payload Payload) (domain.Instance, error) {
	params, err := normalizePayload(payload)
	if err != nil {
		return domain.Instance{}, err
	}
	instance, err := s.instances.Create(ctx, params)
	if err != nil {
		return domain.Instance{}, err
	}
	if err := s.syncNamedTags(ctx, instance.ID, payload.Tags, instance.ProjectID); err != nil {
		return domain.Instance{}, err
	}
	return instance, nil
}

func (s *Service) Update(ctx context.Context, id string, payload Payload) (domain.Instance, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Instance{}, ErrInvalidInstance
	}
	if len(payload.Destinations) > 0 {
		payload.ID = id
		params, err := normalizePayload(payload)
		if err != nil {
			return domain.Instance{}, err
		}
		instance, err := s.instances.Upsert(ctx, params)
		if err != nil {
			return domain.Instance{}, err
		}
		if err := s.syncNamedTags(ctx, instance.ID, payload.Tags, instance.ProjectID); err != nil {
			return domain.Instance{}, err
		}
		return instance, nil
	}

	update := normalizePartialUpdate(payload)
	instance, err := s.instances.Update(ctx, id, update)
	if err != nil {
		return domain.Instance{}, err
	}
	if payload.Tags != nil {
		if err := s.syncNamedTags(ctx, instance.ID, payload.Tags, instance.ProjectID); err != nil {
			return domain.Instance{}, err
		}
	}
	return instance, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status domain.InstanceStatus) (domain.Instance, error) {
	if strings.TrimSpace(id) == "" || status == "" {
		return domain.Instance{}, ErrInvalidInstance
	}
	return s.instances.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInstance
	}
	return s.instances.Delete(ctx, id)
}

func (s *Service) syncNamedTags(ctx context.Context, instanceID string, names []string, projectID string) error {
	if s.tags == nil || names == nil {
		return nil
	}
	tagIDs, err := s.tags.FindOrCreateByNames(ctx, names, projectID)
	if err != nil {
		return err
	}
	_, err = s.tags.SyncTags(ctx, instanceID, tagIDs)
	return err
}

func normalizePayload(payload Payload) (sqlite.UpsertInstanceParams, error) {
	projectID := firstNonEmpty(payload.ProjectID, payload.ProjectIDCamel)
	if strings.TrimSpace(payload.Name) == "" || strings.TrimSpace(projectID) == "" {
		return sqlite.UpsertInstanceParams{}, ErrInvalidInstance
	}

	includeDeviceData := firstBool(false, payload.IncludeDeviceData, payload.IncludeDeviceDataCamel)
	triggerType := firstTrigger(payload.TriggerType, payload.TriggerTypeCamel)
	triggerConfig := firstMap(payload.TriggerConfig, payload.TriggerConfigCamel)
	scriptParameters := normalizeScriptParameters(payload)
	fallbackEnabled, fallbackStrategy, retrySeconds := normalizeFallback(payload)
	onErrorAction, onErrorConfig := normalizeErrorConfig(payload)

	return sqlite.UpsertInstanceParams{
		ID:                           payload.ID,
		Name:                         payload.Name,
		Description:                  payload.Description,
		Tags:                         payload.Tags,
		Status:                       defaultStatus(payload.Status),
		Active:                       payload.Active,
		ProjectID:                    projectID,
		DeviceID:                     firstNonEmpty(payload.DeviceID, payload.DeviceIDCamel),
		ScriptID:                     firstNonEmpty(payload.ScriptID, payload.ScriptIDCamel),
		IncludeDeviceData:            includeDeviceData,
		ScriptParameters:             scriptParameters,
		TriggerType:                  triggerType,
		TriggerConfig:                triggerConfig,
		FallbackEnabled:              fallbackEnabled,
		FallbackStrategy:             fallbackStrategy,
		FallbackRetryIntervalSeconds: retrySeconds,
		OnErrorAction:                onErrorAction,
		OnErrorConfig:                onErrorConfig,
		CreatedBy:                    payload.CreatedBy,
	}, nil
}

func normalizePartialUpdate(payload Payload) sqlite.UpdateInstanceParams {
	update := sqlite.UpdateInstanceParams{}
	if payload.Name != "" {
		update.Name = &payload.Name
	}
	if payload.Description != "" {
		update.Description = &payload.Description
	}
	if payload.Tags != nil {
		update.Tags = &payload.Tags
	}
	if payload.Status != "" {
		update.Status = &payload.Status
	}
	update.Active = payload.Active
	if projectID := firstNonEmpty(payload.ProjectID, payload.ProjectIDCamel); projectID != "" {
		update.ProjectID = &projectID
	}
	if deviceID := firstNonEmpty(payload.DeviceID, payload.DeviceIDCamel); deviceID != "" {
		update.DeviceID = &deviceID
	}
	if scriptID := firstNonEmpty(payload.ScriptID, payload.ScriptIDCamel); scriptID != "" {
		update.ScriptID = &scriptID
	}
	if payload.IncludeDeviceData != nil || payload.IncludeDeviceDataCamel != nil {
		value := firstBool(false, payload.IncludeDeviceData, payload.IncludeDeviceDataCamel)
		update.IncludeDeviceData = &value
	}
	if payload.ScriptInputs != nil || payload.ScriptInputsCamel != nil || payload.ScriptParameters != nil || payload.ScriptParametersCamel != nil {
		params := normalizeScriptParameters(payload)
		update.ScriptParameters = &params
	}
	if payload.TriggerType != "" || payload.TriggerTypeCamel != "" {
		trigger := firstTrigger(payload.TriggerType, payload.TriggerTypeCamel)
		update.TriggerType = &trigger
	}
	if payload.TriggerConfig != nil || payload.TriggerConfigCamel != nil {
		config := firstMap(payload.TriggerConfig, payload.TriggerConfigCamel)
		update.TriggerConfig = &config
	}
	if payload.FallbackEnabled != nil || payload.FallbackConfig != nil || payload.FallbackConfigCamel != nil {
		enabled, strategy, retry := normalizeFallback(payload)
		update.FallbackEnabled = &enabled
		update.FallbackStrategy = &strategy
		update.FallbackRetryIntervalSeconds = &retry
	}
	if payload.OnErrorAction != "" || payload.OnErrorConfig != nil || payload.ErrorConfig != nil || payload.ErrorConfigCamel != nil {
		action, config := normalizeErrorConfig(payload)
		update.OnErrorAction = &action
		update.OnErrorConfig = &config
	}
	return update
}

func normalizeScriptParameters(payload Payload) map[string]any {
	result := map[string]any{}
	for key, value := range firstMap(payload.ScriptInputs, payload.ScriptInputsCamel) {
		result[key] = value
	}
	for key, value := range parametersToMap(payload.ScriptParameters) {
		result[key] = value
	}
	for key, value := range parametersToMap(payload.ScriptParametersCamel) {
		result[key] = value
	}
	return result
}

func parametersToMap(raw any) map[string]any {
	result := map[string]any{}
	switch params := raw.(type) {
	case map[string]any:
		return params
	case []any:
		for _, item := range params {
			if obj, ok := item.(map[string]any); ok {
				if key, ok := obj["key"].(string); ok && key != "" {
					result[key] = obj["value"]
				}
			}
		}
	}
	return result
}

func normalizeFallback(payload Payload) (bool, domain.FallbackStrategy, int) {
	config := firstMap(payload.FallbackConfig, payload.FallbackConfigCamel)
	enabled := firstBool(false, payload.FallbackEnabled)
	if value, ok := config["enabled"].(bool); ok {
		enabled = value
	}
	strategy := payload.FallbackStrategy
	if strategy == "" {
		if value, ok := config["strategy"].(string); ok {
			strategy = domain.FallbackStrategy(value)
		}
	}
	if strategy == "" {
		strategy = domain.FallbackBackgroundJob
	}

	retry := payload.FallbackRetryIntervalSeconds
	if retry == 0 {
		retry = intFromMap(config, "retry_interval_seconds", "retryIntervalSeconds", 300)
	}
	return enabled, strategy, retry
}

func normalizeErrorConfig(payload Payload) (domain.OnErrorAction, map[string]any) {
	config := firstMap(payload.OnErrorConfig, payload.ErrorConfig, payload.ErrorConfigCamel)
	action := payload.OnErrorAction
	if action == "" {
		if value, ok := config["action"].(string); ok {
			action = domain.OnErrorAction(value)
		}
	}
	if action == "" {
		action = domain.OnErrorLogOnly
	}
	return action, config
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return map[string]any{}
}

func firstBool(fallback bool, values ...*bool) bool {
	for _, value := range values {
		if value != nil {
			return *value
		}
	}
	return fallback
}

func firstTrigger(values ...domain.TriggerType) domain.TriggerType {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return domain.TriggerInterval
}

func defaultStatus(status domain.InstanceStatus) domain.InstanceStatus {
	if status == "" {
		return domain.InstanceStatusIdle
	}
	return status
}

func intFromMap(config map[string]any, snake string, camel string, fallback int) int {
	for _, key := range []string{snake, camel} {
		switch value := config[key].(type) {
		case float64:
			return int(value)
		case int:
			return value
		}
	}
	return fallback
}
