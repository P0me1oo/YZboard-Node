#!/usr/bin/env bash
set -Eeuo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
KERNEL_TEST_PARENT=$(cd "${TMPDIR:-/tmp}" && pwd -P)
KERNEL_TEST_ROOT=$(mktemp -d "$KERNEL_TEST_PARENT/yzboard-kernel-defaults.XXXXXX")
cleanup() {
    case "$KERNEL_TEST_ROOT" in
        "$KERNEL_TEST_PARENT"/yzboard-kernel-defaults.*) rm -rf -- "$KERNEL_TEST_ROOT" ;;
        *) echo "拒绝清理范围外的测试目录" >&2; return 1 ;;
    esac
}
trap cleanup EXIT
INSTALLER_LIBRARY="$KERNEL_TEST_ROOT/install-library.sh"
sed '/^main "$@"$/d' "$REPO_ROOT/install.sh" >"$INSTALLER_LIBRARY"
mkdir -p "$KERNEL_TEST_ROOT/bin"

# 只记录生成配置时的参数，不下载程序，也不调用服务管理器。
cat >"$KERNEL_TEST_ROOT/bin/xbctl" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" >"$KERNEL_TEST_ARGS"
while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--credentials-out" ]]; then
        : >"$2"
        shift
    fi
    shift
done
printf 'INSTANCE_ID=kernel-default-test\n'
MOCK
chmod +x "$KERNEL_TEST_ROOT/bin/xbctl"

check_kernel() (
    local expected="$1" explicit="$2"
    shift 2
    # shellcheck source=/dev/null
    source "$INSTALLER_LIBRARY"
    trap - ERR EXIT INT TERM
    parse_args --panel https://kernel.example.invalid --node-id 1 "$@"
    [[ "$KERNEL_TYPE" == "$expected" ]]
    [[ "$KERNEL_EXPLICIT" == "$explicit" ]]
    TMP_DIR="$KERNEL_TEST_ROOT/render"
    BIN_STAGE_DIR="$KERNEL_TEST_ROOT/bin"
    CONFIG_FILE="$KERNEL_TEST_ROOT/existing.yml"
    CREDENTIALS_FILE="$KERNEL_TEST_ROOT/no-credentials.env"
    export KERNEL_TEST_ARGS="$KERNEL_TEST_ROOT/args.txt"
    mkdir -p "$TMP_DIR"
    render_config
    if [[ "$explicit" == 1 ]]; then
        [[ "$(awk '/^--kernel$/{getline; print}' "$KERNEL_TEST_ARGS")" == "$expected" ]]
    elif grep -Fx -- '--kernel' "$KERNEL_TEST_ARGS" >/dev/null; then
        echo "安装器不应把默认内核当成显式覆盖参数" >&2
        exit 1
    fi
)

check_kernel singbox 0
check_kernel xray 0 --node-type vless
check_kernel xray 0 --node-type VLESS
check_kernel xray 1 --kernel xray
check_kernel singbox 1 --node-type vless --kernel singbox
printf 'kernel:\n  type: xray\n' >"$KERNEL_TEST_ROOT/existing.yml"
check_kernel singbox 0
grep -Fx -- '--config' "$KERNEL_TEST_ROOT/args.txt" >/dev/null
echo "安装器默认内核测试通过"
