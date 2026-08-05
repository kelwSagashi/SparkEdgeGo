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
