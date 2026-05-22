#!/usr/bin/env bash
# Stellaris Planet 代理卸载脚本。
# Uninstaller for the Stellaris Planet agent.
set -euo pipefail

PREFIX="${STELLARIS_PREFIX:-/usr/local}"

if [ "$(uname -s)" = "Linux" ] && command -v systemctl >/dev/null; then
    if systemctl list-unit-files | grep -q "^stellaris-cli.service"; then
        systemctl disable --now stellaris-cli 2>/dev/null || true
        rm -f /etc/systemd/system/stellaris-cli.service
        systemctl daemon-reload
        echo "[uninstall] ✓ 已移除 systemd unit"
    fi
fi

rm -f "$PREFIX/bin/stellaris-cli"
echo "[uninstall] ✓ 已删除 $PREFIX/bin/stellaris-cli"
echo "[uninstall] 配置目录未自动清理：~/.stellaris 与 /etc/stellaris"
echo "            如确认不再需要，请手动 rm -rf 删除。"
