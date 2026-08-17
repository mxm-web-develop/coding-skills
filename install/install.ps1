param(
    [ValidateSet("install", "update", "uninstall")]
    [string]$Command = "install",
    [string]$Target = (Get-Location).Path,
    [string]$Source = $env:AI_FLOW_SOURCE,
    [ValidateSet("core")]
    [string]$Profile = "core"
)

$ErrorActionPreference = "Stop"
$PackVersion = "0.1.1"
$CoreSkills = @(
    "initialize-ai-project",
    "orchestrate-ai-delivery",
    "adopt-existing-project",
    "discover-product-goal",
    "plan-product-delivery",
    "research-and-design-solution",
    "specify-tests",
    "implement-work-item",
    "diagnose-and-verify",
    "review-change",
    "integrate-git-change",
    "manage-release",
    "sync-project-knowledge"
)

function Remove-AIFlowBlock([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return }
    $lines = Get-Content -LiteralPath $Path
    $skip = $false
    $kept = foreach ($line in $lines) {
        if ($line -eq "<!-- ai-flow:start -->") { $skip = $true; continue }
        if ($line -eq "<!-- ai-flow:end -->") { $skip = $false; continue }
        if (-not $skip) { $line }
    }
    Set-Content -LiteralPath $Path -Value $kept -Encoding utf8
}

function Set-AIFlowBlock([string]$Path, [string]$BlockPath) {
    $parent = Split-Path -Parent $Path
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    if (-not (Test-Path -LiteralPath $Path)) { New-Item -ItemType File -Path $Path | Out-Null }
    Remove-AIFlowBlock $Path
    Add-Content -LiteralPath $Path -Value ""
    Add-Content -LiteralPath $Path -Value (Get-Content -LiteralPath $BlockPath)
}

$TargetPath = (Resolve-Path -LiteralPath $Target).Path
if ($TargetPath -eq [System.IO.Path]::GetPathRoot($TargetPath)) {
    throw "Refusing to install at filesystem root"
}

if ([string]::IsNullOrWhiteSpace($Source)) {
    $Source = Split-Path -Parent $PSScriptRoot
}
$SourcePath = (Resolve-Path -LiteralPath $Source).Path

function Remove-ManagedFiles {
    foreach ($skillName in $CoreSkills) {
        foreach ($skillRoot in @(".agents/skills", ".cursor/skills", ".claude/skills")) {
            $skillPath = Join-Path $TargetPath "$skillRoot/$skillName"
            if (Test-Path -LiteralPath $skillPath) { Remove-Item -LiteralPath $skillPath -Recurse -Force }
        }
    }
    $managedPaths = @(
        ".claude/skills/ai-flow",
        ".cursor/rules/ai-flow.mdc",
        ".ai-flow/bin/flowctl.exe",
        ".ai-flow/bin/flowctl",
        ".ai-flow/install/version",
        ".ai-flow/install/profile",
        ".ai-flow/runtime/schemas"
    )
    foreach ($relativePath in $managedPaths) {
        $path = Join-Path $TargetPath $relativePath
        if (Test-Path -LiteralPath $path) { Remove-Item -LiteralPath $path -Recurse -Force }
    }
    Remove-AIFlowBlock (Join-Path $TargetPath "AGENTS.md")
    Remove-AIFlowBlock (Join-Path $TargetPath "CLAUDE.md")
}

$InstallMarker = Join-Path $TargetPath ".ai-flow/install/version"
if (($Command -eq "update" -or $Command -eq "uninstall") -and -not (Test-Path -LiteralPath $InstallMarker -PathType Leaf)) {
    throw "No managed AI Flow installation found at target"
}

if ($Command -eq "install" -and -not (Test-Path -LiteralPath $InstallMarker -PathType Leaf)) {
    foreach ($skillName in $CoreSkills) {
        foreach ($skillRoot in @(".agents/skills", ".cursor/skills", ".claude/skills")) {
            if (Test-Path -LiteralPath (Join-Path $TargetPath "$skillRoot/$skillName")) {
                throw "Existing unmanaged Skill would be overwritten: $skillRoot/$skillName"
            }
        }
    }
    if (Test-Path -LiteralPath (Join-Path $TargetPath ".claude/skills/ai-flow")) {
        throw "Existing unmanaged Claude ai-flow entry would be overwritten"
    }
    if (Test-Path -LiteralPath (Join-Path $TargetPath ".cursor/rules/ai-flow.mdc")) {
        throw "Existing unmanaged Cursor ai-flow rule would be overwritten"
    }
}

