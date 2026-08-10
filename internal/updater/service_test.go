package updater

import (
	"context"
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
				{Name: "sparkedge-v0.2.0-linux-armv7.zip", DownloadURL: "https://example.invalid/linux-armv7.zip", Size: 42},
			},
			PublishedAt: time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC),
		},
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

type fakeReleaseClient struct {
	releases []Release
}

func (f fakeReleaseClient) ListReleases(_ context.Context, _ string) ([]Release, error) {
	return f.releases, nil
}
