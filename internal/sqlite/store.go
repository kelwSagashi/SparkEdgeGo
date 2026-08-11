package sqlite

import (
	"context"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	Path               string
	db                 *gorm.DB
	Users              *UsersRepository
	Projects           *ProjectsRepository
	ProjectMembers     *ProjectMembersRepository
	ServerTypes        *ServerTypesRepository
	AuthTypes          *AuthTypesRepository
	Credentials        *CredentialsRepository
	Servers            *ServersRepository
	ServerResources    *ServerResourcesRepository
	ResourceOperations *ResourceOperationsRepository
	Scripts            *ScriptsRepository
	Devices            *DevicesRepository
	Tags               *TagsRepository
	InstanceTags       *InstanceTagsRepository
	Instances          *InstancesRepository
	Destinations       *InstanceDestinationsRepository
	DataMappings       *DataMappingsRepository
	Executions         *InstanceExecutionsRepository
	MqttCommands       *MqttCommandsRepository
	MqttQueue          *MqttQueueRepository
	Edge               *EdgeRepository
	LocalFallback      *LocalFallbackRepository
}

func NewStore() *Store {
	return &Store{Path: "sparkedge.db"}
}

func (s *Store) Open(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.db != nil {
		return nil
	}

	db, err := gorm.Open(sqlite.Open(s.Path), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return err
	}
	if err := migrate(db.WithContext(ctx)); err != nil {
		return err
	}

	s.db = db
	s.Users = NewUsersRepository(db)
	s.Projects = NewProjectsRepository(db)
	s.ProjectMembers = NewProjectMembersRepository(db)
	s.ServerTypes = NewServerTypesRepository(db)
	s.AuthTypes = NewAuthTypesRepository(db)
	s.Credentials = NewCredentialsRepository(db)
	s.Servers = NewServersRepository(db)
	s.ServerResources = NewServerResourcesRepository(db)
	s.ResourceOperations = NewResourceOperationsRepository(db)
	s.Scripts = NewScriptsRepository(db)
	s.Devices = NewDevicesRepository(db)
	s.Tags = NewTagsRepository(db)
	s.InstanceTags = NewInstanceTagsRepository(db)
	s.Instances = NewInstancesRepository(db)
	s.Destinations = NewInstanceDestinationsRepository(db)
	s.DataMappings = NewDataMappingsRepository(db)
	s.Executions = NewInstanceExecutionsRepository(db)
	s.MqttCommands = NewMqttCommandsRepository(db)
	s.MqttQueue = NewMqttQueueRepository(db)
	s.Edge = NewEdgeRepository(db)
	s.LocalFallback = NewLocalFallbackRepository(db)
	return nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&userModel{}, &projectModel{}, &projectMemberModel{}, &serverTypeModel{}, &authTypeModel{}, &credentialModel{}, &serverModel{}, &serverResourceModel{}, &resourceOperationModel{}, &downloadedScriptModel{}, &downloadedScriptHistoryModel{}, &deviceModel{}, &tagModel{}, &instanceTagModel{}, &instanceModel{}, &instanceDestinationModel{}, &dataMappingModel{}, &instanceExecutionModel{}, &mqttCommandModel{}, &mqttQueueModel{}, &edgeConfigModel{}, &edgeIdentityModel{}, &edgeCredentialModel{}, &localFallbackStorageModel{})
}
