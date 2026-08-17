param(
    [ValidateSet("install", "update", "uninstall")]
    [string]$Command = "install",
    [string]$Target = (Get-Location).Path,
    [string]$Source = $env:AI_FLOW_SOURCE,
    [ValidateSet("core")]
    [string]$Profile = "core",
    [switch]$Cursor,
    [switch]$Codex,
    [switch]$Claude,
    [switch]$All,
    [string]$Platforms = $env:AI_FLOW_PLATFORMS
)

$ErrorActionPreference = "Stop"
$PackVersion = "0.2.2"
$CoreSkills = @(
    "initialize-ai-project",
    "orchestrate-ai-delivery",
    "adopt-existing-project",
    "discover-product-goal",
    "plan-product-delivery",
    "profile-project-engineering",
    "research-and-design-solution",
    "specify-tests",
    "implement-work-item",
    "diagnose-and-verify",
    "review-change",
    "integrate-git-change",
    "manage-release",
    "sync-project-knowledge"
)

$SelectedPlatforms = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)
if (-not [string]::IsNullOrWhiteSpace($Platforms)) {
    foreach ($platform in ($Platforms -split '[,\s]+')) {
        if ([string]::IsNullOrWhiteSpace($platform)) { continue }
        switch ($platform.ToLowerInvariant()) {
            "cursor" { [void]$SelectedPlatforms.Add("cursor") }
            "codex" { [void]$SelectedPlatforms.Add("codex") }
            "claude" { [void]$SelectedPlatforms.Add("claude") }
            "claude-code" { [void]$SelectedPlatforms.Add("claude") }
            "all" {
                [void]$SelectedPlatforms.Add("cursor")
                [void]$SelectedPlatforms.Add("codex")
                [void]$SelectedPlatforms.Add("claude")
            }
            default { throw "Unsupported platform: $platform" }
        }
    }
}
if ($Cursor) { [void]$SelectedPlatforms.Add("cursor") }
if ($Codex) { [void]$SelectedPlatforms.Add("codex") }
if ($Claude) { [void]$SelectedPlatforms.Add("claude") }
if ($All) {
    [void]$SelectedPlatforms.Add("cursor")
    [void]$SelectedPlatforms.Add("codex")
    [void]$SelectedPlatforms.Add("claude")
}
$PlatformSelectionExplicit = $SelectedPlatforms.Count -gt 0

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
        ".ai-flow/install/platforms",
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
if ($Command -eq "uninstall" -and $PlatformSelectionExplicit) {
    throw "Platform switches are not supported with uninstall; uninstall removes the complete managed installation"
}

if ($SelectedPlatforms.Count -eq 0) {
    $existingPlatformsFile = Join-Path $TargetPath ".ai-flow/install/platforms"
    if ($Command -eq "update" -and (Test-Path -LiteralPath $existingPlatformsFile -PathType Leaf)) {
        foreach ($platform in (Get-Content -LiteralPath $existingPlatformsFile)) {
            if ($platform -in @("cursor", "codex", "claude")) { [void]$SelectedPlatforms.Add($platform) }
        }
    } else {
        [void]$SelectedPlatforms.Add("cursor")
        [void]$SelectedPlatforms.Add("codex")
        [void]$SelectedPlatforms.Add("claude")
    }
}
$SelectCursor = $SelectedPlatforms.Contains("cursor")
$SelectCodex = $SelectedPlatforms.Contains("codex")
$SelectClaude = $SelectedPlatforms.Contains("claude")
$SelectedSkillRoots = @()
if ($SelectCodex) { $SelectedSkillRoots += ".agents/skills" }
if ($SelectCursor) { $SelectedSkillRoots += ".cursor/skills" }
if ($SelectClaude) { $SelectedSkillRoots += ".claude/skills" }

if ($Command -eq "install" -and -not (Test-Path -LiteralPath $InstallMarker -PathType Leaf)) {
    foreach ($skillName in $CoreSkills) {
        foreach ($skillRoot in $SelectedSkillRoots) {
            if (Test-Path -LiteralPath (Join-Path $TargetPath "$skillRoot/$skillName")) {
                throw "Existing unmanaged Skill would be overwritten: $skillRoot/$skillName"
            }
        }
    }
    if ($SelectClaude -and (Test-Path -LiteralPath (Join-Path $TargetPath ".claude/skills/ai-flow"))) {
        throw "Existing unmanaged Claude ai-flow entry would be overwritten"
    }
    if ($SelectCursor -and (Test-Path -LiteralPath (Join-Path $TargetPath ".cursor/rules/ai-flow.mdc"))) {
        throw "Existing unmanaged Cursor ai-flow rule would be overwritten"
    }
}

