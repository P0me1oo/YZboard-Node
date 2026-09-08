#!/usr/bin/env bash
set -Eeuo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

APP_NAME="xboard-node"
INSTALL_ROOT="/etc/xboard-node"
BACKUP_DIR="${INSTALL_ROOT}/backups"
INSTALL_META="${INSTALL_ROOT}/install-meta.json"
CONFIG_FILE="${INSTALL_ROOT}/config.yml"
CREDENTIALS_FILE="${INSTALL_ROOT}/credentials.env"
DEFAULT_BIN_DIR="/usr/local/bin"
BIN_DIR=""
BIN_DIR_FILE="${INSTALL_ROOT}/bin-dir"
PREVIOUS_BIN_DIR=""
PREVIOUS_BINARY_PATH=""
PREVIOUS_CLI_PATH=""
BINARY_PATH="/usr/local/bin/xboard-node"
SERVICE_NAME="xboard-node"
SYSTEMD_SERVICE_NAME="${SERVICE_NAME}.service"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/${SYSTEMD_SERVICE_NAME}"
OPENRC_SERVICE_PATH="/etc/init.d/${SERVICE_NAME}"
OPENRC_LOG_PATH="/var/log/${SERVICE_NAME}.log"
SERVICE_MANAGER=""
SERVICE_PATH=""
CLI_PATH="/usr/local/bin/xbctl"
CLI_SYMLINK_PATH="/usr/bin/xbctl"
INSTALLER_COPY_PATH="${INSTALL_ROOT}/install.sh"
CLI_BINARY_SOURCE=""
DEFAULT_HEALTH_PORT=65530
DEFAULT_KERNEL="xray"
DEFAULT_MODE="node"
DEFAULT_ACTION="install"
DEFAULT_RELEASE_VERSION="latest"
DEFAULT_LOG_LEVEL="info"
DEFAULT_KERNEL_LOG_LEVEL="warn"
DEFAULT_DOWNLOAD_BASE="https://github.com/P0me1oo/YZboard-Node/releases"

ACTION="${DEFAULT_ACTION}"
MODE=""
PANEL_URL=""
TOKEN=""
NODE_ID=""
NODE_TYPE=""
MACHINE_ID=""
KERNEL_TYPE="${DEFAULT_KERNEL}"
KERNEL_EXPLICIT=0
RELEASE_VERSION="${DEFAULT_RELEASE_VERSION}"
HEALTH_PORT="${DEFAULT_HEALTH_PORT}"
HEALTH_ENABLED=1
RUNTIME_GOMEMLIMIT=""
RUNTIME_GOGC=""
BINARY_SOURCE=""
CLI_BINARY_SOURCE=""
FORCE_RECONFIGURE=0
PURGE=0
YES=0
ARCH=""
OS=""
DOWNLOAD_URL=""
CURRENT_STATE="fresh"
TMP_DIR=""
BIN_STAGE_DIR=""
BIN_DIR_RECORD_TMP=""
CLI_ENTRY_BACKUP_DIR=""
CLI_ENTRY_CHANGED=0
BINARIES_CHANGED=0
BINARY_REPLACE_STARTED=0
CLI_REPLACE_STARTED=0
ROLLBACK_FAILED=0
LOCK_HELD=0
INSTALL_COMMITTED=0
BACKUP_PATH=""
BACKUP_PENDING=""
SERVICE_EXISTED=0
CLEANUP_DONE=0

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${CYAN}[STEP]${NC} ${BOLD}$1${NC}"; }

cleanup_tmp() {
    if [ "$CLEANUP_DONE" -eq 1 ]; then
        return
    fi
    CLEANUP_DONE=1
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
    [ -z "$BIN_DIR_RECORD_TMP" ] || rm -f "$BIN_DIR_RECORD_TMP"
    [ -z "$BACKUP_PENDING" ] || rm -rf "$BACKUP_PENDING"
    if [ "$ROLLBACK_FAILED" -eq 0 ]; then
        [ -z "$BIN_STAGE_DIR" ] || rm -rf "$BIN_STAGE_DIR"
        [ -z "$CLI_ENTRY_BACKUP_DIR" ] || rm -rf "$CLI_ENTRY_BACKUP_DIR"
    else
        log_error "Recovery files retained: ${BACKUP_PATH} ${BIN_STAGE_DIR} ${CLI_ENTRY_BACKUP_DIR}"
    fi
    if [ "$LOCK_HELD" -eq 1 ]; then
        rmdir "$INSTALL_ROOT/.install-lock" || true
    fi
}

load_health_port_from_config() {
    local cfg_path="$1"
    if [ ! -f "$cfg_path" ]; then
        return
    fi
    local parsed
    local reader="${PREVIOUS_CLI_PATH:-$CLI_PATH}"
    if [ -x "$reader" ]; then
        parsed=$("$reader" config health-port --config "$cfg_path" 2>/dev/null)
    else
        parsed=$(grep -m1 'health_port:' "$cfg_path" 2>/dev/null | sed 's/.*health_port:[[:space:]]*//' | tr -cd '0-9')
    fi
    if [ -n "$parsed" ] && [ "$parsed" -ge 0 ] 2>/dev/null; then
        HEALTH_PORT="$parsed"
        if [ "$HEALTH_PORT" -eq 0 ]; then
            HEALTH_ENABLED=0
        else
            HEALTH_ENABLED=1
        fi
    fi
}

same_directory_entry() {
    local first second
    first=$(stat -c '%d:%i' "$1") || return 1
    second=$(stat -c '%d:%i' "$2") || return 1
    [ "$first" = "$second" ]
}

