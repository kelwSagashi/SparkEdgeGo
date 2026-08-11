package updater

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/appfs"
	"github.com/kelwSagashi/sparkedge-go/internal/appmeta"
)

var packageNamePattern = regexp.MustCompile(`^sparkedge-(v\d+\.\d+\.\d+)-(.+)\.zip$`)

type ApplyResult struct {
	Version         string   `json:"version"`
	Target          string   `json:"target"`
	DownloadedPath  string   `json:"downloaded_path"`
	StagingPath     string   `json:"staging_path"`
	BackupPath      string   `json:"backup_path"`
	ScriptPath      string   `json:"script_path,omitempty"`
	RollbackPath    string   `json:"rollback_path,omitempty"`
	Applied         bool     `json:"applied"`
	PreparedOnly    bool     `json:"prepared_only"`
	RestartRequired bool     `json:"restart_required"`
	Message         string   `json:"message"`
	AppliedFiles    []string `json:"applied_files"`
	PreservedFiles  []string `json:"preserved_files"`
	NextSteps       []string `json:"next_steps"`
}

type packageDescriptor struct {
	Version string
	Target  string
}

func (s *Service) ApplyDownloaded(_ context.Context, downloadedPath string) (ApplyResult, error) {
	downloadedPath = strings.TrimSpace(downloadedPath)
	if downloadedPath == "" {
		return ApplyResult{}, errors.New("downloaded_path is required")
	}

	descriptor, err := parsePackageDescriptor(downloadedPath)
	if err != nil {
		return ApplyResult{}, err
	}

	current := appmeta.LoadVersionInfo()
	if current.Target != "" && descriptor.Target != current.Target {
		return ApplyResult{}, fmt.Errorf("download target %s does not match current target %s", descriptor.Target, current.Target)
	}

	appRoot := appfs.AppRoot()
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	stageDir := filepath.Join(appRoot, "updates", "staging", descriptor.Version+"-"+timestamp)
	extractedRoot := filepath.Join(stageDir, "extracted")
	packageRoot := filepath.Join(extractedRoot, "sparkEdge")
	backupDir := filepath.Join(appRoot, "updates", "backups", descriptor.Version+"-"+timestamp)

	if err := os.MkdirAll(extractedRoot, 0o755); err != nil {
		return ApplyResult{}, err
	}
	if err := unzipArchive(downloadedPath, extractedRoot); err != nil {
		return ApplyResult{}, err
	}
	if err := validateExtractedPackage(packageRoot, descriptor.Target); err != nil {
		return ApplyResult{}, err
	}
	if err := backupCurrentInstallation(appRoot, backupDir, descriptor.Target); err != nil {
		return ApplyResult{}, err
	}

	preserved := []string{"config.yml", "sparkedge.db"}
	previous, _ := s.LoadState()
	result := ApplyResult{
		Version:         descriptor.Version,
		Target:          descriptor.Target,
		DownloadedPath:  downloadedPath,
		StagingPath:     packageRoot,
		BackupPath:      backupDir,
		RestartRequired: true,
		PreservedFiles:  preserved,
	}

	if strings.HasPrefix(descriptor.Target, "windows-") {
		scriptPath, rollbackPath, err := writeWindowsScripts(packageRoot, appRoot, backupDir, descriptor.Target)
		if err != nil {
			return ApplyResult{}, err
		}
		result.PreparedOnly = true
		result.ScriptPath = scriptPath
		result.RollbackPath = rollbackPath
		result.Message = "Atualizacao preparada. Pare o SparkEdge e execute o script de aplicacao para concluir no Windows."
		result.NextSteps = []string{
			"Pare o processo atual do SparkEdge.",
			"Execute o script de aplicacao gerado.",
			"Inicie novamente o SparkEdge apos a troca.",
			"Se algo falhar, execute o script de rollback.",
		}
		_ = s.saveStateWithHistory(previous, UpdateState{
			LastDownloadedPackage: downloadedPath,
			LastPreparedVersion:   descriptor.Version,
			LastPreparedTarget:    descriptor.Target,
			LastApplyResult:       &result,
		}, HistoryEntry{
			Type:      "apply",
			Status:    "prepared",
			Version:   descriptor.Version,
			Target:    descriptor.Target,
			Message:   result.Message,
			Artifact:  downloadedPath,
			CreatedAt: time.Now().UTC(),
		})
		return result, nil
	}

	appliedFiles, err := applyPackageFiles(packageRoot, appRoot, descriptor.Target)
	if err != nil {
		return ApplyResult{}, err
	}
	rollbackPath, err := writeUnixRollbackScript(appRoot, backupDir, descriptor.Target)
	if err != nil {
		return ApplyResult{}, err
	}
	result.Applied = true
	result.AppliedFiles = appliedFiles
	result.RollbackPath = rollbackPath
	result.Message = "Arquivos da nova versao aplicados. Reinicie o SparkEdge para carregar o binario e a WebUI atualizados."
	result.NextSteps = []string{
		"Reinicie o SparkEdge para carregar a nova versao.",
		"Valide a WebUI e as rotas principais apos o restart.",
		"Se houver problema, use o script de rollback gerado.",
	}
	_ = s.saveStateWithHistory(previous, UpdateState{
		LastDownloadedPackage: downloadedPath,
		LastPreparedVersion:   descriptor.Version,
		LastPreparedTarget:    descriptor.Target,
		LastApplyResult:       &result,
	}, HistoryEntry{
		Type:      "apply",
		Status:    "applied",
		Version:   descriptor.Version,
		Target:    descriptor.Target,
		Message:   result.Message,
		Artifact:  downloadedPath,
		CreatedAt: time.Now().UTC(),
	})
	return result, nil
}