if (Test-Path -LiteralPath $InstallMarker -PathType Leaf) {
    $installedPackVersion = (Get-Content -LiteralPath $InstallMarker | Select-Object -First 1).Trim()
    foreach ($skillName in $CoreSkills) {
        foreach ($skillRoot in $SelectedSkillRoots) {
            $nativeSkill = Join-Path $TargetPath "$skillRoot/$skillName"
            $managedMarker = Join-Path $nativeSkill ".ai-flow-managed"
            if ((Test-Path -LiteralPath $nativeSkill) -and -not (Test-Path -LiteralPath $managedMarker -PathType Leaf)) {
                if (-not ($skillRoot -eq ".agents/skills" -and $installedPackVersion -eq "0.1.0")) {
                    throw "Existing unmanaged native Skill would be overwritten: $skillRoot/$skillName"
                }
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
$buildFromSource = $env:AI_FLOW_BUILD_SOURCE -eq "1"
if ($buildFromSource -or -not (Test-Path -LiteralPath $runtimePath -PathType Leaf)) {
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($null -eq $goCommand) { throw "No compatible flowctl binary found and Go is unavailable" }
    $buildDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ai-flow-" + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $buildDir | Out-Null
    $runtimePath = Join-Path $buildDir "flowctl.exe"
    Push-Location $SourcePath
    try { & go build -o $runtimePath ./cmd/flowctl } finally { Pop-Location }
}

$directories = @(".ai-flow/bin", ".ai-flow/install", ".ai-flow/runtime")
if ($SelectCodex) { $directories += ".agents/skills" }
if ($SelectCursor) { $directories += @(".cursor/skills", ".cursor/rules") }
if ($SelectClaude) { $directories += ".claude/skills" }
foreach ($relativePath in $directories) {
    New-Item -ItemType Directory -Path (Join-Path $TargetPath $relativePath) -Force | Out-Null
}

foreach ($skillName in $CoreSkills) {
    $sourceSkill = Join-Path $SourcePath "skills/$skillName"
    if (-not (Test-Path -LiteralPath (Join-Path $sourceSkill "SKILL.md"))) { throw "Missing source Skill: $skillName" }
    foreach ($skillRoot in $SelectedSkillRoots) {
        $targetSkill = Join-Path $TargetPath "$skillRoot/$skillName"
        if (Test-Path -LiteralPath $targetSkill) { Remove-Item -LiteralPath $targetSkill -Recurse -Force }
        Copy-Item -LiteralPath $sourceSkill -Destination $targetSkill -Recurse
        Set-Content -LiteralPath (Join-Path $targetSkill ".ai-flow-managed") -Value $PackVersion -Encoding utf8
    }
}

if ($SelectClaude) {
    $claudeTarget = Join-Path $TargetPath ".claude/skills/ai-flow"
    if (Test-Path -LiteralPath $claudeTarget) { Remove-Item -LiteralPath $claudeTarget -Recurse -Force }
    Copy-Item -LiteralPath (Join-Path $SourcePath "adapters/claude/ai-flow") -Destination $claudeTarget -Recurse
    Set-Content -LiteralPath (Join-Path $claudeTarget ".ai-flow-managed") -Value $PackVersion -Encoding utf8
}
if ($SelectCursor) {
    Copy-Item -LiteralPath (Join-Path $SourcePath "adapters/cursor/ai-flow.mdc") -Destination (Join-Path $TargetPath ".cursor/rules/ai-flow.mdc") -Force
}
Copy-Item -LiteralPath $runtimePath -Destination (Join-Path $TargetPath ".ai-flow/bin/flowctl.exe") -Force
$schemaTarget = Join-Path $TargetPath ".ai-flow/runtime/schemas"
if (Test-Path -LiteralPath $schemaTarget) { Remove-Item -LiteralPath $schemaTarget -Recurse -Force }
Copy-Item -LiteralPath (Join-Path $SourcePath "schemas") -Destination $schemaTarget -Recurse

if ($SelectCodex) { Set-AIFlowBlock (Join-Path $TargetPath "AGENTS.md") (Join-Path $SourcePath "adapters/codex/AGENTS.block.md") }
if ($SelectClaude) { Set-AIFlowBlock (Join-Path $TargetPath "CLAUDE.md") (Join-Path $SourcePath "adapters/claude/CLAUDE.block.md") }

Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/install/version") -Value $PackVersion -Encoding utf8
Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/install/profile") -Value $Profile -Encoding utf8
$ActivePlatforms = [System.Collections.Generic.HashSet[string]]::new($SelectedPlatforms, [System.StringComparer]::OrdinalIgnoreCase)
$platformFile = Join-Path $TargetPath ".ai-flow/install/platforms"
if (Test-Path -LiteralPath $platformFile -PathType Leaf) {
    foreach ($platform in (Get-Content -LiteralPath $platformFile)) {
        if ($platform -in @("cursor", "codex", "claude")) { [void]$ActivePlatforms.Add($platform) }
    }
} else {
    if (Test-Path -LiteralPath (Join-Path $TargetPath ".cursor/rules/ai-flow.mdc")) { [void]$ActivePlatforms.Add("cursor") }
    $agentsPath = Join-Path $TargetPath "AGENTS.md"
    if ((Test-Path -LiteralPath $agentsPath) -and (Select-String -LiteralPath $agentsPath -SimpleMatch '<!-- ai-flow:start -->' -Quiet)) { [void]$ActivePlatforms.Add("codex") }
    if (Test-Path -LiteralPath (Join-Path $TargetPath ".claude/skills/ai-flow/SKILL.md")) { [void]$ActivePlatforms.Add("claude") }
}
$activePlatformLines = @("cursor", "codex", "claude") | Where-Object { $ActivePlatforms.Contains($_) }
Set-Content -LiteralPath $platformFile -Value $activePlatformLines -Encoding utf8
$cursorState = [int]($ActivePlatforms.Contains("cursor"))
$codexState = [int]($ActivePlatforms.Contains("codex"))
$claudeState = [int]($ActivePlatforms.Contains("claude"))
Set-Content -LiteralPath (Join-Path $TargetPath ".ai-flow/capabilities.yaml") -Value @(
    "schema_version: 1",
    "profile: $Profile",
    "platforms:",
    "  cursor: $cursorState",
    "  codex: $codexState",
    "  claude_code: $claudeState"
) -Encoding utf8

& (Join-Path $TargetPath ".ai-flow/bin/flowctl.exe") doctor --root $TargetPath
Write-Host "AI Flow $PackVersion $Command completed at $TargetPath"
Write-Host "Next: reload the IDE window, start a new Agent chat, then ask to initialize the project or invoke initialize-ai-project directly."