rollback_install() {
    log_warn "Rolling back installation"
    service_stop >/dev/null 2>&1 || true
    local failed=0 name target previous
    if [ "$BINARIES_CHANGED" -eq 1 ]; then
        for name in xboard-node xbctl; do
            if [ "$name" = xboard-node ] && [ "$BINARY_REPLACE_STARTED" -eq 0 ]; then continue; fi
            if [ "$name" = xbctl ] && [ "$CLI_REPLACE_STARTED" -eq 0 ]; then continue; fi
            target="$BIN_DIR/$name"
            previous="$BIN_STAGE_DIR/previous-$name"
            if [ "$BIN_DIR" = "$PREVIOUS_BIN_DIR" ] && { [ -e "$previous" ] || [ -L "$previous" ]; }; then
                if ! same_directory_entry "$previous" "$target" 2>/dev/null; then
                    mv -f "$previous" "$target" || failed=1
                fi
            else
                rm -f "$target" || failed=1
            fi
        done
    fi
    if [ "$CLI_ENTRY_CHANGED" -eq 1 ]; then
        if [ -n "$CLI_ENTRY_BACKUP_DIR" ] && { [ -e "$CLI_ENTRY_BACKUP_DIR/xbctl" ] || [ -L "$CLI_ENTRY_BACKUP_DIR/xbctl" ]; }; then
            if ! same_directory_entry "$CLI_ENTRY_BACKUP_DIR/xbctl" "$CLI_SYMLINK_PATH" 2>/dev/null; then
                mv -f "$CLI_ENTRY_BACKUP_DIR/xbctl" "$CLI_SYMLINK_PATH" || failed=1
            fi
        else
            rm -f "$CLI_SYMLINK_PATH" || failed=1
        fi
    fi
    if [ -n "$BACKUP_PATH" ] && [ -d "$BACKUP_PATH" ]; then
        local item mode
        for item in config.yml credentials.env install-meta.json bin-dir service; do
            case "$item" in
                config.yml) target="$CONFIG_FILE"; mode=600 ;;
                credentials.env) target="$CREDENTIALS_FILE"; mode=600 ;;
                install-meta.json) target="$INSTALL_META"; mode=644 ;;
                bin-dir) target="$BIN_DIR_FILE"; mode=644 ;;
                service) target="$SERVICE_PATH"; mode=$(service_file_mode) ;;
            esac
            if [ -f "$BACKUP_PATH/$item" ]; then
                install -m "$mode" "$BACKUP_PATH/$item" "$target" || failed=1
            else
                rm -f "$target" || failed=1
            fi
        done
    fi
    if [ "$failed" -ne 0 ]; then
        ROLLBACK_FAILED=1
        log_error "Rollback incomplete; recovery files will be retained"
        return 1
    fi
    CLI_PATH="$PREVIOUS_CLI_PATH"
    load_health_port_from_config "$CONFIG_FILE"
    service_reload || true
    if [ "$SERVICE_EXISTED" -eq 1 ] || [ -f "$SERVICE_PATH" ]; then
        service_reset_failed >/dev/null 2>&1 || true
        service_restart >/dev/null 2>&1 || true
        if ! wait_for_health; then
            log_error "Original files restored but the service did not become healthy"
            show_recent_logs
            return 1
        fi
    else
        service_disable >/dev/null 2>&1 || true
    fi
    log_warn "Rollback complete"
}

on_error() {
    local exit_code=$?
    local line_no=${1:-unknown}
    trap '' INT TERM
    if [ "$exit_code" -ne 0 ]; then
        log_error "Install failed at line ${line_no} (exit=${exit_code})"
        if [ -n "$BACKUP_PATH" ] && [ "$INSTALL_COMMITTED" -eq 0 ]; then
            rollback_install || true
        fi
    fi
    cleanup_tmp
    exit "$exit_code"
}

on_signal() {
    local exit_code="$1"
    trap '' INT TERM
    if [ -n "$BACKUP_PATH" ] && [ "$INSTALL_COMMITTED" -eq 0 ]; then
        rollback_install || true
    fi
    cleanup_tmp
    exit "$exit_code"
}
trap 'on_error $LINENO' ERR
trap cleanup_tmp EXIT
trap 'on_signal 130' INT
trap 'on_signal 143' TERM

usage() {
    cat <<'HELP'

  xboard-node Installer

  ACTIONS:
    install      Install or reconcile the configured deployment (default)
    upgrade      Upgrade binary and restart service
    uninstall    Remove installed service and binary (config kept unless --purge)
    status       Show current installation status
    help         Show this help

  MODES (auto-detected from --node-id or --machine-id if omitted):
    --mode node      Panel single-node mode (default)
    --mode machine   Panel machine mode

  REQUIRED FOR NODE MODE:
    --panel, -a      Panel URL
    --token, -t      Panel server token
    --node-id, -n    Node ID

  REQUIRED FOR MACHINE MODE:
    --panel, -a       Panel URL
    --token, -t       Machine token
    --machine-id      Machine ID

  OPTIONAL:
    --node-type, -T     Explicit node type for node mode
    --kernel, -k        xray or singbox (default: xray)
    --version           Release version or latest (default: latest)
    --binary            Use a local xboard-node binary path instead of downloading
    --xbctl-binary      Use a local xbctl binary path instead of downloading
    --bin-dir           程序目录，默认 /usr/local/bin；升级时可迁移已有安装
    --health-port       Local health port (default: 65530, use 0 to disable)
    --gomemlimit        Runtime GOMEMLIMIT value, e.g. 256MiB
    --gogc              Runtime GOGC value, e.g. 50
    --force-reconfigure Overwrite an existing install even if mode/target changed
    --purge             With uninstall, delete /etc/xboard-node too
    --yes, -y           Non-interactive confirmation for destructive operations

  EXAMPLES:
    sudo bash install.sh --panel https://panel.example.com --token TOKEN --node-id 1
    sudo bash install.sh --panel https://panel.example.com --token TOKEN --machine-id 1
    sudo bash install.sh upgrade
    sudo bash install.sh upgrade --bin-dir /boot/xboard-node
    sudo bash install.sh uninstall --purge --yes

HELP
}

