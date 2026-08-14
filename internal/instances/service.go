package instances

import (
	"context"
	"errors"
	"fmt"
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

type DestinationRepository interface {
	Upsert(ctx context.Context, params sqlite.UpsertInstanceDestinationParams) (domain.InstanceDestination, error)
	FindByID(ctx context.Context, id string) (domain.InstanceDestination, error)
	ListByInstance(ctx context.Context, instanceID string) ([]domain.InstanceDestination, error)
	DeleteByInstance(ctx context.Context, instanceID string) error
	Delete(ctx context.Context, id string) error
}

type DataMappingRepository interface {
	Upsert(ctx context.Context, params sqlite.UpsertDataMappingParams) (domain.DataMapping, error)
	GetByInstanceDestination(ctx context.Context, instanceDestinationID string) (domain.DataMapping, error)
	DeleteByInstanceDestination(ctx context.Context, instanceDestinationID string) error
}

type CircuitBreakerRepository interface {
	ListByDestinationIDs(ctx context.Context, destinationIDs []string) ([]domain.CircuitBreakerState, error)
}

type Service struct {
	instances    Repository
	tags         TagsService
	destinations DestinationRepository
	mappings     DataMappingRepository
	breakers     CircuitBreakerRepository
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
	DependsOn                    []string                `json:"depends_on"`
	DependsOnCamel               []string                `json:"dependsOn"`
	ExecutionMode                domain.ExecutionMode    `json:"execution_mode"`
	ExecutionModeCamel           domain.ExecutionMode    `json:"executionMode"`
	OrchestrationConfig          map[string]any          `json:"orchestration_config"`
	OrchestrationConfigCamel     map[string]any          `json:"orchestrationConfig"`
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
	Destinations                 []DestinationPayload    `json:"destinations"`
}

type DestinationPayload struct {
	Destination          *DestinationData `json:"destination"`
	Mapping              *MappingData     `json:"mapping"`
	DataMapping          *MappingData     `json:"data_mapping"`
	DataMappingCamel     *MappingData     `json:"dataMapping"`
	ResourceOperationID  string           `json:"resource_operation_id"`
	ResourceOperationID2 string           `json:"resourceOperationId"`
	Enabled              *bool            `json:"enabled"`
	Priority             int              `json:"priority"`
	RetryPolicy          map[string]any   `json:"retry_policy"`
	RetryPolicyCamel     map[string]any   `json:"retryPolicy"`
}

type DestinationData struct {
	ID                   string         `json:"id"`
	ResourceOperationID  string         `json:"resource_operation_id"`
	ResourceOperationID2 string         `json:"resourceOperationId"`
	Enabled              *bool          `json:"enabled"`
	Priority             int            `json:"priority"`
	RetryPolicy          map[string]any `json:"retry_policy"`
	RetryPolicyCamel     map[string]any `json:"retryPolicy"`
}

type MappingData struct {
	ID                   string                      `json:"id"`
	Mapping              map[string]any              `json:"mapping"`
	PayloadTemplate      map[string]any              `json:"payload_template"`
	PayloadTemplateCamel map[string]any              `json:"payloadTemplate"`
	CustomFields         []domain.MappingCustomField `json:"custom_fields"`
	CustomFieldsCamel    []domain.MappingCustomField `json:"customFields"`
	TransformScript      string                      `json:"transform_script"`
	TransformScriptCamel string                      `json:"transformScript"`
}

type WithDestinations struct {
	Instance     domain.Instance                         `json:"instance"`
	Destinations []domain.InstanceDestinationWithMapping `json:"destinations"`
}

func NewService(instances Repository, tags TagsService, repos ...any) *Service {
	service := &Service{instances: instances, tags: tags}
	for _, repo := range repos {
		switch typed := repo.(type) {
		case DestinationRepository:
			service.destinations = typed
		case DataMappingRepository:
			service.mappings = typed
		case CircuitBreakerRepository:
			service.breakers = typed
		}
	}
	return service
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
	destinations, err := s.destinationsWithMappings(ctx, instance.ID)
	if err != nil {
		return WithDestinations{}, err
	}
	return WithDestinations{Instance: instance, Destinations: destinations}, nil
}

func (s *Service) Create(ctx context.Context, payload Payload) (domain.Instance, error) {
	params, err := normalizePayload(payload)
	if err != nil {
		return domain.Instance{}, err
	}
	if err := s.validateDependencies(ctx, "", params.DependsOn); err != nil {
		return domain.Instance{}, err
	}
	instance, err := s.instances.Create(ctx, params)
	if err != nil {
		return domain.Instance{}, err
	}
	if err := s.syncNamedTags(ctx, instance.ID, payload.Tags, instance.ProjectID); err != nil {
		return domain.Instance{}, err
	}
	if err := s.syncDestinations(ctx, instance.ID, payload.Destinations); err != nil {
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
		if err := s.validateDependencies(ctx, id, params.DependsOn); err != nil {
			return domain.Instance{}, err
		}
		instance, err := s.instances.Upsert(ctx, params)
		if err != nil {
			return domain.Instance{}, err
		}
		if err := s.syncNamedTags(ctx, instance.ID, payload.Tags, instance.ProjectID); err != nil {
			return domain.Instance{}, err
		}
		if err := s.syncDestinations(ctx, instance.ID, payload.Destinations); err != nil {
			return domain.Instance{}, err
		}
		return instance, nil
	}

	update := normalizePartialUpdate(payload)
	if update.DependsOn != nil {
		if err := s.validateDependencies(ctx, id, *update.DependsOn); err != nil {
			return domain.Instance{}, err
		}
	}
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

func (s *Service) ListDestinations(ctx context.Context, instanceID string) ([]domain.InstanceDestinationWithMapping, error) {
	if strings.TrimSpace(instanceID) == "" {
		return nil, ErrInvalidInstance
	}
	return s.destinationsWithMappings(ctx, instanceID)
}

func (s *Service) AddDestination(ctx context.Context, instanceID string, payload DestinationPayload) (domain.InstanceDestinationWithMapping, error) {
	if strings.TrimSpace(instanceID) == "" || s.destinations == nil {
		return domain.InstanceDestinationWithMapping{}, ErrInvalidInstance
	}
	destination, err := s.destinations.Upsert(ctx, destinationParams(instanceID, payload))
	if err != nil {
		return domain.InstanceDestinationWithMapping{}, err
	}
	item := domain.InstanceDestinationWithMapping{Destination: destination}
	mapping := firstMappingData(payload.Mapping, payload.DataMapping, payload.DataMappingCamel)
	if s.mappings != nil && mapping != nil {
		saved, err := s.mappings.Upsert(ctx, mappingParams(destination.ID, *mapping))
		if err != nil {
			return domain.InstanceDestinationWithMapping{}, err
		}
		item.Mapping = &saved
	}
	return item, nil
}

func (s *Service) UpdateDestination(ctx context.Context, destinationID string, payload DestinationPayload) (domain.InstanceDestinationWithMapping, error) {
	if strings.TrimSpace(destinationID) == "" || s.destinations == nil {
		return domain.InstanceDestinationWithMapping{}, ErrInvalidInstance
	}
	existing, err := s.destinations.FindByID(ctx, destinationID)
	if err != nil {
		return domain.InstanceDestinationWithMapping{}, err
	}
	payload.Destination = &DestinationData{
		ID:                  destinationID,
		ResourceOperationID: existing.ResourceOperationID,
		Enabled:             &existing.Enabled,
		Priority:            existing.Priority,
		RetryPolicy: map[string]any{
			"max_retries":    existing.RetryPolicy.MaxRetries,
			"retry_interval": existing.RetryPolicy.RetryInterval,
		},
	}
	destination, err := s.destinations.Upsert(ctx, destinationParams(existing.InstanceID, payload))
	if err != nil {
		return domain.InstanceDestinationWithMapping{}, err
	}
	item := domain.InstanceDestinationWithMapping{Destination: destination}
	if s.mappings != nil {
		mapping, err := s.mappings.GetByInstanceDestination(ctx, destination.ID)
		if err == nil {
			item.Mapping = &mapping
		} else if !errors.Is(err, sqlite.ErrNotFound) {
			return domain.InstanceDestinationWithMapping{}, err
		}
	}
	return item, nil
}

func (s *Service) DeleteDestination(ctx context.Context, destinationID string) error {
	if strings.TrimSpace(destinationID) == "" || s.destinations == nil {
		return ErrInvalidInstance
	}
	if s.mappings != nil {
		if err := s.mappings.DeleteByInstanceDestination(ctx, destinationID); err != nil {
			return err
		}
	}
	return s.destinations.Delete(ctx, destinationID)
}

func (s *Service) SetDataMapping(ctx context.Context, destinationID string, payload MappingData) (domain.DataMapping, error) {
	if strings.TrimSpace(destinationID) == "" || s.mappings == nil {
		return domain.DataMapping{}, ErrInvalidInstance
	}
	return s.mappings.Upsert(ctx, mappingParams(destinationID, payload))
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInstance
	}
	return s.instances.Delete(ctx, id)
}

func (s *Service) validateDependencies(ctx context.Context, instanceID string, dependsOn []string) error {
	if len(dependsOn) == 0 {
		return nil
	}

	allInstances, err := s.instances.ListAll(ctx)
	if err != nil {
		return err
	}

	graph := make(map[string][]string, len(allInstances)+1)
	known := make(map[string]struct{}, len(allInstances))
	for _, item := range allInstances {
		graph[item.ID] = append([]string{}, item.DependsOn...)
		known[item.ID] = struct{}{}
	}

	normalized := make([]string, 0, len(dependsOn))
	seen := map[string]struct{}{}
	for _, dependency := range dependsOn {
		dependency = strings.TrimSpace(dependency)
		if dependency == "" {
			continue
		}
		if dependency == instanceID && instanceID != "" {
			return fmt.Errorf("%w: instancia nao pode depender dela mesma", ErrInvalidInstance)
		}
		if _, ok := known[dependency]; !ok {
			return fmt.Errorf("%w: dependencia %s nao existe", ErrInvalidInstance, dependency)
		}
		if _, ok := seen[dependency]; ok {
			continue
		}
		seen[dependency] = struct{}{}
		normalized = append(normalized, dependency)
	}

	if instanceID == "" {
		return nil
	}

	graph[instanceID] = normalized
	if hasDependencyCycle(instanceID, graph) {
		return fmt.Errorf("%w: dependencias geram ciclo na orquestracao", ErrInvalidInstance)
	}
	return nil
}

func hasDependencyCycle(root string, graph map[string][]string) bool {
	visited := map[string]bool{}
	active := map[string]bool{}

	var visit func(string) bool
	visit = func(node string) bool {
		if active[node] {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		active[node] = true
		for _, dependency := range graph[node] {
			if visit(dependency) {
				return true
			}
		}
		delete(active, node)
		return false
	}

	return visit(root)
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

func (s *Service) destinationsWithMappings(ctx context.Context, instanceID string) ([]domain.InstanceDestinationWithMapping, error) {
	if s.destinations == nil {
		return []domain.InstanceDestinationWithMapping{}, nil
	}
	destinations, err := s.destinations.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.InstanceDestinationWithMapping, 0, len(destinations))
	breakerByDestination := map[string]domain.CircuitBreakerState{}
	if s.breakers != nil {
		ids := make([]string, 0, len(destinations))
		for _, destination := range destinations {
			ids = append(ids, destination.ID)
		}
		states, err := s.breakers.ListByDestinationIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		for _, state := range states {
			breakerByDestination[state.DestinationID] = state
		}
	}
	for _, destination := range destinations {
		item := domain.InstanceDestinationWithMapping{Destination: destination}
		if s.mappings != nil {
			mapping, err := s.mappings.GetByInstanceDestination(ctx, destination.ID)
			if err == nil {
				item.Mapping = &mapping
			} else if !errors.Is(err, sqlite.ErrNotFound) {
				return nil, err
			}
		}
		if state, ok := breakerByDestination[destination.ID]; ok {
			stateCopy := state
			item.BreakerState = &stateCopy
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) syncDestinations(ctx context.Context, instanceID string, payloads []DestinationPayload) error {
	if s.destinations == nil || payloads == nil {
		return nil
	}

	if s.mappings != nil {
		existing, err := s.destinations.ListByInstance(ctx, instanceID)
		if err != nil {
			return err
		}
		for _, destination := range existing {
			if err := s.mappings.DeleteByInstanceDestination(ctx, destination.ID); err != nil {
				return err
			}
		}
	}
	if err := s.destinations.DeleteByInstance(ctx, instanceID); err != nil {
		return err
	}

	for _, payload := range payloads {
		destination, err := s.destinations.Upsert(ctx, destinationParams(instanceID, payload))
		if err != nil {
			return err
		}
		mapping := firstMappingData(payload.Mapping, payload.DataMapping, payload.DataMappingCamel)
		if s.mappings != nil && mapping != nil {
			if _, err := s.mappings.Upsert(ctx, mappingParams(destination.ID, *mapping)); err != nil {
				return err
			}
		}
	}
	return nil
}

func destinationParams(instanceID string, payload DestinationPayload) sqlite.UpsertInstanceDestinationParams {
	data := payload.Destination
	resourceOperationID := firstNonEmpty(payload.ResourceOperationID, payload.ResourceOperationID2)
	enabled := payload.Enabled
	priority := payload.Priority
	retryPolicy := firstMap(payload.RetryPolicy, payload.RetryPolicyCamel)
	id := ""

	if data != nil {
		id = data.ID
		resourceOperationID = firstNonEmpty(resourceOperationID, data.ResourceOperationID, data.ResourceOperationID2)
		enabled = firstBoolPointer(enabled, data.Enabled)
		if priority == 0 {
			priority = data.Priority
		}
		if len(retryPolicy) == 0 {
			retryPolicy = firstMap(data.RetryPolicy, data.RetryPolicyCamel)
		}
	}

	return sqlite.UpsertInstanceDestinationParams{
		ID:                  id,
		InstanceID:          instanceID,
		ResourceOperationID: resourceOperationID,
		Enabled:             firstBool(true, enabled),
		Priority:            priority,
		RetryPolicy: domain.RetryPolicy{
			MaxRetries:                    intFromMap(retryPolicy, "max_retries", "maxRetries", 3),
			RetryInterval:                 intFromMap(retryPolicy, "retry_interval", "retryInterval", 60),
			TimeoutSeconds:                intFromMap(retryPolicy, "timeout_seconds", "timeoutSeconds", 30),
			ContinueOnError:               boolFromMap(retryPolicy, "continue_on_error", "continueOnError", false),
			IsolationMode:                 stringFromMap(retryPolicy, "isolation_mode", "isolationMode", "isolate"),
			CircuitBreakerThreshold:       intFromMap(retryPolicy, "circuit_breaker_threshold", "circuitBreakerThreshold", 0),
			CircuitBreakerCooldownSeconds: intFromMap(retryPolicy, "circuit_breaker_cooldown_seconds", "circuitBreakerCooldownSeconds", 0),
		},
	}
}

func mappingParams(destinationID string, payload MappingData) sqlite.UpsertDataMappingParams {
	return sqlite.UpsertDataMappingParams{
		ID:                    payload.ID,
		InstanceDestinationID: destinationID,
		Mapping:               payload.Mapping,
		PayloadTemplate:       firstMap(payload.PayloadTemplate, payload.PayloadTemplateCamel),
		CustomFields:          firstCustomFields(payload.CustomFields, payload.CustomFieldsCamel),
		TransformScript:       firstNonEmpty(payload.TransformScript, payload.TransformScriptCamel),
	}
}

func normalizePayload(payload Payload) (sqlite.UpsertInstanceParams, error) {
	projectID := firstNonEmpty(payload.ProjectID, payload.ProjectIDCamel)
	if strings.TrimSpace(payload.Name) == "" || strings.TrimSpace(projectID) == "" {
		return sqlite.UpsertInstanceParams{}, ErrInvalidInstance
	}

	includeDeviceData := firstBool(false, payload.IncludeDeviceData, payload.IncludeDeviceDataCamel)
	triggerType := firstTrigger(payload.TriggerType, payload.TriggerTypeCamel)
	triggerConfig := firstMap(payload.TriggerConfig, payload.TriggerConfigCamel)
	dependsOn := firstStrings(payload.DependsOn, payload.DependsOnCamel)
	executionMode := firstExecutionMode(payload.ExecutionMode, payload.ExecutionModeCamel)
	orchestrationConfig := firstMap(payload.OrchestrationConfig, payload.OrchestrationConfigCamel)
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
		DependsOn:                    dependsOn,
		ExecutionMode:                executionMode,
		OrchestrationConfig:          orchestrationConfig,
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
	if payload.DependsOn != nil || payload.DependsOnCamel != nil {
		dependsOn := firstStrings(payload.DependsOn, payload.DependsOnCamel)
		update.DependsOn = &dependsOn
	}
	if payload.ExecutionMode != "" || payload.ExecutionModeCamel != "" {
		mode := firstExecutionMode(payload.ExecutionMode, payload.ExecutionModeCamel)
		update.ExecutionMode = &mode
	}
	if payload.OrchestrationConfig != nil || payload.OrchestrationConfigCamel != nil {
		config := firstMap(payload.OrchestrationConfig, payload.OrchestrationConfigCamel)
		update.OrchestrationConfig = &config
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

func firstBoolPointer(values ...*bool) *bool {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstMappingData(values ...*MappingData) *MappingData {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstCustomFields(values ...[]domain.MappingCustomField) []domain.MappingCustomField {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return []domain.MappingCustomField{}
}

func firstTrigger(values ...domain.TriggerType) domain.TriggerType {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return domain.TriggerInterval
}

func firstExecutionMode(values ...domain.ExecutionMode) domain.ExecutionMode {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return domain.ExecutionModeSequential
}

func firstStrings(values ...[]string) []string {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return []string{}
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

func boolFromMap(config map[string]any, snake string, camel string, fallback bool) bool {
	for _, key := range []string{snake, camel} {
		if value, ok := config[key].(bool); ok {
			return value
		}
	}
	return fallback
}

func stringFromMap(config map[string]any, snake string, camel string, fallback string) string {
	for _, key := range []string{snake, camel} {
		if value, ok := config[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return fallback
}
