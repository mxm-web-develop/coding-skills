param(
    [ValidateSet("install", "update")]
    [string]$Command = $(if ($env:AI_FLOW_COMMAND) { $env:AI_FLOW_COMMAND } else { "install" }),
    [string]$Version = $(if ($env:AI_FLOW_VERSION) { $env:AI_FLOW_VERSION } else { "latest" }),
    [string]$Target = $(if ($env:AI_FLOW_TARGET) { $env:AI_FLOW_TARGET } else { (Get-Location).Path }),
    [switch]$Cursor,
    [switch]$Codex,
    [switch]$Claude,
    [switch]$All,
    [string]$Platforms = $env:AI_FLOW_PLATFORMS
)

$ErrorActionPreference = "Stop"
$Repository = "mxm-web-develop/coding-skills"
$platformValues = @()
if (-not [string]::IsNullOrWhiteSpace($Platforms)) { $platformValues += $Platforms }
if ($Cursor) { $platformValues += "cursor" }
if ($Codex) { $platformValues += "codex" }
if ($Claude) { $platformValues += "claude" }
if ($All) { $platformValues += "all" }
$platformSelection = $platformValues -join ","

if ($Version -eq "latest") {
    $downloadBase = "https://github.com/$Repository/releases/latest/download"
} elseif ($Version -match '^v[0-9]+\.[0-9]+\.[0-9]+$') {
    $downloadBase = "https://github.com/$Repository/releases/download/$Version"
} else {
    throw "Version must be latest or a vX.Y.Z tag"
}

$targetPath = (Resolve-Path -LiteralPath $Target).Path
$bootstrapDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ai-flow-bootstrap-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $bootstrapDir | Out-Null

try {
    $archive = Join-Path $bootstrapDir "coding-skills.zip"
    $checksums = Join-Path $bootstrapDir "checksums.txt"
    Write-Host "ai-flow bootstrap: downloading AI Flow $Version release package"
    Invoke-WebRequest -Uri "$downloadBase/coding-skills.zip" -OutFile $archive
    Invoke-WebRequest -Uri "$downloadBase/checksums.txt" -OutFile $checksums

    Write-Host "ai-flow bootstrap: verifying release checksum"
    $checksumLine = Get-Content -LiteralPath $checksums | Where-Object { $_ -match 'coding-skills\.zip$' } | Select-Object -First 1
    if ($null -eq $checksumLine) { throw "Release checksum does not contain coding-skills.zip" }
    $expected = ($checksumLine -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expected -ne $actual) { throw "Release checksum mismatch" }

    Write-Host "ai-flow bootstrap: extracting release package"
    Expand-Archive -LiteralPath $archive -DestinationPath $bootstrapDir
    $sourceDir = Join-Path $bootstrapDir "coding-skills"
    $installer = Join-Path $sourceDir "install/install.ps1"
    if (-not (Test-Path -LiteralPath $installer -PathType Leaf)) { throw "Release package is incomplete" }
    Write-Host "ai-flow bootstrap: installing into $targetPath"
    & $installer -Command $Command -Target $targetPath -Source $sourceDir -Profile core -Platforms $platformSelection
} finally {
    if (Test-Path -LiteralPath $bootstrapDir) { Remove-Item -LiteralPath $bootstrapDir -Recurse -Force }
}
