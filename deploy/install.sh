#!/usr/bin/env bash
# Stellaris Planet 代理一键安装脚本。
# One-shot installer for the Stellaris Planet agent.
#
# 用法：
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Veni222987/stellaris/main/deploy/install.sh | sh
#
# 环境变量：
# Environment variables:
#   STELLARIS_VERSION  指定版本，默认 latest（pin a release version, default: latest）
#   STELLARIS_BIN_DIR  安装目录，默认 ~/.local/bin（遵循 XDG，无需 sudo）
#   STELLARIS_MIRROR   GitHub release 镜像前缀（mirror URL prefix for GitHub releases）

set -euo pipefail

VERSION="${STELLARIS_VERSION:-}"
# 优先用 XDG_BIN_HOME，否则 ~/.local/bin（XDG 惯例）
_DEFAULT_BIN="${XDG_BIN_HOME:-${XDG_DATA_HOME:-$HOME/.local/share}/../bin}"
BIN_DIR="${STELLARIS_BIN_DIR:-$_DEFAULT_BIN}"
# 规范化路径（去掉 ..）
BIN_DIR="$(cd "$(dirname "$BIN_DIR/x")" 2>/dev/null && pwd || echo "$BIN_DIR")"
REPO="Veni222987/stellaris"
MIRROR="${STELLARIS_MIRROR:-https://github.com/${REPO}/releases/download}"

# VERSION 为空时通过 latest 重定向拿实际 tag（无需 API token，不受匿名限速影响）
resolve_version() {
    local url tag
    url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
        "https://github.com/${REPO}/releases/latest")"
    tag="${url##*/}"
    [ -n "$tag" ] && [ "$tag" != "latest" ] || \
        { echo "ERROR: 无法解析最新版本号，请手动设置 STELLARIS_VERSION" >&2; exit 1; }
    echo "$tag"
}

[ -n "$VERSION" ] || VERSION="$(resolve_version)"

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
EXTRACTED_BIN="$(find "$TMP" -name stellaris-cli -type f | head -1)"
[ -n "$EXTRACTED_BIN" ] || { echo "ERROR: tar 内没找到 stellaris-cli" >&2; exit 1; }

mkdir -p "$BIN_DIR"
cp "$EXTRACTED_BIN" "$BIN_DIR/stellaris-cli"
chmod 0755 "$BIN_DIR/stellaris-cli"
echo "[install] ✓ 已安装 $BIN_DIR/stellaris-cli"

# Linux 下检查 BIN_DIR 是否在 PATH 里，不在则提示
if [ "$(uname -s)" = "Linux" ] && command -v systemctl >/dev/null && [ "$(id -u)" = "0" ]; then
    cat > /etc/systemd/system/stellaris-cli.service <<EOF
[Unit]
Description=Stellaris Planet 代理
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$BIN_DIR/stellaris-cli start
ExecStop=$BIN_DIR/stellaris-cli stop
Restart=on-failure
RestartSec=5
Environment=STELLARIS_CONFIG_DIR=/etc/stellaris

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    echo "[install] ✓ systemd unit 已就位（未自动 enable）。"
    echo "          先执行 'stellaris-cli orbit ...' 加入星系，再 'systemctl enable --now stellaris-cli'。"
fi

# 检查 BIN_DIR 是否在 PATH 里
case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) echo "WARN: $BIN_DIR 不在 PATH 中，请在 shell 配置文件里添加：" >&2
       echo "      export PATH=\"\$PATH:$BIN_DIR\"" >&2 ;;
esac

cat <<EOF

下一步：
  1. 加入星系：  stellaris-cli orbit <ip>:<port> <gid> --token <node-token>
  2. 声明 Agent：stellaris-cli agent add <name> --type <openclaw|hermes|workbuddy> --binary <path>
  3. 启动守护：  stellaris-cli start  （或 systemctl enable --now stellaris-cli）
EOF