parse_args() {
    local positional=()
    while [ $# -gt 0 ]; do
        case "$1" in
            install|upgrade|uninstall|status|help)
                ACTION="$1"
                shift
                ;;
            --mode)
                MODE="$2"
                shift 2
                ;;
            --panel|-a|--api)
                PANEL_URL="$2"
                shift 2
                ;;
            --token|-t)
                TOKEN="$2"
                shift 2
                ;;
            --node-id|-n)
                NODE_ID="$2"
                shift 2
                ;;
            --node-type|-T)
                NODE_TYPE="$2"
                shift 2
                ;;
            --machine-id)
                MACHINE_ID="$2"
                shift 2
                ;;
            --kernel|-k)
                KERNEL_TYPE="$2"
                KERNEL_EXPLICIT=1
                shift 2
                ;;
            --version)
                RELEASE_VERSION="$2"
                shift 2
                ;;
            --binary)
                BINARY_SOURCE="$2"
                shift 2
                ;;
            --xbctl-binary)
                CLI_BINARY_SOURCE="$2"
                shift 2
                ;;
            --bin-dir)
                if [ $# -lt 2 ] || [ -z "$2" ]; then
                    log_error "--bin-dir requires an absolute directory" >&2
                    return 1
                fi
                BIN_DIR="$2"
                shift 2
                ;;
            --health-port)
                HEALTH_PORT="$2"
                shift 2
                ;;
            --gomemlimit)
                RUNTIME_GOMEMLIMIT="$2"
                shift 2
                ;;
            --gogc)
                RUNTIME_GOGC="$2"
                shift 2
                ;;
            --force-reconfigure)
                FORCE_RECONFIGURE=1
                shift
                ;;
            --purge)
                PURGE=1
                shift
                ;;
            --yes|-y)
                YES=1
                shift
                ;;
            --help|-h)
                ACTION="help"
                shift
                ;;
            *)
                positional+=("$1")
                shift
                ;;
        esac
    done

    if [ ${#positional[@]} -gt 0 ] && [ "$ACTION" = "install" ]; then
        ACTION="${positional[0]}"
    fi

    case "$KERNEL_TYPE" in
        singbox|SingBox|SINGBOX) KERNEL_TYPE="singbox" ;;
        xray|Xray|XRAY) KERNEL_TYPE="xray" ;;
        *) ;;
    esac

    # 默认内核为 xray，但 xray 的入站协议少于 sing-box。未显式指定内核时，
    # 若节点协议 xray 不支持，则回退到 sing-box，避免装完直接起不来。
    if [ "$KERNEL_EXPLICIT" -eq 0 ] && [ "$KERNEL_TYPE" = "xray" ] && [ -n "$NODE_TYPE" ]; then
        case "$(printf '%s' "$NODE_TYPE" | tr '[:upper:]' '[:lower:]')" in
            tuic|naive|anytls|mieru|socks|http)
                log_warn "Node type '${NODE_TYPE}' is not supported by the xray kernel; falling back to singbox."
                log_warn "Pass --kernel xray explicitly to override this fallback."
                KERNEL_TYPE="singbox"
                ;;
        esac
    fi

    # Auto-detect mode from arguments when --mode is not specified.
    if [ -z "$MODE" ]; then
        if [ -n "$MACHINE_ID" ]; then
            MODE="machine"
        else
            MODE="node"
        fi
    fi

    case "$MODE" in
        node|machine) ;;
        *)
            log_error "Unsupported mode: $MODE"
            usage
            exit 1
            ;;
    esac
}

check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log_error "Please run as root or with sudo"
        exit 1
    fi
}

detect_arch() {
    local raw
    raw=$(uname -m)
    case "$raw" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            log_error "Unsupported architecture: $raw"
            exit 1
            ;;
    esac
}

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS="$ID"
    else
        OS="unknown"
    fi
}

detect_service_manager() {
    if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
        SERVICE_MANAGER="systemd"
        SERVICE_PATH="$SYSTEMD_SERVICE_PATH"
        return
    fi
    if command -v rc-service >/dev/null 2>&1 && \
       command -v rc-update >/dev/null 2>&1 && \
       command -v supervise-daemon >/dev/null 2>&1 && \
       [ -x /sbin/openrc-run ]; then
        SERVICE_MANAGER="openrc"
        SERVICE_PATH="$OPENRC_SERVICE_PATH"
        return
    fi
    log_error "A running systemd or OpenRC service manager is required"
    exit 1
}

normalize_bin_dir() {
    local dir="$1"
    while [[ "$dir" == */ ]] && [ "$dir" != / ]; do dir="${dir%/}"; done
    if ! [[ "$dir" =~ ^/[A-Za-z0-9._/-]+$ ]] || [ "$dir" = / ] ||
       [[ "$dir/" == *'//'* || "$dir/" == *'/./'* || "$dir/" == *'/../'* ]]; then
        log_error "Binary directory must be an absolute path using letters, numbers, /, ., _ and -, without . or .. components" >&2
        return 1
    fi
    printf '%s\n' "$dir"
}

