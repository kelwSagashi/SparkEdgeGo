package sqlite

import "time"

type userModel struct {
	ID           string `gorm:"primaryKey;type:text"`
	Email        string `gorm:"uniqueIndex;not null;type:text"`
	FirstName    string `gorm:"type:text"`
	LastName     string `gorm:"type:text"`
	PasswordHash string `gorm:"type:text"`
	Role         string `gorm:"not null;default:viewer;type:text"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	APIKey       string `gorm:"uniqueIndex;not null;type:text"`
}

func (userModel) TableName() string {
	return "users"
}

type projectModel struct {
	ID          string `gorm:"primaryKey;type:text"`
	Name        string `gorm:"not null;type:text"`
	Key         string `gorm:"uniqueIndex;not null;type:text"`
	Description string `gorm:"type:text"`
	Visibility  string `gorm:"not null;default:private;type:text"`
	OwnerID     string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (projectModel) TableName() string {
	return "projects"
}

type projectMemberModel struct {
	ID        string `gorm:"primaryKey;type:text"`
	ProjectID string `gorm:"uniqueIndex:project_members_project_user_idx;not null;type:text"`
	UserID    string `gorm:"uniqueIndex:project_members_project_user_idx;not null;type:text"`
	Role      string `gorm:"not null;default:viewer;type:text"`
	CreatedAt time.Time
}

func (projectMemberModel) TableName() string {
	return "project_members"
}

type serverTypeModel struct {
	ID          string `gorm:"primaryKey;type:text"`
	Key         string `gorm:"uniqueIndex;not null;type:text"`
	Name        string `gorm:"not null;type:text"`
	Description string `gorm:"type:text"`
}

func (serverTypeModel) TableName() string {
	return "server_types"
}

type authTypeModel struct {
	ID           string             `gorm:"primaryKey;type:text"`
	Name         string             `gorm:"not null;type:text"`
	Strategy     string             `gorm:"not null;type:text"`
	Fields       authTypeFieldsJSON `gorm:"type:text"`
	ServerTypeID string             `gorm:"type:text"`
}

func (authTypeModel) TableName() string {
	return "auth_type"
}

type credentialModel struct {
	ID         string  `gorm:"primaryKey;type:text"`
	Name       string  `gorm:"not null;type:text"`
	AuthTypeID string  `gorm:"not null;type:text"`
	Data       mapJSON `gorm:"not null;type:text"`
	OwnerID    string  `gorm:"type:text"`
	ProjectID  string  `gorm:"type:text"`
	CreatedAt  time.Time
}

func (credentialModel) TableName() string {
	return "credentials"
}

type serverModel struct {
	ID           string  `gorm:"primaryKey;type:text"`
	Name         string  `gorm:"not null;type:text"`
	Type         string  `gorm:"not null;type:text"`
	ServerTypeID string  `gorm:"type:text"`
	DriverKey    string  `gorm:"not null;type:text"`
	CredentialID string  `gorm:"type:text"`
	Headers      mapJSON `gorm:"type:text"`
	ProjectID    string  `gorm:"not null;type:text"`
	CreatedBy    string  `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (serverModel) TableName() string {
	return "servers"
}

type serverResourceModel struct {
	ID        string  `gorm:"primaryKey;type:text"`
	ServerID  string  `gorm:"not null;index;type:text"`
	Name      string  `gorm:"not null;type:text"`
	Type      string  `gorm:"not null;type:text"`
	Config    mapJSON `gorm:"not null;type:text"`
	CreatedAt time.Time
}

func (serverResourceModel) TableName() string {
	return "server_resources"
}

type resourceOperationModel struct {
	ID           string  `gorm:"primaryKey;type:text"`
	ResourceID   string  `gorm:"not null;index;type:text"`
	Name         string  `gorm:"not null;type:text"`
	Type         string  `gorm:"not null;type:text"`
	Config       mapJSON `gorm:"type:text"`
	InputSchema  mapJSON `gorm:"type:text"`
	OutputSchema mapJSON `gorm:"type:text"`
	CreatedAt    time.Time
}

func (resourceOperationModel) TableName() string {
	return "resource_operations"
}

type downloadedScriptModel struct {
	ID               string          `gorm:"primaryKey;type:text"`
	Name             string          `gorm:"not null;type:text"`
	Description      string          `gorm:"type:text"`
	Author           string          `gorm:"not null;type:text"`
	Version          string          `gorm:"not null;default:1.0.0;type:text"`
	Source           string          `gorm:"not null;default:local;type:text"`
	GitHubRepo       string          `gorm:"type:text"`
	GitHubRef        string          `gorm:"type:text"`
	LocalPath        string          `gorm:"not null;type:text"`
	MainFile         string          `gorm:"not null;type:text"`
	VenvPath         string          `gorm:"type:text"`
	RequirementsFile string          `gorm:"type:text"`
	VenvReady        bool            `gorm:"not null;default:false"`
	Language         string          `gorm:"not null;default:python;type:text"`
	Tags             stringSliceJSON `gorm:"type:text"`
	SchemaConfig     mapJSON         `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (downloadedScriptModel) TableName() string {
	return "downloaded_scripts"
}

type deviceModel struct {
	ID                  string                `gorm:"primaryKey;type:text"`
	DeviceID            string                `gorm:"uniqueIndex;type:text"`
	Name                string                `gorm:"not null;type:text"`
	Brand               string                `gorm:"not null;type:text"`
	SerialNumber        string                `gorm:"type:text"`
	ConnectionMethod    string                `gorm:"not null;default:none;type:text"`
	IPAddress           string                `gorm:"type:text"`
	Location            string                `gorm:"type:text"`
	Description         string                `gorm:"type:text"`
	Others              deviceOtherFieldsJSON `gorm:"type:text"`
	ResourceOperationID string                `gorm:"type:text"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (deviceModel) TableName() string {
	return "devices"
}

type tagModel struct {
	ID        string `gorm:"primaryKey;type:text"`
	Name      string `gorm:"uniqueIndex:tags_project_name_idx;not null;type:text"`
	Color     string `gorm:"type:text;default:#6b7280"`
	ProjectID string `gorm:"uniqueIndex:tags_project_name_idx;type:text"`
	CreatedAt time.Time
}

func (tagModel) TableName() string {
	return "tags"
}

type instanceTagModel struct {
	ID         string `gorm:"primaryKey;type:text"`
	InstanceID string `gorm:"uniqueIndex:instance_tags_instance_tag_idx;not null;type:text"`
	TagID      string `gorm:"uniqueIndex:instance_tags_instance_tag_idx;not null;type:text"`
	CreatedAt  time.Time
}

func (instanceTagModel) TableName() string {
	return "instance_tags"
}

type instanceModel struct {
	ID                           string          `gorm:"primaryKey;type:text"`
	Name                         string          `gorm:"not null;type:text"`
	Description                  string          `gorm:"type:text"`
	Tags                         stringSliceJSON `gorm:"type:text"`
	Status                       string          `gorm:"not null;default:idle;type:text"`
	Active                       bool            `gorm:"not null;default:true"`
	ProjectID                    string          `gorm:"not null;type:text"`
	DeviceID                     string          `gorm:"type:text"`
	ScriptID                     string          `gorm:"type:text"`
	IncludeDeviceData            bool            `gorm:"not null;default:false"`
	ScriptParameters             mapJSON         `gorm:"type:text"`
	TriggerType                  string          `gorm:"not null;default:interval;type:text"`
	TriggerConfig                mapJSON         `gorm:"type:text"`
	FallbackEnabled              bool            `gorm:"not null;default:true"`
	FallbackStrategy             string          `gorm:"type:text;default:background_job"`
	FallbackRetryIntervalSeconds int             `gorm:"default:300"`
	OnErrorAction                string          `gorm:"not null;default:log_only;type:text"`
	OnErrorConfig                mapJSON         `gorm:"type:text"`
	CreatedBy                    string          `gorm:"type:text"`
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

func (instanceModel) TableName() string {
	return "instances"
}

type instanceDestinationModel struct {
	ID                  string  `gorm:"primaryKey;type:text"`
	InstanceID          string  `gorm:"not null;index;type:text"`
	ResourceOperationID string  `gorm:"not null;type:text"`
	Enabled             bool    `gorm:"not null;default:true"`
	Priority            int     `gorm:"default:0"`
	RetryPolicy         mapJSON `gorm:"type:text"`
	CreatedAt           time.Time
}

func (instanceDestinationModel) TableName() string {
	return "instance_destinations"
}

type dataMappingModel struct {
	ID                    string                  `gorm:"primaryKey;type:text"`
	InstanceDestinationID string                  `gorm:"not null;index;type:text"`
	Mapping               mapJSON                 `gorm:"type:text"`
	PayloadTemplate       mapJSON                 `gorm:"type:text"`
	CustomFields          mappingCustomFieldsJSON `gorm:"type:text"`
	TransformScript       string                  `gorm:"type:text"`
	CreatedAt             time.Time
}

func (dataMappingModel) TableName() string {
	return "data_mappings"
}

type instanceExecutionModel struct {
	ID              string `gorm:"primaryKey;type:text"`
	InstanceID      string `gorm:"not null;index;type:text"`
	Status          string `gorm:"not null;default:queued;type:text"`
	TriggerType     string `gorm:"not null;default:interval;type:text"`
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationMS      *int
	Logs            executionLogsJSON `gorm:"type:text"`
	Output          string            `gorm:"type:text"`
	ErrorMessage    string            `gorm:"type:text"`
	DestinationSent bool              `gorm:"not null;default:false"`
	FallbackUsed    bool              `gorm:"not null;default:false"`
	CreatedAt       time.Time
}

func (instanceExecutionModel) TableName() string {
	return "instance_executions"
}

type mqttCommandModel struct {
	ID         string  `gorm:"primaryKey;type:text"`
	CommandID  string  `gorm:"uniqueIndex;not null;type:text"`
	Type       string  `gorm:"not null;type:text"`
	Payload    mapJSON `gorm:"type:text"`
	Status     string  `gorm:"not null;default:pending;type:text"`
	Result     mapJSON `gorm:"type:text"`
	Error      string  `gorm:"type:text"`
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}

func (mqttCommandModel) TableName() string {
	return "mqtt_commands"
}

type mqttQueueModel struct {
	ID            string `gorm:"primaryKey;type:text"`
	Topic         string `gorm:"not null;type:text"`
	Payload       string `gorm:"not null;type:text"`
	Attempts      int    `gorm:"not null;default:0"`
	LastAttemptAt *time.Time
	CreatedAt     time.Time
}

func (mqttQueueModel) TableName() string {
	return "mqtt_queue"
}

type edgeConfigModel struct {
	ID             string          `gorm:"primaryKey;type:text"`
	EdgeName       string          `gorm:"type:text"`
	Lat            string          `gorm:"type:text"`
	Lng            string          `gorm:"type:text"`
	LocationSource string          `gorm:"type:text;default:manual"`
	Tags           stringSliceJSON `gorm:"type:text"`
	OS             string          `gorm:"column:os;type:text"`
	OSVersion      string          `gorm:"type:text"`
	EdgeVersion    string          `gorm:"type:text"`
	Hardware       string          `gorm:"type:text"`
	Environment    string          `gorm:"type:text;default:production"`
	Description    string          `gorm:"type:text"`
	UpdatedAt      time.Time
}

func (edgeConfigModel) TableName() string {
	return "edge_config"
}

type edgeIdentityModel struct {
	ID          string `gorm:"primaryKey;type:text"`
	EdgeID      string `gorm:"uniqueIndex;not null;type:text"`
	EdgeName    string `gorm:"type:text"`
	Provisioned bool   `gorm:"not null;default:false"`
	CreatedAt   time.Time
}

func (edgeIdentityModel) TableName() string {
	return "edge_identity"
}

type edgeCredentialModel struct {
	ID        string `gorm:"primaryKey;type:text"`
	Type      string `gorm:"uniqueIndex;not null;default:mqtt;type:text"`
	BrokerURL string `gorm:"type:text"`
	Username  string `gorm:"type:text"`
	Password  string `gorm:"type:text"`
	UpdatedAt time.Time
}

func (edgeCredentialModel) TableName() string {
	return "edge_credentials"
}