func parsePackageDescriptor(downloadedPath string) (packageDescriptor, error) {
	base := filepath.Base(downloadedPath)
	match := packageNamePattern.FindStringSubmatch(base)
	if match == nil {
		return packageDescriptor{}, errors.New("downloaded package name is invalid")
	}
	return packageDescriptor{
		Version: strings.TrimSpace(match[1]),
		Target:  strings.TrimSpace(match[2]),
	}, nil
}

func unzipArchive(zipPath string, destination string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	root := filepath.Clean(destination) + string(os.PathSeparator)
	for _, file := range reader.File {
		targetPath := filepath.Join(destination, file.Name)
		cleanTarget := filepath.Clean(targetPath)
		if !strings.HasPrefix(cleanTarget, root) {
			return fmt.Errorf("unsafe path in archive: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return err
		}

		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close()
		srcErr := src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if srcErr != nil {
			return srcErr
		}
	}

	return nil
}

func validateExtractedPackage(packageRoot string, target string) error {
	required := []string{
		filepath.Join(packageRoot, "README.md"),
		filepath.Join(packageRoot, "version.txt"),
		filepath.Join(packageRoot, "config.yml"),
		filepath.Join(packageRoot, "webui", "dist", "index.html"),
		filepath.Join(packageRoot, executableNameForTarget(target)),
	}
	for _, path := range required {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return fmt.Errorf("invalid extracted package: missing %s", path)
		}
	}
	return nil
}

