param(
    [ValidateSet("install", "update")]
    [string]$Command = $(if ($env:AI_FLOW_COMMAND) { $env:AI_FLOW_COMMAND } else { "install" }),
    [string]$Version = $(if ($env:AI_FLOW_VERSION) { $env:AI_FLOW_VERSION } else { "latest" }),
    [string]$Target = $(if ($env:AI_FLOW_TARGET) { $env:AI_FLOW_TARGET } else { (Get-Location).Path })
)

$ErrorActionPreference = "Stop"
$Repository = "mxm-web-develop/coding-skills"

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
    Invoke-WebRequest -Uri "$downloadBase/coding-skills.zip" -OutFile $archive
    Invoke-WebRequest -Uri "$downloadBase/checksums.txt" -OutFile $checksums

    $checksumLine = Get-Content -LiteralPath $checksums | Where-Object { $_ -match 'coding-skills\.zip$' } | Select-Object -First 1
    if ($null -eq $checksumLine) { throw "Release checksum does not contain coding-skills.zip" }
    $expected = ($checksumLine -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expected -ne $actual) { throw "Release checksum mismatch" }

    Expand-Archive -LiteralPath $archive -DestinationPath $bootstrapDir
    $sourceDir = Join-Path $bootstrapDir "coding-skills"
    $installer = Join-Path $sourceDir "install/install.ps1"
    if (-not (Test-Path -LiteralPath $installer -PathType Leaf)) { throw "Release package is incomplete" }
    & $installer -Command $Command -Target $targetPath -Source $sourceDir -Profile core
} finally {
    if (Test-Path -LiteralPath $bootstrapDir) { Remove-Item -LiteralPath $bootstrapDir -Recurse -Force }
}