load_install_paths() {
    local saved="$DEFAULT_BIN_DIR"
    if [ -f "$BIN_DIR_FILE" ]; then
        local saved_lines=()
        mapfile -t saved_lines <"$BIN_DIR_FILE"
        if [ "${#saved_lines[@]}" -ne 1 ]; then
            log_error "Invalid binary directory record: $BIN_DIR_FILE" >&2
            return 1
        fi
        saved="${saved_lines[0]}"
    fi
    PREVIOUS_BIN_DIR=$(normalize_bin_dir "$saved")
    if [ -d "$PREVIOUS_BIN_DIR" ]; then
        PREVIOUS_BIN_DIR=$(cd "$PREVIOUS_BIN_DIR" && pwd -P)
        PREVIOUS_BIN_DIR=$(normalize_bin_dir "$PREVIOUS_BIN_DIR")
    fi
    BIN_DIR=$(normalize_bin_dir "${BIN_DIR:-$PREVIOUS_BIN_DIR}")
    if [ -d "$BIN_DIR" ]; then
        BIN_DIR=$(cd "$BIN_DIR" && pwd -P)
    fi
    PREVIOUS_BINARY_PATH="$PREVIOUS_BIN_DIR/xboard-node"
    PREVIOUS_CLI_PATH="$PREVIOUS_BIN_DIR/xbctl"
    BINARY_PATH="$BIN_DIR/xboard-node"
    CLI_PATH="$BIN_DIR/xbctl"
    if [ "$ACTION" != install ] && [ "$ACTION" != upgrade ] && [ "$BIN_DIR" != "$PREVIOUS_BIN_DIR" ]; then
        log_error "Use install or upgrade to change --bin-dir" >&2
        return 1
    fi
}

lock_installation() {
    mkdir -p "$INSTALL_ROOT"
    if ! mkdir "$INSTALL_ROOT/.install-lock"; then
        log_error "Another installer may be running; check $INSTALL_ROOT/.install-lock before retrying" >&2
        return 1
    fi
    LOCK_HELD=1
}

service_file_mode() {
    if [ "$SERVICE_MANAGER" = "openrc" ]; then
        echo 755
    else
        echo 644
    fi
}

service_reload() {
    if [ "$SERVICE_MANAGER" = "systemd" ]; then
        systemctl daemon-reload
    fi
}

service_reset_failed() {
    if [ "$SERVICE_MANAGER" = "systemd" ]; then
        systemctl reset-failed "$SYSTEMD_SERVICE_NAME"
    fi
}

service_is_active() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl is-active "$SYSTEMD_SERVICE_NAME" >/dev/null 2>&1 ;;
        openrc) rc-service --quiet "$SERVICE_NAME" status >/dev/null 2>&1 ;;
        *) return 1 ;;
    esac
}

service_start() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl start "$SYSTEMD_SERVICE_NAME" ;;
        openrc) rc-service "$SERVICE_NAME" start ;;
    esac
}

service_stop() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl stop "$SYSTEMD_SERVICE_NAME" ;;
        openrc) rc-service "$SERVICE_NAME" stop ;;
    esac
}

service_restart() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl restart "$SYSTEMD_SERVICE_NAME" ;;
        openrc) rc-service "$SERVICE_NAME" restart ;;
    esac
}

service_enable() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl enable "$SYSTEMD_SERVICE_NAME" ;;
        openrc) rc-update add "$SERVICE_NAME" default ;;
    esac
}

service_disable() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl disable "$SYSTEMD_SERVICE_NAME" ;;
        openrc) rc-update del "$SERVICE_NAME" default ;;
    esac
}

service_status() {
    case "$SERVICE_MANAGER" in
        systemd) systemctl status "$SYSTEMD_SERVICE_NAME" --no-pager ;;
        openrc) rc-service "$SERVICE_NAME" status ;;
    esac
}

run_with_retry() {
    local attempts="$1"
    local delay="$2"
    shift 2
    local i=1
    while [ "$i" -le "$attempts" ]; do
        if "$@"; then
            return 0
        fi
        if [ "$i" -lt "$attempts" ]; then
            log_warn "Command failed, retrying in ${delay}s: $*"
            sleep "$delay"
        fi
        i=$((i + 1))
    done
    return 1
}

install_dependencies() {
    case "$OS" in
        ubuntu|debian)
            DEBIAN_FRONTEND=noninteractive run_with_retry 10 3 apt-get update -qq
            DEBIAN_FRONTEND=noninteractive run_with_retry 10 3 apt-get install -y -qq curl wget ca-certificates coreutils >/dev/null 2>&1
            ;;
        centos|rhel|rocky|almalinux|fedora)
            if command -v dnf >/dev/null 2>&1; then
                run_with_retry 5 3 dnf install -y -q curl wget ca-certificates coreutils >/dev/null 2>&1
            else
                run_with_retry 5 3 yum install -y -q curl wget ca-certificates coreutils >/dev/null 2>&1
            fi
            ;;
        alpine)
            run_with_retry 5 3 apk add --no-cache curl wget ca-certificates coreutils openrc >/dev/null 2>&1
            ;;
        *)
            log_warn "OS ${OS} is not in the official support set; continuing best-effort"
            ;;
    esac
}

