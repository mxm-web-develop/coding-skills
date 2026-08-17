$ErrorActionPreference = "Stop"
$env:AI_FLOW_BUILD_SOURCE = "1"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "../..")).Path
$forwardRoot = Join-Path $env:RUNNER_TEMP ("ai-flow-forward-" + [guid]::NewGuid())
$reverseRoot = Join-Path $env:RUNNER_TEMP ("ai-flow-reverse-" + [guid]::NewGuid())

function Assert-Path([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path)) { throw "Expected path: $Path" }
}

function Assert-AllPlatforms([string]$Root) {
    $platforms = @((Get-Content -LiteralPath (Join-Path $Root ".ai-flow/install/platforms")))
    if (($platforms -join ",") -ne "cursor,codex,claude") {
        throw "Unexpected platform set: $($platforms -join ',')"
    }
    Assert-Path (Join-Path $Root ".cursor/skills/initialize-ai-project/SKILL.md")
    Assert-Path (Join-Path $Root ".agents/skills/initialize-ai-project/SKILL.md")
    Assert-Path (Join-Path $Root ".claude/skills/initialize-ai-project/SKILL.md")
    Assert-Path (Join-Path $Root ".ai-flow/bin/flowctl.exe")
}

try {
    New-Item -ItemType Directory -Path $forwardRoot, $reverseRoot | Out-Null

    & (Join-Path $repoRoot "install/install.ps1") -Command install -Cursor -Target $forwardRoot -Source $repoRoot | Out-Null
    & (Join-Path $repoRoot "install/install.ps1") -Command install -Codex -Target $forwardRoot -Source $repoRoot | Out-Null
    & (Join-Path $repoRoot "install/install.ps1") -Command install -Claude -Target $forwardRoot -Source $repoRoot | Out-Null
    Assert-AllPlatforms $forwardRoot

    & (Join-Path $repoRoot "install/install.ps1") -Command install -Claude -Target $reverseRoot -Source $repoRoot | Out-Null
    & (Join-Path $repoRoot "install/install.ps1") -Command install -Codex -Target $reverseRoot -Source $repoRoot | Out-Null
    & (Join-Path $repoRoot "install/install.ps1") -Command install -Cursor -Target $reverseRoot -Source $repoRoot | Out-Null
    Assert-AllPlatforms $reverseRoot

    & (Join-Path $reverseRoot ".ai-flow/bin/flowctl.exe") project init --root $reverseRoot --mode greenfield --name "PowerShell Multi IDE" | Out-Null
    Assert-Path (Join-Path $reverseRoot "docs/board/STATUS.md")
    Write-Host "AI Flow PowerShell platform selection E2E passed"
}
finally {
    if (Test-Path -LiteralPath $forwardRoot) { Remove-Item -LiteralPath $forwardRoot -Recurse -Force }
    if (Test-Path -LiteralPath $reverseRoot) { Remove-Item -LiteralPath $reverseRoot -Recurse -Force }
}
