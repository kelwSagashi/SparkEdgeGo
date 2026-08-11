param(
    [Parameter(Mandatory = $true)]
    [string]$Version,
    [string]$PackagesDir = "dist/packages"
)

$ErrorActionPreference = "Stop"

$resolvedPackagesDir = (Resolve-Path $PackagesDir).Path
$zipFiles = Get-ChildItem -Path $resolvedPackagesDir -Filter "*.zip" | Sort-Object Name

if (-not $zipFiles) {
    throw "No release packages were found in $resolvedPackagesDir"
}

$packages = @()
$checksums = @()

foreach ($file in $zipFiles) {
    if ($file.BaseName -notmatch "^sparkedge-(v\d+\.\d+\.\d+)-(.+)$") {
        continue
    }

    $fileVersion = $Matches[1]
    $target = $Matches[2]

    if ($fileVersion -ne $Version) {
        continue
    }

    $hash = (Get-FileHash -Path $file.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
    $packages += [pscustomobject]@{
        target    = $target
        file_name = $file.Name
        sha256    = $hash
        size      = [int64]$file.Length
    }
    $checksums += "$hash  $($file.Name)"
}

if (-not $packages) {
    throw "No release packages matching version $Version were found in $resolvedPackagesDir"
}

$manifest = [pscustomobject]@{
    version      = $Version
    generated_at = (Get-Date).ToUniversalTime().ToString("o")
    packages     = $packages
}

$manifestPath = Join-Path $resolvedPackagesDir "manifest.json"
$checksumsPath = Join-Path $resolvedPackagesDir "checksums.txt"

$manifest | ConvertTo-Json -Depth 5 | Set-Content -Encoding utf8 $manifestPath
$checksums | Set-Content -Encoding utf8 $checksumsPath

Write-Host "Manifest generated at $manifestPath"
Write-Host "Checksums generated at $checksumsPath"