ensure_dirs() {
    mkdir -p "$INSTALL_ROOT" "$BACKUP_DIR" "$BIN_DIR"
    chmod 700 "$INSTALL_ROOT"
    BIN_DIR=$(cd "$BIN_DIR" && pwd -P)
    BIN_DIR=$(normalize_bin_dir "$BIN_DIR")
    BINARY_PATH="$BIN_DIR/xboard-node"
    CLI_PATH="$BIN_DIR/xbctl"
    local target
    for target in "$BINARY_PATH" "$CLI_PATH"; do
        if [ -e "$target" ] || [ -L "$target" ]; then
            if [ "$BIN_DIR" != "$PREVIOUS_BIN_DIR" ]; then
                # 迁移到 /usr/bin 时，只允许替换已经指向原 xbctl 的入口。
                if [ "$target" = "$CLI_SYMLINK_PATH" ] && [ -L "$target" ] &&
                   [ "$(readlink "$target")" = "$PREVIOUS_CLI_PATH" ]; then
                    continue
                fi
                log_error "Destination already exists: $target"
                return 1
            fi
            if [ ! -f "$target" ] && [ ! -L "$target" ]; then
                log_error "Binary path is not a regular file or symlink: $target"
                return 1
            fi
        fi
    done
    # 大文件直接暂存在目标分区；配置和凭据仍暂存在系统临时目录。
    BIN_STAGE_DIR=$(mktemp -d "$BIN_DIR/.xboard-node-install.XXXXXX")
}

validate_positive_int() {
    local label="$1"
    local value="$2"
    if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -le 0 ]; then
        log_error "${label} must be a positive integer, got: ${value}"
        exit 1
    fi
}

validate_install_request() {
    if [ -z "$PANEL_URL" ]; then
        log_error "Panel URL is required"
        exit 1
    fi
    if [ -z "$TOKEN" ]; then
        log_error "Token is required"
        exit 1
    fi
    if ! [[ "$HEALTH_PORT" =~ ^[0-9]+$ ]]; then
        log_error "health-port must be a non-negative integer"
        exit 1
    fi
    if [ "$HEALTH_PORT" -eq 0 ]; then
        HEALTH_ENABLED=0
    fi
    case "$KERNEL_TYPE" in
        singbox|xray) ;;
        *)
            log_error "Kernel must be singbox or xray"
            exit 1
            ;;
    esac
    case "$MODE" in
        node)
            validate_positive_int "Node ID" "$NODE_ID"
            ;;
        machine)
            validate_positive_int "Machine ID" "$MACHINE_ID"
            ;;
    esac
}

detect_current_state() {
    local has_binary=0 has_config=0 has_service=0
    [ -x "${PREVIOUS_BINARY_PATH:-$BINARY_PATH}" ] && has_binary=1
    [ -f "$CONFIG_FILE" ] && has_config=1
    [ -f "$SERVICE_PATH" ] && has_service=1

    if [ "$has_binary" -eq 0 ] && [ "$has_config" -eq 0 ] && [ "$has_service" -eq 0 ]; then
        CURRENT_STATE="fresh"
    elif [ "$has_binary" -eq 1 ] && [ "$has_config" -eq 1 ] && [ "$has_service" -eq 1 ]; then
        CURRENT_STATE="installed"
    else
        CURRENT_STATE="partial"
    fi
}

require_reconfigure_confirmation() {
    return
}

select_binary_source() {
    if [ -n "$BINARY_SOURCE" ]; then
        if [ ! -f "$BINARY_SOURCE" ]; then
            log_error "Binary source not found: $BINARY_SOURCE"
            exit 1
        fi
        echo "$BINARY_SOURCE"
        return
    fi
    if [ -f "./xboard-node" ]; then
        echo "./xboard-node"
        return
    fi
    if [ -f "./xboard-node-linux-${ARCH}" ]; then
        echo "./xboard-node-linux-${ARCH}"
        return
    fi
    echo ""
}

resolve_download_url() {
    local artifact="$1"
    if [ "$RELEASE_VERSION" = "latest" ]; then
        DOWNLOAD_URL="${DEFAULT_DOWNLOAD_BASE}/latest/download/${artifact}"
    else
        DOWNLOAD_URL="${DEFAULT_DOWNLOAD_BASE}/download/${RELEASE_VERSION}/${artifact}"
    fi
}

verify_release_checksum() {
    local file="$1"
    local artifact="$2"
    local checksums="$TMP_DIR/SHA256SUMS"
    if [ ! -f "$checksums" ]; then
        resolve_download_url "SHA256SUMS"
        log_step "Downloading release checksums: ${DOWNLOAD_URL}"
        if ! curl -fsSL "$DOWNLOAD_URL" -o "$checksums"; then
            log_error "Failed to download release checksums from ${DOWNLOAD_URL}"
            exit 1
        fi
    fi

    local expected
    local actual
    expected=$(awk -v name="$artifact" '$2 == name || $2 == "*" name { print $1; exit }' "$checksums")
    if [ -z "$expected" ]; then
        log_error "No checksum found for ${artifact}"
        exit 1
    fi
    actual=$(sha256sum "$file" | awk '{ print $1 }')
    if [ "$actual" != "$expected" ]; then
        log_error "Checksum mismatch for ${artifact}"
        exit 1
    fi
    log_info "Checksum verified: ${artifact}"
}

stage_binary() {
    local staged="$BIN_STAGE_DIR/xboard-node"
    local local_src
    local_src=$(select_binary_source)
    if [ -n "$local_src" ]; then
        log_step "Using local binary: ${local_src}"
        cp "$local_src" "$staged"
    else
        local artifact="xboard-node-linux-${ARCH}"
        resolve_download_url "$artifact"
        log_step "Downloading binary: ${DOWNLOAD_URL}"
        if ! curl -fsSL "$DOWNLOAD_URL" -o "$staged"; then
            log_error "Failed to download binary from ${DOWNLOAD_URL}"
            exit 1
        fi
        verify_release_checksum "$staged" "$artifact"
    fi
    chmod 755 "$staged"
    if ! "$staged" -v >/dev/null 2>&1; then
        log_error "Downloaded binary failed version check"
        exit 1
    fi
}

