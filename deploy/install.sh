#!/usr/bin/env bash
# Stellaris Planet 代理一键安装脚本。
# One-shot installer for the Stellaris Planet agent.
#
# 用法：
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Veni222987/stellaris/main/deploy/install.sh | sh
#   sudo bash deploy/install.sh
#
# 环境变量：
# Environment variables:
#   STELLARIS_VERSION  指定版本，默认 latest（pin a release version, default: latest）
#   STELLARIS_PREFIX   安装前缀，默认 /usr/local（installation prefix, default: /usr/local）
#   STELLARIS_MIRROR   GitHub release 镜像前缀（mirror URL prefix for GitHub releases）

set -euo pipefail

VERSION="${STELLARIS_VERSION:-latest}"
PREFIX="${STELLARIS_PREFIX:-/usr/local}"
MIRROR="${STELLARIS_MIRROR:-https://github.com/stellaris-dev/stellaris/releases/download}"

detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "ERROR: 不支持的 CPU 架构: $(uname -m)" >&2; exit 1 ;;
    esac
    case "$os" in
        linux|darwin) echo "${os}-${arch}" ;;
        *) echo "ERROR: 不支持的操作系统: $os" >&2; exit 1 ;;
    esac
}

PLATFORM="$(detect_platform)"
TARBALL="stellaris-cli-${VERSION}-${PLATFORM}.tar.gz"
URL="${MIRROR}/${VERSION}/${TARBALL}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "[install] 平台: $PLATFORM"
echo "[install] 版本: $VERSION"
echo "[install] 下载: $URL"

curl -fsSL "$URL" -o "$TMP/$TARBALL"
if curl -fsSL "${URL}.sha256" -o "$TMP/${TARBALL}.sha256" 2>/dev/null; then
    (cd "$TMP" && sha256sum -c "${TARBALL}.sha256")
    echo "[install] ✓ sha256 校验通过"
else
    echo "WARN: 找不到 .sha256 校验文件，跳过校验。生产部署务必启用！" >&2
fi

tar -xzf "$TMP/$TARBALL" -C "$TMP"
# 在解压目录中定位二进制
# Locate binary within extracted archive
BIN="$(find "$TMP" -name stellaris-cli -type f | head -1)"
[ -n "$BIN" ] || { echo "ERROR: tar 内没找到 stellaris-cli" >&2; exit 1; }
install -m 0755 "$BIN" "$PREFIX/bin/stellaris-cli"
echo "[install] ✓ 已安装 $PREFIX/bin/stellaris-cli"

# Linux root 环境自动写入 systemd unit 文件（不自动启用）
# Install systemd unit on Linux root (not auto-enabled)
if [ "$(uname -s)" = "Linux" ] && [ "$(id -u)" = "0" ] && command -v systemctl >/dev/null; then
    cat > /etc/systemd/system/stellaris-cli.service <<EOF
[Unit]
Description=Stellaris Planet 代理
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$PREFIX/bin/stellaris-cli start
ExecStop=$PREFIX/bin/stellaris-cli stop
Restart=on-failure
RestartSec=5
User=root
Environment=STELLARIS_CONFIG_DIR=/etc/stellaris

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    echo "[install] ✓ systemd unit 已就位（未自动 enable）。"
    echo "          先执行 'stellaris-cli orbit ...' 加入星系，再 'systemctl enable --now stellaris-cli'。"
fi

cat <<EOF

下一步：
  1. 加入星系：  $PREFIX/bin/stellaris-cli orbit <ip>:<port> <gid> --token <node-token>
  2. 声明 Agent：$PREFIX/bin/stellaris-cli agent add <name> --type <openclaw|hermes|workbuddy> --binary <path>
  3. 启动守护：  $PREFIX/bin/stellaris-cli start  （或 systemctl enable --now stellaris-cli）
EOF