if (Test-Path -LiteralPath $InstallMarker -PathType Leaf) {
    foreach ($skillName in $CoreSkills) {
        foreach ($skillRoot in @(".cursor/skills", ".claude/skills")) {
            $nativeSkill = Join-Path $TargetPath "$skillRoot/$skillName"
            $managedMarker = Join-Path $nativeSkill ".ai-flow-managed"
            if ((Test-Path -LiteralPath $nativeSkill) -and -not (Test-Path -LiteralPath $managedMarker -PathType Leaf)) {
                throw "Existing unmanaged native Skill would be overwritten: $skillRoot/$skillName"
            }
        }
    }
}

if ($Command -eq "uninstall") {
    Remove-ManagedFiles
    Write-Host "Removed AI Flow managed runtime, Skills, and platform entries from $TargetPath"
    Write-Host "Project state under .ai-flow and human reports under docs/board were preserved."
    exit 0
}

if (-not (Test-Path -LiteralPath (Join-Path $SourcePath "skills") -PathType Container)) {
    throw "Source has no skills directory; use -Source or AI_FLOW_SOURCE"
}

$architecture = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}
$runtimePath = Join-Path $SourcePath "dist/flowctl-windows-$architecture.exe"
if (-not (Test-Path -LiteralPath $runtimePath -PathType Leaf)) {
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($null -eq $goCommand) { throw "No compatible flowctl binary found and Go is unavailable" }
    $buildDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ai-flow-" + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $buildDir | Out-Null
    $runtimePath = Join-Path $buildDir "flowctl.exe"
    Push-Location $SourcePath
    try { & go build -o $runtimePath ./cmd/flowctl } finally { Pop-Location }
}

$directories = @(".agents/skills", ".cursor/skills", ".claude/skills", ".cursor/rules", ".ai-flow/bin", ".ai-flow/install", ".ai-flow/runtime")
foreach ($relativePath in $directories) {
    New-Item -ItemType Directory -Path (Join-Path $TargetPath $relativePath) -Force | Out-Null
}

foreach ($skillName in $CoreSkills) {
    $sourceSkill = Join-Path $SourcePath "skills/$skillName"
    if (-not (Test-Path -LiteralPath (Join-Path $sourceSkill "SKILL.md"))) { throw "Missing source Skill: $skillName" }
    foreach ($skillRoot in @(".agents/skills", ".cursor/skills", ".claude/skills")) {
        $targetSkill = Join-Path $TargetPath "$skillRoot/$skillName"
        if (Test-Path -LiteralPath $targetSkill) { Remove-Item -LiteralPath $targetSkill -Recurse -Force }
        Copy-Item -LiteralPath $sourceSkill -Destination $targetSkill -Recurse
        Set-Content -LiteralPath (Join-Path $targetSkill ".ai-flow-managed") -Value $PackVersion -Encoding utf8
    }
}

$claudeTarget = Join-Path $TargetPath ".claude/skills/ai-flow"
if (Test-Path -LiteralPath $claudeTarget) { Remove-Item -LiteralPath $claudeTarget -Recurse -Force }
Copy-Item -LiteralPath (Join-Path $SourcePath "adapters/claude/ai-flow") -Destination $claudeTarget -Recurse
Set-Content -LiteralPath (Join-Path $claudeTarget ".ai-flow-managed") -Value $PackVersion -Encoding utf8
Copy-Item -LiteralPath (Join-Path $SourcePath "adapters/cursor/ai-flow.mdc") -Destination (Join-Path $TargetPath ".cursor/rules/ai-flow.mdc") -Force
Copy-Item -LiteralPath $runtimePath -Destination (Join-Path $TargetPath ".ai-flow/bin/flowctl.exe") -Force
$schemaTarget = Join-Path $TargetPath ".ai-flow/runtime/schemas"
if (Test-Path -LiteralPath $schemaTarget) { Remove-Item -LiteralPath $schemaTarget -Recurse -Force }
Copy-Item -LiteralPath (Join-Path $SourcePath "schemas") -Destination $schemaTarget -Recurse

Set-AIFlowBlock (Join-Path $TargetPath "AGENTS.md") (Join-Path $SourcePath "adapters/codex/AGENTS.block.md")
Set-AIFlowBlock (Join-Path $TargetPath "CLAUDE.md") (Join-Path $SourcePath "adapters/claude/CLAUDE.block.md")

Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/install/version") -Value $PackVersion -Encoding utf8
Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/install/profile") -Value $Profile -Encoding utf8
Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/capabilities.yaml") -Value @(
    "schema_version: 1",
    "profile: $Profile",
    "platforms:",
    "  cursor: detected",
    "  codex: detected",
    "  claude_code: adapter"
) -Encoding utf8

& (Join-Path $TargetPath ".ai-flow/bin/flowctl.exe") doctor --root $TargetPath
Write-Host "AI Flow $PackVersion $Command completed at $TargetPath"
Write-Host "Next: reload the IDE window, start a new Agent chat, then ask to initialize the project or invoke initialize-ai-project directly."