stage_xbctl() {
    local staged="$BIN_STAGE_DIR/xbctl"
    local local_src=""
    if [ -n "$CLI_BINARY_SOURCE" ]; then
        if [ ! -f "$CLI_BINARY_SOURCE" ]; then
            log_error "xbctl binary source not found: $CLI_BINARY_SOURCE"
            exit 1
        fi
        local_src="$CLI_BINARY_SOURCE"
    elif [ -f "./xbctl" ]; then
        local_src="./xbctl"
    elif [ -f "./xbctl-linux-${ARCH}" ]; then
        local_src="./xbctl-linux-${ARCH}"
    fi
    if [ -n "$local_src" ]; then
        log_step "Using local xbctl binary: ${local_src}"
        cp "$local_src" "$staged"
    else
        local artifact="xbctl-linux-${ARCH}"
        resolve_download_url "$artifact"
        log_step "Downloading xbctl: ${DOWNLOAD_URL}"
        if ! curl -fsSL "$DOWNLOAD_URL" -o "$staged"; then
            log_error "Failed to download xbctl from ${DOWNLOAD_URL}"
            exit 1
        fi
        verify_release_checksum "$staged" "$artifact"
    fi
    chmod 755 "$staged"
    if ! "$staged" version > /dev/null 2>&1; then
        log_error "Downloaded xbctl failed version check"
        exit 1
    fi
    if [ "$BIN_DIR" != "$DEFAULT_BIN_DIR" ] && ! "$staged" config bin-dir >/dev/null 2>&1; then
        log_error "This release does not support --bin-dir; migrate to $DEFAULT_BIN_DIR before downgrading"
        exit 1
    fi
}

render_config() {
    local init_args=(
        config init
        --mode "$MODE"
        --panel-url "$PANEL_URL"
        --kernel "${KERNEL_TYPE:-xray}"
        --health-port "${HEALTH_PORT:-0}"
        --token "$TOKEN"
        --version "$RELEASE_VERSION"
        --output "$TMP_DIR/config.yml"
        --credentials-out "$TMP_DIR/credentials.env"
        --meta "$TMP_DIR/install-meta.json"
        --install-root "$INSTALL_ROOT"
    )
    if [ -f "$CONFIG_FILE" ]; then
        init_args+=(--config "$CONFIG_FILE")
    fi
    if [ -f "$CREDENTIALS_FILE" ]; then
        init_args+=(--credentials-in "$CREDENTIALS_FILE")
    fi
    if [ "$MODE" = "machine" ]; then
        init_args+=(--machine-id "$MACHINE_ID")
    else
        init_args+=(--node-id "$NODE_ID")
        if [ -n "$NODE_TYPE" ]; then
            init_args+=(--node-type "$NODE_TYPE")
        fi
    fi
    if [ -n "$RUNTIME_GOMEMLIMIT" ]; then
        init_args+=(--gomemlimit "$RUNTIME_GOMEMLIMIT")
    fi
    if [ -n "$RUNTIME_GOGC" ] && [ "$RUNTIME_GOGC" -gt 0 ] 2>/dev/null; then
        init_args+=(--gogc "$RUNTIME_GOGC")
    fi

    local output
    output=$("$BIN_STAGE_DIR/xbctl" "${init_args[@]}") || {
        log_error "xbctl config init failed"
        exit 1
    }

    INSTANCE_ID=$(echo "$output" | grep '^INSTANCE_ID=' | cut -d= -f2-)
    chmod 600 "$TMP_DIR/credentials.env"
}

render_service() {
    if [ "$SERVICE_MANAGER" = "openrc" ]; then
        cat >"$TMP_DIR/service" <<EOF_OPENRC
#!/sbin/openrc-run

name="Xboard Node Backend"
description="YZboard node backend"
supervisor=supervise-daemon
command="${BINARY_PATH}"
command_args="-c ${CONFIG_FILE}"
directory="${INSTALL_ROOT}"
pidfile="/run/${SERVICE_NAME}.pid"
output_log="${OPENRC_LOG_PATH}"
error_log="${OPENRC_LOG_PATH}"
respawn_delay=5
respawn_max=0
retry="TERM/150/KILL/5"
no_new_privs=true
rc_ulimit="-n 1048576"

depend() {
    need net localmount
}

start_pre() {
    local line key value
    if [ -r "${CREDENTIALS_FILE}" ]; then
        while IFS= read -r line || [ -n "\$line" ]; do
            case "\$line" in
                ""|\#*) continue ;;
            esac
            key=\${line%%=*}
            if [ "\$key" = "\$line" ]; then
                eerror "Invalid credentials entry: missing '='"
                return 1
            fi
            case "\$key" in
                ""|[0-9]*|*[!A-Za-z0-9_]*)
                    eerror "Invalid credentials key: \$key"
                    return 1
                    ;;
            esac
            value=\${line#*=}
            export "\${key}=\${value}"
        done < "${CREDENTIALS_FILE}"
    fi
    checkpath -f -m 0640 -o root:root "${OPENRC_LOG_PATH}"
}
EOF_OPENRC
        return
    fi

    cat >"$TMP_DIR/service" <<EOF_UNIT
[Unit]
Description=Xboard Node Backend
Documentation=https://github.com/P0me1oo/YZboard-Node
After=network-online.target
Wants=network-online.target
RequiresMountsFor=${BIN_DIR:-/usr/local/bin}

[Service]
Type=simple
WorkingDirectory=${INSTALL_ROOT}
EnvironmentFile=-${CREDENTIALS_FILE}
ExecStart=${BINARY_PATH} -c ${CONFIG_FILE}
Restart=always
RestartSec=5
TimeoutStopSec=150s
LimitNOFILE=1048576
NoNewPrivileges=true
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF_UNIT
}

