package serverinfra

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidInput = errors.New("invalid server infrastructure input")

type Service struct {
	serverTypes        *sqlite.ServerTypesRepository
	authTypes          *sqlite.AuthTypesRepository
	credentials        *sqlite.CredentialsRepository
	servers            *sqlite.ServersRepository
	serverResources    *sqlite.ServerResourcesRepository
	resourceOperations *sqlite.ResourceOperationsRepository
}

func NewService(store *sqlite.Store) *Service {
	return &Service{
		serverTypes: store.ServerTypes, authTypes: store.AuthTypes, credentials: store.Credentials, servers: store.Servers,
		serverResources: store.ServerResources, resourceOperations: store.ResourceOperations,
	}
}

type CredentialRequest struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	AuthTypeID string         `json:"auth_type_id"`
	Data       map[string]any `json:"data"`
	OwnerID    string         `json:"owner_id"`
	ProjectID  string         `json:"project_id"`
}

type ServerRequest struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	ServerTypeID string         `json:"server_type_id"`
	DriverKey    string         `json:"driver_key"`
	CredentialID string         `json:"credential_id"`
	Headers      map[string]any `json:"headers"`
	ProjectID    string         `json:"project_id"`
	CreatedBy    string         `json:"created_by"`
}

func (s *Service) ListServerTypes(ctx context.Context) ([]domain.ServerType, error) {
	return s.serverTypes.ListAll(ctx)
}
func (s *Service) ListAuthTypes(ctx context.Context, serverTypeID string) ([]domain.AuthType, error) {
	return s.authTypes.ListByServerType(ctx, serverTypeID)
}
func (s *Service) SeedCatalog(ctx context.Context, serverTypes []domain.ServerType, authTypes []domain.AuthType) error {
	for _, item := range serverTypes {
		if _, err := s.serverTypes.Upsert(ctx, sqlite.UpsertServerTypeParams{ID: item.ID, Key: item.Key, Name: item.Name, Description: item.Description}); err != nil {
			return err
		}
	}
	for _, item := range authTypes {
		if _, err := s.authTypes.Upsert(ctx, sqlite.UpsertAuthTypeParams{ID: item.ID, Name: item.Name, Strategy: item.Strategy, Fields: item.Fields, ServerTypeID: item.ServerTypeID}); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) ListCredentials(ctx context.Context, ownerID string) ([]domain.Credential, error) {
	return s.credentials.ListByOwner(ctx, ownerID)
}
func (s *Service) FindCredential(ctx context.Context, id string) (domain.Credential, error) {
	return s.credentials.FindByID(ctx, id)
}
func (s *Service) UpsertCredential(ctx context.Context, req CredentialRequest) (domain.Credential, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.AuthTypeID) == "" {
		return domain.Credential{}, ErrInvalidInput
	}
	return s.credentials.Upsert(ctx, sqlite.UpsertCredentialParams{ID: req.ID, Name: req.Name, AuthTypeID: req.AuthTypeID, Data: req.Data, OwnerID: req.OwnerID, ProjectID: req.ProjectID})
}
func (s *Service) DeleteCredential(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.credentials.Delete(ctx, id)
}
func (s *Service) ListServers(ctx context.Context, projectID string) ([]domain.Server, error) {
	return s.servers.ListByProject(ctx, projectID)
}
func (s *Service) FindServer(ctx context.Context, id string) (domain.Server, error) {
	return s.servers.FindByID(ctx, id)
}
func (s *Service) UpsertServer(ctx context.Context, req ServerRequest) (domain.Server, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.DriverKey) == "" || strings.TrimSpace(req.ProjectID) == "" {
		return domain.Server{}, ErrInvalidInput
	}
	return s.servers.Upsert(ctx, sqlite.UpsertServerParams{ID: req.ID, Name: req.Name, Type: req.Type, ServerTypeID: req.ServerTypeID, DriverKey: req.DriverKey, CredentialID: req.CredentialID, Headers: req.Headers, ProjectID: req.ProjectID, CreatedBy: req.CreatedBy})
}
func (s *Service) DeleteServer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.servers.Delete(ctx, id)
}
func (s *Service) ListResources(ctx context.Context, serverID string) ([]domain.ServerResource, error) {
	return s.serverResources.ListByServer(ctx, serverID)
}
func (s *Service) ListOperations(ctx context.Context, resourceID string) ([]domain.ResourceOperation, error) {
	return s.resourceOperations.ListByResource(ctx, resourceID)
}
