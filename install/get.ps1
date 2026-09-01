<#
AI Flow smart entry for Windows PowerShell. Mirrors install/get.sh: tries
raw.githubusercontent.com first, falls back to $AI_FLOW_DOWNLOAD_MIRROR (default
https://ghproxy.com) on failure. Override mirror or set
AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR=1 to skip the fallback entirely.
#>
[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]] $BootstrapArgs = @()
)

$ErrorActionPreference = 'Stop'
$repo = if ($env:AI_FLOW_REPOSITORY) { $env:AI_FLOW_REPOSITORY } else { 'mxm-web-develop/coding-skills' }
$mirror = if ($env:AI_FLOW_DOWNLOAD_MIRROR) { $env:AI_FLOW_DOWNLOAD_MIRROR } else { 'https://ghproxy.com' }
$skipMirror = ($env:AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR -eq '1')

$rawUrl = "https://raw.githubusercontent.com/$repo/main/install/bootstrap.ps1"
$mirrorUrl = "$mirror/$rawUrl"

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("ai-flow-get-" + [guid]::NewGuid() + ".ps1")

function Get-AiFlowBootstrapScript {
    param([string] $Url, [int] $TimeoutSec)
    try {
        Invoke-WebRequest -Uri $Url -OutFile $tmp -TimeoutSec $TimeoutSec -UseBasicParsing -ErrorAction Stop
        return $true
    } catch {
        return $false
    }
}

if (Get-AiFlowBootstrapScript -Url $rawUrl -TimeoutSec 12) {
    # fast path succeeded
} elseif ($skipMirror) {
    Write-Host "ai-flow: 直连 raw.githubusercontent.com 失败，且 AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR=1 已禁用镜像。" -ForegroundColor Yellow
    Write-Host "请在浏览器打开下面的 URL，把脚本内容存成本地文件再执行："
    Write-Host "  $rawUrl"
    exit 1
} elseif (Get-AiFlowBootstrapScript -Url $mirrorUrl -TimeoutSec 25) {
    Write-Host "ai-flow: 直连 raw.githubusercontent.com 不通，自动改走镜像 $mirror"
} else {
    Write-Host "ai-flow: 直连和镜像都无法下载 bootstrap.ps1。" -ForegroundColor Yellow
    Write-Host "请在浏览器打开下面任一链接，把脚本内容存成本地文件再执行（推荐先跑 install/diagnose-update.sh 看本机哪些路径通）："
    Write-Host "  $rawUrl"
    Write-Host "  $mirrorUrl"
    exit 1
}

& $tmp @BootstrapArgs