backup_existing_state() {
    local snapshot item source name
    BACKUP_PENDING=$(mktemp -d "$BACKUP_DIR/install-XXXXXX")
    snapshot="$BACKUP_PENDING"
    for item in config.yml credentials.env install-meta.json bin-dir service; do
        case "$item" in
            config.yml) source="$CONFIG_FILE" ;;
            credentials.env) source="$CREDENTIALS_FILE" ;;
            install-meta.json) source="$INSTALL_META" ;;
            bin-dir) source="$BIN_DIR_FILE" ;;
            service) source="$SERVICE_PATH" ;;
        esac
        if [ -f "$source" ]; then
            cp "$source" "$snapshot/$item"
        fi
    done
    if [ -f "$SERVICE_PATH" ]; then SERVICE_EXISTED=1; else SERVICE_EXISTED=0; fi
    if [ "$BIN_DIR" = "$PREVIOUS_BIN_DIR" ]; then
        for name in xboard-node xbctl; do
            source="$BIN_DIR/$name"
            if [ -e "$source" ] || [ -L "$source" ]; then
                # 硬链接保留原文件，不复制二进制内容；替换时必须使用同分区重命名。
                ln "$source" "$BIN_STAGE_DIR/previous-$name"
            fi
        done
    fi
    if [ "$BIN_DIR" != "$PREVIOUS_BIN_DIR" ] || [ "$CLI_PATH" != "$CLI_SYMLINK_PATH" ]; then
        if [ -e "$CLI_SYMLINK_PATH" ] || [ -L "$CLI_SYMLINK_PATH" ]; then
            CLI_ENTRY_BACKUP_DIR=$(mktemp -d "$(dirname "$CLI_SYMLINK_PATH")/.xbctl-entry.XXXXXX")
            ln "$CLI_SYMLINK_PATH" "$CLI_ENTRY_BACKUP_DIR/xbctl"
        fi
    fi
    # 所有快照完成后才允许错误处理执行恢复，避免备份失败时删除原文件。
    BACKUP_PATH="$snapshot"
    BACKUP_PENDING=""
}

replace_binary_files() {
    BINARIES_CHANGED=1
    # 在重命名前登记恢复范围，覆盖命令完成后立即收到中断信号的情况。
    BINARY_REPLACE_STARTED=1
    mv -f "$BIN_STAGE_DIR/xboard-node" "$BINARY_PATH"
    if [ "$CLI_PATH" = "$CLI_SYMLINK_PATH" ] && [ "$BIN_DIR" != "$PREVIOUS_BIN_DIR" ]; then
        CLI_ENTRY_CHANGED=1
    fi
    CLI_REPLACE_STARTED=1
    mv -f "$BIN_STAGE_DIR/xbctl" "$CLI_PATH"
    if [ "$CLI_PATH" != "$CLI_SYMLINK_PATH" ]; then
        CLI_ENTRY_CHANGED=1
        ln -sfn "$CLI_PATH" "$CLI_SYMLINK_PATH"
    fi
    printf '%s\n' "$BIN_DIR" >"$TMP_DIR/bin-dir"
    BIN_DIR_RECORD_TMP=$(mktemp "$BIN_DIR_FILE.XXXXXX")
    install -m 644 "$TMP_DIR/bin-dir" "$BIN_DIR_RECORD_TMP"
    mv -f "$BIN_DIR_RECORD_TMP" "$BIN_DIR_FILE"
}

finish_install() {
    # 新服务验证成功后才删除原目录中的两个程序，其他文件全部保留。
    INSTALL_COMMITTED=1
    if [ "$BIN_DIR" != "$PREVIOUS_BIN_DIR" ]; then
        if ! rm -f "$PREVIOUS_BINARY_PATH"; then
            log_warn "New installation is active; could not remove $PREVIOUS_BINARY_PATH"
        fi
        if [ "$PREVIOUS_CLI_PATH" != "$CLI_SYMLINK_PATH" ]; then
            if ! rm -f "$PREVIOUS_CLI_PATH"; then
                log_warn "New installation is active; could not remove $PREVIOUS_CLI_PATH"
            fi
        fi
    fi
    log_info "Binary directory: $BIN_DIR"
}

stop_existing_service() {
    if [ -f "$SERVICE_PATH" ] || service_is_active; then
        service_stop >/dev/null 2>&1 || true
    fi
}

install_staged_files() {
    stop_existing_service
    replace_binary_files
    install -m 600 "$TMP_DIR/config.yml" "$CONFIG_FILE"
    install -m 600 "$TMP_DIR/credentials.env" "$CREDENTIALS_FILE"
    install -m 644 "$TMP_DIR/install-meta.json" "$INSTALL_META"
    if [ -f "$0" ] && [ "$(realpath "$0")" != "$(realpath "$INSTALLER_COPY_PATH" 2>/dev/null || echo "$INSTALLER_COPY_PATH")" ]; then
        install -m 755 "$0" "$INSTALLER_COPY_PATH"
    fi
    install -m "$(service_file_mode)" "$TMP_DIR/service" "$SERVICE_PATH"
    service_reload
    service_enable > /dev/null 2>&1
}

