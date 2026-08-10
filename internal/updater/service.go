package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
	"github.com/kelwSagashi/sparkedge-go/internal/appmeta"
)

var semverPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type Config struct {
	Enabled         bool
	Provider        string
	Repo            string
	Channel         string
	AllowPrerelease bool
	ServiceName     string
	RestartCommand  string
}

type Service struct {
	config Config
	client ReleaseClient
}

type ReleaseClient interface {
	ListReleases(ctx context.Context, repo string) ([]Release, error)
	OpenAsset(ctx context.Context, downloadURL string) (io.ReadCloser, error)
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
	IntegrityReady   bool      `json:"integrity_ready"`
	ExpectedSHA256   string    `json:"expected_sha256,omitempty"`
	CompatibilityMsg string    `json:"compatibility_message,omitempty"`
}

type AssetRef struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Size        int64  `json:"size"`
}

type Manifest struct {
	Version     string            `json:"version"`
	GeneratedAt string            `json:"generated_at,omitempty"`
	Packages    []ManifestPackage `json:"packages"`
}

type ManifestPackage struct {
	Target   string `json:"target"`
	FileName string `json:"file_name"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
}

type DownloadResult struct {
	Version          string `json:"version"`
	Target           string `json:"target"`
	AssetName        string `json:"asset_name"`
	ReleaseURL       string `json:"release_url"`
	DownloadedPath   string `json:"downloaded_path"`
	Size             int64  `json:"size"`
	SHA256           string `json:"sha256"`
	ChecksumVerified bool   `json:"checksum_verified"`
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

	resolved, ok, err := s.resolveLatestCompatible(ctx, releases, versionInfo.Target)
	if err != nil {
		return result, err
	}
	if !ok {
		result.CompatibilityMsg = "Nenhuma release compativel foi encontrada para esta plataforma."
		return result, nil
	}

	result.LatestVersion = resolved.Release.Version
	result.ReleaseName = firstNonBlank(resolved.Release.Name, resolved.Release.Version)
	result.ReleaseNotes = resolved.Release.Body
	result.ReleaseURL = resolved.Release.HTMLURL
	if !resolved.Release.PublishedAt.IsZero() {
		result.PublishedAt = resolved.Release.PublishedAt.UTC().Format(time.RFC3339)
	}
	result.CompatibleAsset = &AssetRef{
		Name:        resolved.Asset.Name,
		DownloadURL: resolved.Asset.DownloadURL,
		Size:        resolved.Asset.Size,
	}
	if resolved.Package != nil && strings.TrimSpace(resolved.Package.SHA256) != "" {
		result.IntegrityReady = true
		result.ExpectedSHA256 = resolved.Package.SHA256
	}

	cmp, comparable := compareVersions(versionInfo.Version, resolved.Release.Version)
	result.CanCompare = comparable
	result.UpdateAvailable = comparable && cmp < 0
	if !comparable {
		result.CompatibilityMsg = "A versao local nao esta em formato semver; comparacao automatica indisponivel."
	}

	return result, nil
}

func (s *Service) DownloadLatest(ctx context.Context) (DownloadResult, error) {
	versionInfo := appmeta.LoadVersionInfo()
	releases, err := s.client.ListReleases(ctx, s.config.Repo)
	if err != nil {
		return DownloadResult{}, err
	}

	resolved, ok, err := s.resolveLatestCompatible(ctx, releases, versionInfo.Target)
	if err != nil {
		return DownloadResult{}, err
	}
	if !ok {
		return DownloadResult{}, errors.New("nenhuma release compativel foi encontrada para esta plataforma")
	}
	if resolved.Package == nil || strings.TrimSpace(resolved.Package.SHA256) == "" {
		return DownloadResult{}, errors.New("release manifest or checksum unavailable for the compatible package")
	}

	stream, err := s.client.OpenAsset(ctx, resolved.Asset.DownloadURL)
	if err != nil {
		return DownloadResult{}, err
	}
	defer stream.Close()

	downloadDir := appfs.ResolveFromRoot("updates", "downloads", resolved.Release.Version)
	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return DownloadResult{}, err
	}

	fileName := filepath.Base(resolved.Asset.Name)
	outputPath := filepath.Join(downloadDir, fileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return DownloadResult{}, err
	}

	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hasher), stream)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(outputPath)
		return DownloadResult{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(outputPath)
		return DownloadResult{}, closeErr
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(sum, resolved.Package.SHA256) {
		_ = os.Remove(outputPath)
		return DownloadResult{}, fmt.Errorf("checksum mismatch for %s", resolved.Asset.Name)
	}

	result := DownloadResult{
		Version:          resolved.Release.Version,
		Target:           versionInfo.Target,
		AssetName:        resolved.Asset.Name,
		ReleaseURL:       resolved.Release.HTMLURL,
		DownloadedPath:   outputPath,
		Size:             written,
		SHA256:           sum,
		ChecksumVerified: true,
	}
	_ = s.saveState(UpdateState{
		LastDownloadedPackage: outputPath,
		LastPreparedVersion:   resolved.Release.Version,
		LastPreparedTarget:    versionInfo.Target,
		LastDownloadResult:    &result,
		UpdatedAt:             time.Now().UTC(),
	})
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

func (c *GitHubClient) OpenAsset(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "sparkedge-go-updater")
	if token := strings.TrimSpace(c.Token); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("asset download failed with status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

type resolvedRelease struct {
	Release  Release
	Asset    Asset
	Manifest *Manifest
	Package  *ManifestPackage
}

func (s *Service) resolveLatestCompatible(ctx context.Context, releases []Release, target string) (resolvedRelease, bool, error) {
	filtered := make([]Release, 0, len(releases))
	for _, release := range releases {
		if !s.config.AllowPrerelease && release.Prerelease {
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
		manifest, err := s.loadManifest(ctx, release)
		if err != nil {
			return resolvedRelease{}, false, err
		}
		if manifest != nil {
			if pkg, asset, ok := matchManifestPackage(release, *manifest, target); ok {
				return resolvedRelease{
					Release:  release,
					Asset:    asset,
					Manifest: manifest,
					Package:  &pkg,
				}, true, nil
			}
		}
		for _, asset := range release.Assets {
			if matchesTargetAsset(asset.Name, release.Version, target) {
				return resolvedRelease{
					Release: release,
					Asset:   asset,
				}, true, nil
			}
		}
	}

	return resolvedRelease{}, false, nil
}

func (s *Service) loadManifest(ctx context.Context, release Release) (*Manifest, error) {
	asset, ok := findManifestAsset(release.Assets)
	if !ok {
		return nil, nil
	}

	stream, err := s.client.OpenAsset(ctx, asset.DownloadURL)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	var manifest Manifest
	if err := decodeJSON(stream, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func findManifestAsset(assets []Asset) (Asset, bool) {
	for _, asset := range assets {
		if strings.EqualFold(strings.TrimSpace(asset.Name), "manifest.json") {
			return asset, true
		}
	}
	return Asset{}, false
}

func matchManifestPackage(release Release, manifest Manifest, target string) (ManifestPackage, Asset, bool) {
	for _, pkg := range manifest.Packages {
		if !strings.EqualFold(strings.TrimSpace(pkg.Target), strings.TrimSpace(target)) {
			continue
		}
		for _, asset := range release.Assets {
			if strings.EqualFold(strings.TrimSpace(asset.Name), strings.TrimSpace(pkg.FileName)) {
				return pkg, asset, true
			}
		}
	}
	return ManifestPackage{}, Asset{}, false
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
