#!/usr/bin/env bash
set -Eeuo pipefail

# 仅供隔离安装验收调用，所有安装路径由测试创建，使用独立网络命名空间。
: "${YZ_TEST_ROOT:?缺少测试目录}"
: "${YZ_TEST_ARTIFACTS:?缺少构建目录}"
: "${YZ_TEST_SERVICE:?缺少测试服务名}"
: "${YZ_TEST_NETNS:?缺少测试网络命名空间}"
: "${YZ_TEST_TOKEN:?缺少临时面板凭据}"

test_root=$(realpath "$YZ_TEST_ROOT")
case "$test_root" in
    /var/tmp/yzboard-singbox-test-*) ;;
    *) echo "测试目录不在允许范围内" >&2; exit 1 ;;
esac
[[ "$YZ_TEST_SERVICE" =~ ^yzboard-singbox-test-[a-f0-9]+$ ]]
[[ -f "$test_root/.owned-test-instance" ]]
[[ "$(readlink /proc/self/ns/net)" != "$(readlink /proc/1/ns/net)" ]]
[[ "$(readlink "$YZ_TEST_NETNS")" == "$(readlink /proc/self/ns/net)" ]]

sed '/^main "$@"$/d' "$YZ_TEST_ARTIFACTS/install.sh" >"$test_root/install-library.sh"
# shellcheck source=/dev/null
source "$test_root/install-library.sh"

INSTALL_ROOT="$test_root/instance"
BACKUP_DIR="$INSTALL_ROOT/backups"
INSTALL_META="$INSTALL_ROOT/install-meta.json"
CONFIG_FILE="$INSTALL_ROOT/config.yml"
CREDENTIALS_FILE="$INSTALL_ROOT/credentials.env"
BINARY_PATH="$test_root/bin/xboard-node"
CLI_PATH="$test_root/bin/xbctl"
CLI_SYMLINK_PATH="$test_root/bin/xbctl-link"
INSTALLER_COPY_PATH="$INSTALL_ROOT/install.sh"
SERVICE_NAME="$YZ_TEST_SERVICE"
SYSTEMD_SERVICE_NAME="$SERVICE_NAME.service"
SYSTEMD_SERVICE_PATH="/run/systemd/system/$SYSTEMD_SERVICE_NAME"
SERVICE_PATH="$SYSTEMD_SERVICE_PATH"
SERVICE_MANAGER=systemd
mkdir -p "$test_root/bin" "$test_root/tmp"
export TMPDIR="$test_root/tmp"

# 复用正式服务渲染，仅给测试服务加入命名空间并关闭敏感日志的持久化。
eval "$(declare -f render_service | sed '1s/render_service/render_service_base/')"
render_service() {
    render_service_base
    sed -i \
        -e "/^\[Service\]$/a NetworkNamespacePath=$YZ_TEST_NETNS" \
        -e '/^\[Service\]$/a MemoryMax=768M' \
        -e '/^\[Service\]$/a CPUQuota=100%' \
        -e 's/^StandardOutput=.*/StandardOutput=null/' \
        -e 's/^StandardError=.*/StandardError=null/' \
        "$TMP_DIR/service"
}

service_enable() { systemctl enable --runtime "$SYSTEMD_SERVICE_NAME"; }
service_disable() { systemctl disable --runtime "$SYSTEMD_SERVICE_NAME"; }

parse_args install --mode node --kernel singbox --panel "$YZ_TEST_PANEL" \
    --token "$YZ_TEST_TOKEN" --node-id "$YZ_TEST_NODE_ID" --node-type socks \
    --health-port "$YZ_TEST_HEALTH_PORT" --version "$YZ_TEST_VERSION" \
    --binary "$YZ_TEST_ARTIFACTS/xboard-node-linux-amd64" \
    --xbctl-binary "$YZ_TEST_ARTIFACTS/xbctl-linux-amd64" --force-reconfigure --yes
check_root
detect_arch
detect_service_manager
perform_install