wait_for_health() {
    if ! service_is_active; then
        return 1
    fi
    if [ "$HEALTH_ENABLED" -eq 0 ]; then
        return 0
    fi
    local attempt=0
    local max_attempts=30
    while [ "$attempt" -lt "$max_attempts" ]; do
        if ! service_is_active; then
            return 1
        fi
        if curl -fsS "http://127.0.0.1:${HEALTH_PORT}/healthz" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    return 1
}

show_recent_logs() {
    if [ "$SERVICE_MANAGER" = "systemd" ] && command -v journalctl >/dev/null 2>&1; then
        journalctl -u "$SYSTEMD_SERVICE_NAME" -n 30 --no-pager || true
    elif [ "$SERVICE_MANAGER" = "openrc" ] && [ -f "$OPENRC_LOG_PATH" ]; then
        tail -n 30 "$OPENRC_LOG_PATH" || true
    fi
}

start_service() {
    if service_is_active; then
        service_restart
    else
        service_start
    fi
    if ! wait_for_health; then
        log_error "Service failed health check"
        show_recent_logs
        return 1
    fi
}

perform_install() {
    validate_install_request
    detect_current_state
    require_reconfigure_confirmation
    TMP_DIR=$(mktemp -d)
    ensure_dirs
    stage_binary
    stage_xbctl
    render_config
    render_service
    backup_existing_state
    install_staged_files
    start_service
    finish_install

    log_info "Installation succeeded"
    log_info "Service: ${SERVICE_NAME} (${SERVICE_MANAGER})"
    log_info "Config: ${CONFIG_FILE}"
    log_info "Credentials: ${CREDENTIALS_FILE}"
    if [ "$HEALTH_ENABLED" -eq 1 ]; then
        log_info "Health: http://127.0.0.1:${HEALTH_PORT}/healthz"
    fi
    log_info "CLI: ${CLI_PATH}  (run '${CLI_PATH} list' if xbctl is not in PATH)"
}

perform_upgrade() {
    detect_current_state
    if [ "$CURRENT_STATE" = "fresh" ]; then
        log_warn "No existing install found; falling back to install"
        perform_install
        return
    fi
    load_health_port_from_config "$CONFIG_FILE"
    TMP_DIR=$(mktemp -d)
    ensure_dirs
    stage_binary
    stage_xbctl
    render_service
    backup_existing_state
    replace_binary_files
    install -m "$(service_file_mode)" "$TMP_DIR/service" "$SERVICE_PATH"
    service_reload
    service_restart
    if ! wait_for_health; then
        log_error "Upgrade health check failed"
        show_recent_logs
        return 1
    fi
    "$CLI_PATH" config refresh-meta \
        --config "$CONFIG_FILE" \
        --meta "$INSTALL_META" \
        --version "$RELEASE_VERSION"
    finish_install
    log_info "Upgrade succeeded"
}

confirm_uninstall() {
    if [ "$YES" -eq 1 ]; then
        return
    fi
    echo
    read -r -p "Proceed with uninstall? [y/N]: " answer
    if ! [[ "$answer" =~ ^[Yy]$ ]]; then
        log_warn "Uninstall cancelled"
        exit 0
    fi
}

perform_uninstall() {
    confirm_uninstall
    if [ -f "$SERVICE_PATH" ]; then
        service_stop >/dev/null 2>&1 || true
        service_disable >/dev/null 2>&1 || true
        rm -f "$SERVICE_PATH"
        service_reload || true
    fi
    rm -f "$BINARY_PATH"
    rm -f "$CLI_PATH"
    rm -f "$CLI_SYMLINK_PATH" 2>/dev/null || true
    rm -f "$BIN_DIR_FILE"
    if [ "$PURGE" -eq 1 ]; then
        rm -rf "$INSTALL_ROOT"
        log_info "Removed ${INSTALL_ROOT}"
    else
        rm -f "$INSTALL_META"
        log_info "Config preserved under ${INSTALL_ROOT}"
    fi
    log_info "Uninstall complete"
}

perform_status() {
    detect_current_state
    echo
    echo -e "${BOLD}xboard-node install status${NC}"
    echo "  state:   ${CURRENT_STATE}"
    echo "  bin-dir: ${BIN_DIR}"
    if [ -f "$INSTALL_META" ]; then
        echo "  meta:    ${INSTALL_META}"
        if [ -x "$CLI_PATH" ]; then
            "$CLI_PATH" list 2>/dev/null || true
        else
            # Simple key extraction from JSON (no Python needed)
            local val
            for key in config_mode version latest_instance_id instance_count updated_at; do
                val=$(sed -n "s/.*\"${key}\": *\"\{0,1\}\([^\"]*\)\"\{0,1\}.*/\1/p" "$INSTALL_META" | head -1)
                val="${val%,}"  # strip trailing comma from numeric JSON values
                [ -n "$val" ] && echo "  ${key}: ${val}"
            done
        fi
    fi
    if [ -f "$SERVICE_PATH" ]; then
        echo "  service: ${SERVICE_NAME}"
        echo "  manager: ${SERVICE_MANAGER}"
        service_status || true
    fi
}

main() {
    parse_args "$@"
    case "$ACTION" in
        help)
            usage
            exit 0
            ;;
        status)
            load_install_paths
            detect_service_manager
            perform_status
            exit 0
            ;;
    esac

    check_root
    lock_installation
    load_install_paths
    detect_arch
    detect_os
    detect_service_manager
    install_dependencies

    case "$ACTION" in
        install)
            perform_install
            ;;
        upgrade)
            perform_upgrade
            ;;
        uninstall)
            perform_uninstall
            ;;
        *)
            log_error "Unknown action: $ACTION"
            usage
            exit 1
            ;;
    esac
}

main "$@"
