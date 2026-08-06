param(
    [string]$TargetOS,
    [string]$TargetArch,
    [string]$GoArm = "",
    [string]$Version = "dev",
    [string]$OutputRoot = "dist/packages",
    [switch]$SkipWebUIBuild
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

function Get-DefaultTargetOS {
    if ($IsWindows) { return "windows" }
    if ($IsMacOS) { return "darwin" }
    return "linux"
}

function Get-TargetLabel {
    param(
        [string]$OSName,
        [string]$ArchName,
        [string]$ArmValue
    )

    if ($ArchName -eq "arm" -and $ArmValue) {
        return "$OSName-$ArchName`v$ArmValue"
    }
    return "$OSName-$ArchName"
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $repoRoot

if (-not $TargetOS) {
    $TargetOS = Get-DefaultTargetOS
}
if (-not $TargetArch) {
    $TargetArch = "amd64"
}

$targetLabel = Get-TargetLabel -OSName $TargetOS -ArchName $TargetArch -ArmValue $GoArm
$goExe = Get-GoExe
$npmExe = Get-NpmExe
$packageRoot = Join-Path $repoRoot $OutputRoot
$stagingRoot = Join-Path $packageRoot "staging"
$artifactDir = Join-Path $stagingRoot $targetLabel
$appDir = Join-Path $artifactDir "sparkEdge"
$webuiDistDir = Join-Path $repoRoot "webui/dist"
$archivePath = Join-Path $packageRoot ("sparkedge-{0}-{1}.zip" -f $Version, $targetLabel)
$binaryName = if ($TargetOS -eq "windows") { "sparkedge.exe" } else { "sparkedge" }

if (-not $SkipWebUIBuild) {
    Push-Location (Join-Path $repoRoot "webui")
    & $npmExe run build
    Pop-Location
}

if (-not (Test-Path $webuiDistDir)) {
    throw "webui/dist not found. Build the webui before packaging."
}

if (Test-Path $artifactDir) {
    Remove-Item -Recurse -Force $artifactDir
}
New-Item -ItemType Directory -Force $appDir | Out-Null
New-Item -ItemType Directory -Force (Join-Path $appDir "webui") | Out-Null
New-Item -ItemType Directory -Force (Join-Path $appDir "config") | Out-Null

$originalGOOS = $env:GOOS
$originalGOARCH = $env:GOARCH
$originalGOARM = $env:GOARM
$originalCGO = $env:CGO_ENABLED
$originalGOCACHE = $env:GOCACHE
$originalGOMODCACHE = $env:GOMODCACHE

try {
    $env:GOOS = $TargetOS
    $env:GOARCH = $TargetArch
    $env:CGO_ENABLED = "0"
    if ($GoArm) {
        $env:GOARM = $GoArm
    } else {
        Remove-Item Env:GOARM -ErrorAction SilentlyContinue
    }

    $goCacheDir = Join-Path $repoRoot ".codex-tmp/go-build"
    $goModCacheDir = Join-Path $repoRoot ".codex-tmp/go-mod"
    New-Item -ItemType Directory -Force $goCacheDir | Out-Null
    New-Item -ItemType Directory -Force $goModCacheDir | Out-Null
    $env:GOCACHE = $goCacheDir
    $env:GOMODCACHE = $goModCacheDir

    & $goExe build -o (Join-Path $appDir $binaryName) ./cmd/sparkedge-api
}
finally {
    $env:GOOS = $originalGOOS
    $env:GOARCH = $originalGOARCH
    $env:CGO_ENABLED = $originalCGO
    $env:GOCACHE = $originalGOCACHE
    $env:GOMODCACHE = $originalGOMODCACHE
    if ($originalGOARM) {
        $env:GOARM = $originalGOARM
    } else {
        Remove-Item Env:GOARM -ErrorAction SilentlyContinue
    }
}

Copy-Item -Recurse -Force $webuiDistDir (Join-Path $appDir "webui")
Copy-Item -Force (Join-Path $repoRoot "README.md") (Join-Path $appDir "README.md")

@"
JWT_SECRET=change-me
SPARKEDGE_HTTP_ADDR=:3009
SPARKEDGE_WEBUI_DIST=./webui/dist
SPARKEDGE_SAMPLES_DIR=
SPARK_CLOUD_URL=
"@ | Set-Content -Encoding utf8 (Join-Path $appDir "config/.env.example")

@"
Version: $Version
Target: $targetLabel
BuiltAtUtc: $((Get-Date).ToUniversalTime().ToString("o"))
"@ | Set-Content -Encoding utf8 (Join-Path $appDir "version.txt")

if (Test-Path $archivePath) {
    Remove-Item -Force $archivePath
}
Compress-Archive -Path $appDir -DestinationPath $archivePath -CompressionLevel Optimal

Write-Host "Package generated at $archivePath"
