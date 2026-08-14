package domain

import "time"

type ScriptInput struct {
	Values map[string]any
}

type ScriptResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Data      map[string]any
	ErrorData map[string]any
}

type ScriptSource string

const (
	ScriptSourceLocal     ScriptSource = "local"
	ScriptSourceHubGitHub ScriptSource = "hub_github"
)

type ScriptLanguage string

const ScriptLanguagePython ScriptLanguage = "python"

type DownloadedScript struct {
	ID               string
	Name             string
	Description      string
	Author           string
	Version          string
	Source           ScriptSource
	GitHubRepo       string
	GitHubRef        string
	LocalPath        string
	MainFile         string
	VenvPath         string
	RequirementsFile string
	VenvReady        bool
	Language         ScriptLanguage
	Tags             []string
	SchemaConfig     map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ScriptHistoryEntry struct {
	ID               string
	ScriptID         string
	Action           string
	Name             string
	Description      string
	Author           string
	Version          string
	MainFile         string
	RequirementsFile string
	Tags             []string
	SchemaConfig     map[string]any
	ChangeSummary    []string
	BundlePath       string
	CreatedAt        time.Time
}
