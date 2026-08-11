param(
    [string]$Version = "dev",
    [switch]$SkipGoTests,
    [switch]$SkipWebUIBuild,
    [switch]$SkipPackaging,
    [switch]$IncludeArm32
)

$ErrorActionPreference = "Stop"

function Get-GoExe {
    if ($env:GO_EXE) {
        return $env:GO_EXE
    }
    if ($IsWindows -and (Test-Path "C:\Program Files\Go\bin\go.exe")) {
        return "C:\Program Files\Go\bin\go.exe"
    }
    return "go"
}

function Get-NpmExe {
    if ($IsWindows) {
        return "npm.cmd"
    }
    return "npm"
}

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name" -ForegroundColor Cyan
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw "Step failed: $Name (exit code $LASTEXITCODE)"
    }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$goExe = Get-GoExe
$npmExe = Get-NpmExe

Set-Location $repoRoot

$originalGOCACHE = $env:GOCACHE
$originalGOMODCACHE = $env:GOMODCACHE
$goCacheDir = Join-Path $repoRoot ".codex-tmp/go-build"
$goModCacheDir = Join-Path $repoRoot ".codex-tmp/go-mod"
New-Item -ItemType Directory -Force $goCacheDir | Out-Null
New-Item -ItemType Directory -Force $goModCacheDir | Out-Null
$env:GOCACHE = $goCacheDir
$env:GOMODCACHE = $goModCacheDir

$targets = @(
    @{ os = "windows"; arch = "amd64"; goarm = "" },
    @{ os = "windows"; arch = "arm64"; goarm = "" },
    @{ os = "linux"; arch = "amd64"; goarm = "" },
    @{ os = "linux"; arch = "arm64"; goarm = "" },
    @{ os = "darwin"; arch = "amd64"; goarm = "" },
    @{ os = "darwin"; arch = "arm64"; goarm = "" }
)

if ($IncludeArm32) {
    $targets += @(
        @{ os = "linux"; arch = "arm"; goarm = "7" },
        @{ os = "linux"; arch = "arm"; goarm = "6" }
    )
}

if (-not $SkipGoTests) {
    Invoke-Step "Running Go test suite" {
        & $goExe test ./...
    }
}

if (-not $SkipWebUIBuild) {
    Invoke-Step "Building webui" {
        Push-Location (Join-Path $repoRoot "webui")
        try {
            & $npmExe run build
        }
        finally {
            Pop-Location
        }
    }
}

if (-not $SkipPackaging) {
    foreach ($target in $targets) {
        $label = if ($target.goarm) {
            "$($target.os)-$($target.arch)v$($target.goarm)"
        } else {
            "$($target.os)-$($target.arch)"
        }

        Invoke-Step "Packaging $label" {
            $params = @{
                TargetOS = $target.os
                TargetArch = $target.arch
                Version = $Version
                SkipWebUIBuild = $true
            }
            if ($target.goarm) {
                $params.GoArm = $target.goarm
            }

            & (Join-Path $repoRoot "scripts/build-release.ps1") @params
        }
    }

    Invoke-Step "Generating manifest and checksums" {
        & (Join-Path $repoRoot "scripts/generate-release-manifest.ps1") -Version $Version
    }
}

$env:GOCACHE = $originalGOCACHE
$env:GOMODCACHE = $originalGOMODCACHE

Write-Host ""
Write-Host "Release validation completed successfully." -ForegroundColor Green
