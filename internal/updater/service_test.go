package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestCheckDetectsNewVersion(t *testing.T) {
	service := NewService(Config{
		Enabled: true,
		Repo:    "kelwSagashi/SparkEdgeGo",
	}, fakeReleaseClient{releases: []Release{
		{
			Version: "v0.2.0",
			Name:    "v0.2.0",
			Assets: []Asset{
				{Name: "manifest.json", DownloadURL: "https://example.invalid/manifest.json", Size: 200},
				{Name: "sparkedge-v0.2.0-linux-armv7.zip", DownloadURL: "https://example.invalid/linux-armv7.zip", Size: 42},
			},
			PublishedAt: time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC),
		},
	}, downloads: map[string]string{
		"https://example.invalid/manifest.json": `{"version":"v0.2.0","packages":[{"target":"linux-armv7","file_name":"sparkedge-v0.2.0-linux-armv7.zip","sha256":"abc123","size":42}]}`,
	}})

	t.Setenv("SPARKEDGE_VERSION", "v0.1.0")
	t.Setenv("SPARKEDGE_TARGET_LABEL", "linux-armv7")

	result, err := service.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatalf("expected update to be available: %#v", result)
	}
	if !result.IntegrityReady {
		t.Fatalf("expected integrity metadata to be available: %#v", result)
	}
	if result.CompatibleAsset == nil || result.CompatibleAsset.Name != "sparkedge-v0.2.0-linux-armv7.zip" {
		t.Fatalf("unexpected compatible asset: %#v", result.CompatibleAsset)
	}
}

func TestCheckHandlesNonSemverLocalVersion(t *testing.T) {
	service := NewService(Config{
		Enabled: true,
		Repo:    "kelwSagashi/SparkEdgeGo",
	}, fakeReleaseClient{releases: []Release{
		{
			Version: "v0.2.0",
			Assets: []Asset{
				{Name: "sparkedge-v0.2.0-linux-amd64.zip", DownloadURL: "https://example.invalid/linux-amd64.zip"},
			},
		},
	}})

	t.Setenv("SPARKEDGE_VERSION", "dev")
	t.Setenv("SPARKEDGE_TARGET_LABEL", "linux-amd64")

	result, err := service.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CanCompare {
		t.Fatalf("expected comparison to be unavailable: %#v", result)
	}
	if result.UpdateAvailable {
		t.Fatalf("did not expect update_available for non-semver local version: %#v", result)
	}
}

func TestDownloadLatestVerifiesChecksum(t *testing.T) {
	payload := "zip-content"
	hash := sha256.Sum256([]byte(payload))
	expectedHash := hex.EncodeToString(hash[:])
	service := NewService(Config{
		Enabled: true,
		Repo:    "kelwSagashi/SparkEdgeGo",
	}, fakeReleaseClient{
		releases: []Release{
			{
				Version: "v0.2.0",
				HTMLURL: "https://example.invalid/release",
				Assets: []Asset{
					{Name: "manifest.json", DownloadURL: "https://example.invalid/manifest.json"},
					{Name: "sparkedge-v0.2.0-linux-amd64.zip", DownloadURL: "https://example.invalid/linux-amd64.zip"},
				},
			},
		},
		downloads: map[string]string{
			"https://example.invalid/manifest.json":   `{"version":"v0.2.0","packages":[{"target":"linux-amd64","file_name":"sparkedge-v0.2.0-linux-amd64.zip","sha256":"` + expectedHash + `","size":11}]}`,
			"https://example.invalid/linux-amd64.zip": payload,
		},
	})

	t.Setenv("SPARKEDGE_VERSION", "v0.1.0")
	t.Setenv("SPARKEDGE_TARGET_LABEL", "linux-amd64")
	t.Setenv("SPARKEDGE_HOME", t.TempDir())

	result, err := service.DownloadLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ChecksumVerified {
		t.Fatalf("expected checksum verification: %#v", result)
	}
	if result.SHA256 != expectedHash {
		t.Fatalf("unexpected checksum %q", result.SHA256)
	}
	if !strings.HasSuffix(result.DownloadedPath, "sparkedge-v0.2.0-linux-amd64.zip") {
		t.Fatalf("unexpected download path: %s", result.DownloadedPath)
	}
}

func TestDownloadLatestFailsWhenChecksumDoesNotMatch(t *testing.T) {
	service := NewService(Config{
		Enabled: true,
		Repo:    "kelwSagashi/SparkEdgeGo",
	}, fakeReleaseClient{
		releases: []Release{
			{
				Version: "v0.2.0",
				Assets: []Asset{
					{Name: "manifest.json", DownloadURL: "https://example.invalid/manifest.json"},
					{Name: "sparkedge-v0.2.0-linux-amd64.zip", DownloadURL: "https://example.invalid/linux-amd64.zip"},
				},
			},
		},
		downloads: map[string]string{
			"https://example.invalid/manifest.json":   `{"version":"v0.2.0","packages":[{"target":"linux-amd64","file_name":"sparkedge-v0.2.0-linux-amd64.zip","sha256":"deadbeef","size":11}]}`,
			"https://example.invalid/linux-amd64.zip": "zip-content",
		},
	})

	t.Setenv("SPARKEDGE_VERSION", "v0.1.0")
	t.Setenv("SPARKEDGE_TARGET_LABEL", "linux-amd64")
	t.Setenv("SPARKEDGE_HOME", t.TempDir())

	_, err := service.DownloadLatest(context.Background())
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
}

func TestRestartPlanUsesServiceNameWhenConfigured(t *testing.T) {
	service := NewService(Config{
		ServiceName: "sparkedge",
	}, fakeReleaseClient{})

	result, err := service.Restart(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ManualRequired {
		t.Fatalf("expected manual plan result: %#v", result)
	}
	if !strings.Contains(result.Command, "sparkedge") {
		t.Fatalf("unexpected restart command: %q", result.Command)
	}
}

type fakeReleaseClient struct {
	releases  []Release
	downloads map[string]string
}

func (f fakeReleaseClient) ListReleases(_ context.Context, _ string) ([]Release, error) {
	return f.releases, nil
}

func (f fakeReleaseClient) OpenAsset(_ context.Context, downloadURL string) (io.ReadCloser, error) {
	value, ok := f.downloads[downloadURL]
	if !ok {
		return nil, errors.New("missing download stub")
	}
	return io.NopCloser(strings.NewReader(value)), nil
}
