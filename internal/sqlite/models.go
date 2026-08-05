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