func backupCurrentInstallation(appRoot string, backupDir string, target string) error {
	for _, relative := range replaceableRelativePaths(target) {
		source := filepath.Join(appRoot, relative)
		if _, err := os.Stat(source); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}

		destination := filepath.Join(backupDir, relative)
		if err := copyPath(source, destination); err != nil {
			return err
		}
	}

	for _, optional := range []string{"config.yml"} {
		source := filepath.Join(appRoot, optional)
		if _, err := os.Stat(source); err == nil {
			if err := copyPath(source, filepath.Join(backupDir, optional)); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyPackageFiles(packageRoot string, appRoot string, target string) ([]string, error) {
	relativePaths := replaceableRelativePaths(target)
	applied := make([]string, 0, len(relativePaths))
	for _, relative := range relativePaths {
		source := filepath.Join(packageRoot, relative)
		destination := filepath.Join(appRoot, relative)
		if err := os.RemoveAll(destination); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err := copyPath(source, destination); err != nil {
			return nil, err
		}
		applied = append(applied, relative)
	}
	return applied, nil
}

func copyPath(source string, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDirectory(source, destination)
	}
	return copyFile(source, destination, info.Mode())
}

func copyDirectory(source string, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(source string, destination string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func replaceableRelativePaths(target string) []string {
	return []string{
		executableNameForTarget(target),
		"README.md",
		"version.txt",
		filepath.Join("config", ".env.example"),
		"webui",
	}
}

func executableNameForTarget(target string) string {
	if strings.HasPrefix(strings.TrimSpace(target), "windows-") {
		return "sparkedge.exe"
	}
	return "sparkedge"
}

func writeWindowsScripts(packageRoot string, appRoot string, backupDir string, target string) (string, string, error) {
	scriptDir := filepath.Dir(packageRoot)
	scriptPath := filepath.Join(scriptDir, "apply-update.ps1")
	rollbackPath := filepath.Join(scriptDir, "rollback-update.ps1")
	lines := []string{
		"$ErrorActionPreference = 'Stop'",
		"$sourceRoot = '" + strings.ReplaceAll(packageRoot, "'", "''") + "'",
		"$targetRoot = '" + strings.ReplaceAll(appRoot, "'", "''") + "'",
		"$backupRoot = '" + strings.ReplaceAll(backupDir, "'", "''") + "'",
		"",
		"Write-Host 'Applying SparkEdge update...'",
	}
	for _, relative := range replaceableRelativePaths(target) {
		escapedRelative := strings.ReplaceAll(relative, "\\", "\\\\")
		lines = append(lines,
			fmt.Sprintf("if (Test-Path (Join-Path $targetRoot '%s')) { Remove-Item -Recurse -Force (Join-Path $targetRoot '%s') }", escapedRelative, escapedRelative),
			fmt.Sprintf("Copy-Item -Recurse -Force (Join-Path $sourceRoot '%s') (Join-Path $targetRoot '%s')", escapedRelative, escapedRelative),
		)
	}
	lines = append(lines,
		"Write-Host 'Update applied. Backup preserved at:' $backupRoot",
	)

	content := strings.Join(lines, "\r\n") + "\r\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		return "", "", err
	}

	rollbackLines := []string{
		"$ErrorActionPreference = 'Stop'",
		"$backupRoot = '" + strings.ReplaceAll(backupDir, "'", "''") + "'",
		"$targetRoot = '" + strings.ReplaceAll(appRoot, "'", "''") + "'",
		"",
		"Write-Host 'Rolling back SparkEdge update...'",
	}
	for _, relative := range replaceableRelativePaths(target) {
		escapedRelative := strings.ReplaceAll(relative, "\\", "\\\\")
		rollbackLines = append(rollbackLines,
			fmt.Sprintf("if (Test-Path (Join-Path $backupRoot '%s')) {", escapedRelative),
			fmt.Sprintf("  if (Test-Path (Join-Path $targetRoot '%s')) { Remove-Item -Recurse -Force (Join-Path $targetRoot '%s') }", escapedRelative, escapedRelative),
			fmt.Sprintf("  Copy-Item -Recurse -Force (Join-Path $backupRoot '%s') (Join-Path $targetRoot '%s')", escapedRelative, escapedRelative),
			"}",
		)
	}
	rollbackLines = append(rollbackLines, "Write-Host 'Rollback completed.'")
	rollbackContent := strings.Join(rollbackLines, "\r\n") + "\r\n"
	if err := os.WriteFile(rollbackPath, []byte(rollbackContent), 0o644); err != nil {
		return "", "", err
	}

	return scriptPath, rollbackPath, nil
}

func writeUnixRollbackScript(appRoot string, backupDir string, target string) (string, error) {
	scriptPath := filepath.Join(backupDir, "rollback-update.sh")
	lines := []string{
		"#!/usr/bin/env sh",
		"set -eu",
		"",
		"backup_root='" + escapeSingleQuotes(backupDir) + "'",
		"target_root='" + escapeSingleQuotes(appRoot) + "'",
		"",
		"echo 'Rolling back SparkEdge update...'",
	}
	for _, relative := range replaceableRelativePaths(target) {
		lines = append(lines,
			"if [ -e \"$backup_root/"+relative+"\" ]; then",
			"  rm -rf \"$target_root/"+relative+"\"",
			"  cp -R \"$backup_root/"+relative+"\" \"$target_root/"+relative+"\"",
			"fi",
		)
	}
	lines = append(lines, "echo 'Rollback completed.'")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
		return "", err
	}
	return scriptPath, nil
}

func escapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "'\"'\"'")
}
