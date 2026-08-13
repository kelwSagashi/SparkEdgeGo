# Assisted Update Flow

This document explains how the assisted update flow works in `SparkEdgeGo`.

The updater uses GitHub Releases as the official distribution source and is designed to keep the process safe and user-driven.

## Goal

The assisted update flow lets the operator:

- check whether a new version exists;
- see the current local version and target platform;
- inspect the latest compatible release;
- download the matching asset for the current platform;
- prepare the update with backup and staging;
- rollback if needed;
- trigger or plan a restart.

The first stages are intentionally conservative. The system does not silently replace the running binary.

## Update source

The updater reads releases from GitHub and resolves the correct asset for the current platform.

Release compatibility is based on:

- semantic version tags in the form `vX.Y.Z`;
- the local target platform;
- the optional manifest embedded in the release;
- checksum validation when available.

## Configuration

The updater is configured through `config.yml`.

Relevant fields:

```yaml
update:
  enabled: true
  provider: github
  repo: kelwSagashi/SparkEdgeGo
  channel: stable
  allow_prerelease: false
  service_name: ""
  restart_command: ""
```

### Field meaning

- `enabled`: turns the updater on or off
- `provider`: current update provider, normally `github`
- `repo`: GitHub repository that publishes the releases
- `channel`: update channel, usually `stable`
- `allow_prerelease`: allows prerelease versions when enabled
- `service_name`: optional service name used to suggest restart commands
- `restart_command`: optional custom restart command

## Environment overrides

The runtime also honors environment-based overrides for update settings.

The update system is part of the general config priority chain:

```text
config.yml > environment variables > built-in defaults
```

## HTTP endpoints

The updater is exposed through the local HTTP API.

### Check current state

`GET /api/update/check`

Returns:

- current version
- current target
- latest compatible release
- compatible asset
- checksum readiness
- whether an update is available

### Read persisted state

`GET /api/update/status`

Returns:

- last downloaded package
- last prepared version
- last prepared target
- last apply result
- last rollback result
- last restart result
- update history

### Download update

`POST /api/update/download`

Downloads the compatible asset for the current platform, validates its checksum, and stores it in the local update workspace.

### Apply update

`POST /api/update/apply`

Request body:

```json
{
  "downloaded_path": "path/to/downloaded/package"
}
```

This prepares the update by:

- creating a backup;
- staging the new package;
- validating the expected structure;
- recording the next steps for the operator.

### Rollback

`POST /api/update/rollback`

Restores the previously prepared or applied state when a rollback artifact is available.

### Restart

`POST /api/update/restart`

Request body:

```json
{
  "execute": true
}
```

If `execute` is `true`, the system attempts to run the restart command.
If `execute` is `false`, it only prepares the restart plan.

## WebUI behavior

The update page in the WebUI uses the endpoints above to show:

- current version
- current target
- compatibility status
- release metadata
- download result
- apply result
- rollback result
- restart result
- persisted history

The main page is:

- `/settings/update`

## Compatibility resolution

The updater resolves the most recent compatible release using these steps:

1. load releases from GitHub;
2. ignore non-semver tags;
3. ignore prereleases unless allowed by config;
4. look for `manifest.json` when present;
5. match the current target to the correct asset;
6. fall back to the direct asset name pattern if no manifest is available;
7. compare versions using semver;
8. mark the update as available only when the remote version is newer.

## Manifest and checksum

When a release includes `manifest.json`, the updater can use it to:

- resolve the asset for the current target more reliably;
- validate the expected SHA256 checksum;
- reduce dependence on asset naming only.

If the manifest or checksum is not available, the updater still reports the compatible release, but integrity support is reduced.

## Local storage

The updater persists state so the UI can show historical actions.

Typical persisted data includes:

- last downloaded package path;
- last prepared version and target;
- latest apply details;
- latest rollback details;
- latest restart details;
- action history with timestamps.

## Safe operating model

The update flow is designed to be cautious:

- the user always initiates the download and apply steps;
- checksum verification happens before the package is accepted;
- backup is created before changing files;
- rollback remains available as long as the prepared state exists;
- restart can be only planned if the user prefers.

## Recommended release naming

For best compatibility, releases should follow:

```text
vMAJOR.MINOR.PATCH
```

Example:

- `v0.1.0`
- `v1.0.0`
- `v1.2.3`

## Relevant source files

- `SparkEdgeGo/internal/updater/service.go`
- `SparkEdgeGo/internal/httpapi/update_handlers.go`
- `SparkEdgeGo/webui/src/pages/UpdateSettings.tsx`
- `SparkEdgeGo/webui/src/rest-api-client/update.service.ts`
- `SparkEdgeGo/docs/assisted-update-plan.md`
- `SparkEdgeGo/docs/release-checklist.md`

