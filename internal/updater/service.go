package updater

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appmeta"
)

var semverPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type Config struct {
	Enabled         bool
	Provider        string
	Repo            string
	Channel         string
	AllowPrerelease bool
}

type Service struct {
	config Config
	client ReleaseClient
}

type ReleaseClient interface {
	ListReleases(ctx context.Context, repo string) ([]Release, error)
}

type Release struct {
	Version     string
	Name        string
	Body        string
	HTMLURL     string
	PublishedAt time.Time
	Prerelease  bool
	Assets      []Asset
}

type Asset struct {
	Name        string
	DownloadURL string
	Size        int64
}

type CheckResult struct {
	Enabled          bool      `json:"enabled"`
	Provider         string    `json:"provider"`
	Repository       string    `json:"repository"`
	CurrentVersion   string    `json:"current_version"`
	CurrentTarget    string    `json:"current_target"`
	CanCompare       bool      `json:"can_compare"`
	UpdateAvailable  bool      `json:"update_available"`
	CheckedAt        time.Time `json:"checked_at"`
	LatestVersion    string    `json:"latest_version,omitempty"`
	ReleaseName      string    `json:"release_name,omitempty"`
	ReleaseNotes     string    `json:"release_notes,omitempty"`
	ReleaseURL       string    `json:"release_url,omitempty"`
	PublishedAt      string    `json:"published_at,omitempty"`
	CompatibleAsset  *AssetRef `json:"compatible_asset,omitempty"`
	CompatibilityMsg string    `json:"compatibility_message,omitempty"`
}

type AssetRef struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Size        int64  `json:"size"`
}

func NewService(config Config, client ReleaseClient) *Service {
	return &Service{
		config: config,
		client: client,
	}
}

func (s *Service) Check(ctx context.Context) (CheckResult, error) {
	versionInfo := appmeta.LoadVersionInfo()
	result := CheckResult{
		Enabled:        s.config.Enabled,
		Provider:       firstNonBlank(s.config.Provider, "github"),
		Repository:     s.config.Repo,
		CurrentVersion: versionInfo.Version,
		CurrentTarget:  versionInfo.Target,
		CheckedAt:      time.Now().UTC(),
	}

	if !s.config.Enabled {
		return result, nil
	}
	if strings.TrimSpace(s.config.Repo) == "" {
		return result, errors.New("update repository is not configured")
	}

	releases, err := s.client.ListReleases(ctx, s.config.Repo)
	if err != nil {
		return result, err
	}

	release, asset, ok := selectLatestCompatibleRelease(releases, versionInfo.Target, s.config.AllowPrerelease)
	if !ok {
		result.CompatibilityMsg = "Nenhuma release compativel foi encontrada para esta plataforma."
		return result, nil
	}

	result.LatestVersion = release.Version
	result.ReleaseName = firstNonBlank(release.Name, release.Version)
	result.ReleaseNotes = release.Body
	result.ReleaseURL = release.HTMLURL
	if !release.PublishedAt.IsZero() {
		result.PublishedAt = release.PublishedAt.UTC().Format(time.RFC3339)
	}
	result.CompatibleAsset = &AssetRef{
		Name:        asset.Name,
		DownloadURL: asset.DownloadURL,
		Size:        asset.Size,
	}

	cmp, comparable := compareVersions(versionInfo.Version, release.Version)
	result.CanCompare = comparable
	result.UpdateAvailable = comparable && cmp < 0
	if !comparable {
		result.CompatibilityMsg = "A versao local nao esta em formato semver; comparacao automatica indisponivel."
	}

	return result, nil
}

type GitHubClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func (c *GitHubClient) ListReleases(ctx context.Context, repo string) ([]Release, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return nil, errors.New("github repository is required")
	}

	baseURL := firstNonBlank(c.BaseURL, "https://api.github.com")
	requestURL := strings.TrimRight(baseURL, "/") + "/repos/" + repo + "/releases?per_page=20"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "sparkedge-go-updater")
	if token := strings.TrimSpace(c.Token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github releases request failed with status %d", resp.StatusCode)
	}

	var payload []struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
		Prerelease  bool   `json:"prerelease"`
		Assets      []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
			Size int64  `json:"size"`
		} `json:"assets"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return nil, err
	}

	releases := make([]Release, 0, len(payload))
	for _, item := range payload {
		release := Release{
			Version:    strings.TrimSpace(item.TagName),
			Name:       strings.TrimSpace(item.Name),
			Body:       item.Body,
			HTMLURL:    strings.TrimSpace(item.HTMLURL),
			Prerelease: item.Prerelease,
		}
		if publishedAt, err := time.Parse(time.RFC3339, item.PublishedAt); err == nil {
			release.PublishedAt = publishedAt
		}
		for _, asset := range item.Assets {
			release.Assets = append(release.Assets, Asset{
				Name:        strings.TrimSpace(asset.Name),
				DownloadURL: strings.TrimSpace(asset.URL),
				Size:        asset.Size,
			})
		}
		releases = append(releases, release)
	}

	return releases, nil
}

func selectLatestCompatibleRelease(releases []Release, target string, allowPrerelease bool) (Release, Asset, bool) {
	filtered := make([]Release, 0, len(releases))
	for _, release := range releases {
		if !allowPrerelease && release.Prerelease {
			continue
		}
		if !semverPattern.MatchString(strings.TrimSpace(release.Version)) {
			continue
		}
		filtered = append(filtered, release)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		cmp, _ := compareVersions(filtered[i].Version, filtered[j].Version)
		return cmp > 0
	})

	for _, release := range filtered {
		for _, asset := range release.Assets {
			if matchesTargetAsset(asset.Name, release.Version, target) {
				return release, asset, true
			}
		}
	}

	return Release{}, Asset{}, false
}

func matchesTargetAsset(assetName string, version string, target string) bool {
	expected := fmt.Sprintf("sparkedge-%s-%s.zip", version, target)
	return strings.EqualFold(strings.TrimSpace(assetName), expected)
}

func compareVersions(current string, latest string) (int, bool) {
	currentParts, ok := parseSemver(current)
	if !ok {
		return 0, false
	}
	latestParts, ok := parseSemver(latest)
	if !ok {
		return 0, false
	}
	for idx := range currentParts {
		switch {
		case currentParts[idx] < latestParts[idx]:
			return -1, true
		case currentParts[idx] > latestParts[idx]:
			return 1, true
		}
	}
	return 0, true
}

func parseSemver(value string) ([3]int, bool) {
	match := semverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return [3]int{}, false
	}

	var parts [3]int
	for idx := 0; idx < 3; idx++ {
		parsed, err := strconv.Atoi(match[idx+1])
		if err != nil {
			return [3]int{}, false
		}
		parts[idx] = parsed
	}

	return parts, true
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
