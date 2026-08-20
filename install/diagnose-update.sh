#!/bin/sh
# Diagnose which install paths are reachable from this machine.
# Read-only: never writes anything, never installs.

REPOSITORY="${AI_FLOW_REPOSITORY:-mxm-web-develop/coding-skills}"
VERSION="${AI_FLOW_VERSION:-main}"

probe() {
  label="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    printf '  [OK]   %s\n' "$label"
    return 0
  fi
  printf '  [FAIL] %s\n' "$label"
  return 1
}

section() {
  printf '\n— %s —\n' "$1"
}

have_curl=0
have_git=0
have_ssh=0
command -v curl >/dev/null 2>&1 && have_curl=1
command -v git  >/dev/null 2>&1 && have_git=1
command -v ssh  >/dev/null 2>&1 && have_ssh=1

printf 'AI Flow 安装连通性诊断\n'
printf '目标仓库：%s\n' "$REPOSITORY"
printf '探测版本：%s（可用 AI_FLOW_VERSION=vX.Y.Z 指定特定版本）\n' "$VERSION"
printf '工具：curl=%s git=%s ssh=%s\n' "$have_curl" "$have_git" "$have_ssh"

section 'HTTPS 路径'
https_paths_ok=0
if [ "$have_curl" -eq 1 ]; then
  probe "https://raw.githubusercontent.com/$REPOSITORY/main/install/bootstrap.sh" \
    sh -c "curl -fsSL --max-time 8 -o /dev/null https://raw.githubusercontent.com/$REPOSITORY/main/install/bootstrap.sh" && https_paths_ok=$((https_paths_ok + 1))

  if printf '%s' "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
    probe "https://github.com/$REPOSITORY/releases/download/$VERSION/coding-skills.tar.gz" \
      sh -c "curl -fsSL --max-time 8 -o /dev/null https://github.com/$REPOSITORY/releases/download/$VERSION/coding-skills.tar.gz" && https_paths_ok=$((https_paths_ok + 1))

    probe "https://codeload.github.com/$REPOSITORY/tar.gz/refs/tags/$VERSION" \
      sh -c "curl -fsSL --max-time 8 -o /dev/null https://codeload.github.com/$REPOSITORY/tar.gz/refs/tags/$VERSION" && https_paths_ok=$((https_paths_ok + 1))
  else
    printf '  [SKIP] 版本 "%s" 不是 vX.Y.Z，跳过 release / codeload 探测\n' "$VERSION"
  fi
else
  printf '  [SKIP] curl 未安装，跳过 HTTPS 探测\n'
fi

section 'git+SSH 路径'
ssh_ok=0
if [ "$have_git" -eq 1 ] && [ "$have_ssh" -eq 1 ]; then
  if [ "$VERSION" = "main" ] || ! printf '%s' "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
    probe "git+ssh://git@github.com/$REPOSITORY.git  refs/heads/main" \
      sh -c "git ls-remote --heads git@github.com:$REPOSITORY.git refs/heads/main" && ssh_ok=$((ssh_ok + 1))
  else
    probe "git+ssh://git@github.com/$REPOSITORY.git  refs/tags/$VERSION" \
      sh -c "git ls-remote git@github.com:$REPOSITORY.git refs/tags/$VERSION" && ssh_ok=$((ssh_ok + 1))
  fi
else
  printf '  [SKIP] git 或 ssh 未安装，跳过 SSH 探测\n'
fi

section '结论'
if [ "$https_paths_ok" -gt 0 ]; then
  printf '  HTTPS 可用 → 正常用 curl | sh 即可：\n'
  printf '    curl -fsSL https://raw.githubusercontent.com/%s/main/install/bootstrap.sh | sh\n' "$REPOSITORY"
fi
if [ "$ssh_ok" -gt 0 ]; then
  printf '  SSH 可用 → 绕过 HTTPS 直接拉源码：\n'
  if printf '%s' "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
    printf '    git clone --depth 1 --branch %s git@github.com:%s.git /tmp/coding-skills\n' "$VERSION" "$REPOSITORY"
  else
    printf '    git clone --depth 1 git@github.com:%s.git /tmp/coding-skills\n' "$REPOSITORY"
  fi
  printf '    cd /tmp/coding-skills && bash install/install.sh update --target <你的项目路径> --source .\n'
fi
if [ "$https_paths_ok" -eq 0 ] && [ "$ssh_ok" -eq 0 ]; then
  printf '  HTTPS 和 SSH 都不可用 → 浏览器手抄：\n'
  printf '    浏览器打开 https://github.com/%s/releases/tag/%s\n' "$REPOSITORY" "$VERSION"
  printf '    下载 assets 里的 coding-skills.tar.gz 和 checksums.txt\n'
  printf '    解压后跑 bash install/install.sh update --target <你的项目路径> --source .\n'
fi
