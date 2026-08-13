package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	serviceName             = "xboard-node"
	systemdServiceName      = serviceName + ".service"
	systemdServiceFilePath  = "/etc/systemd/system/" + systemdServiceName
	openRCServiceFilePath   = "/etc/init.d/" + serviceName
	openRCLogPath           = "/var/log/" + serviceName + ".log"
	systemdRuntimeDirectory = "/run/systemd/system"
)

type serviceManager string

const (
	serviceManagerSystemd serviceManager = "systemd"
	serviceManagerOpenRC  serviceManager = "openrc"
)

type serviceCommandSpec struct {
	name string
	args []string
}

func detectServiceManager() (serviceManager, error) {
	return detectServiceManagerWith(exec.LookPath, func(path string) bool {
		info, err := os.Stat(path)
		return err == nil && info.IsDir()
	})
}

func detectServiceManagerWith(
	lookPath func(string) (string, error),
	isDir func(string) bool,
) (serviceManager, error) {
	if _, err := lookPath("systemctl"); err == nil && isDir(systemdRuntimeDirectory) {
		return serviceManagerSystemd, nil
	}
	for _, command := range []string{"openrc-run", "rc-service", "rc-update", "supervise-daemon"} {
		if _, err := lookPath(command); err != nil {
			return "", errors.New("a running systemd or OpenRC service manager is required")
		}
	}
	return serviceManagerOpenRC, nil
}

func serviceFilePathFor(manager serviceManager) (string, error) {
	switch manager {
	case serviceManagerSystemd:
		return systemdServiceFilePath, nil
	case serviceManagerOpenRC:
		return openRCServiceFilePath, nil
	default:
		return "", fmt.Errorf("unsupported service manager: %s", manager)
	}
}

func serviceCommandFor(manager serviceManager, action string, useSudo bool, rest []string) (serviceCommandSpec, error) {
	var spec serviceCommandSpec
	switch manager {
	case serviceManagerSystemd:
		switch action {
		case "status":
			spec = serviceCommandSpec{name: "systemctl", args: append([]string{"status", systemdServiceName, "--no-pager"}, rest...)}
		case "start", "stop", "restart", "enable", "disable":
			spec = serviceCommandSpec{name: "systemctl", args: append([]string{action, systemdServiceName}, rest...)}
		case "logs":
			if len(rest) == 0 {
				rest = []string{"-f"}
			}
			spec = serviceCommandSpec{name: "journalctl", args: append([]string{"-u", systemdServiceName}, rest...)}
		default:
			return serviceCommandSpec{}, fmt.Errorf("unknown service command: %s", action)
		}
	case serviceManagerOpenRC:
		switch action {
		case "status", "start", "stop", "restart":
			spec = serviceCommandSpec{name: "rc-service", args: append([]string{serviceName, action}, rest...)}
		case "enable":
			spec = serviceCommandSpec{name: "rc-update", args: []string{"add", serviceName, "default"}}
		case "disable":
			spec = serviceCommandSpec{name: "rc-update", args: []string{"del", serviceName, "default"}}
		case "logs":
			if len(rest) == 0 {
				rest = []string{"-f"}
			}
			spec = serviceCommandSpec{name: "tail", args: append(rest, openRCLogPath)}
		default:
			return serviceCommandSpec{}, fmt.Errorf("unknown service command: %s", action)
		}
	default:
		return serviceCommandSpec{}, fmt.Errorf("unsupported service manager: %s", manager)
	}

	if useSudo {
		spec.args = append([]string{spec.name}, spec.args...)
		spec.name = "sudo"
	}
	return spec, nil
}

func runManagedService(manager serviceManager, action string, useSudo bool, rest ...string) error {
	spec, err := serviceCommandFor(manager, action, useSudo, rest)
	if err != nil {
		return err
	}
	return runCommand(spec.name, spec.args...)
}

func runDetectedManagedService(action string, useSudo bool, rest ...string) error {
	manager, err := detectServiceManager()
	if err != nil {
		return err
	}
	return runManagedService(manager, action, useSudo, rest...)
}

func reloadServiceManager(manager serviceManager) error {
	if manager != serviceManagerSystemd {
		return nil
	}
	return runCommand("systemctl", "daemon-reload")
}

func serviceState() string {
	manager, err := detectServiceManager()
	if err != nil {
		return "unknown"
	}
	return serviceStateFor(manager)
}

func serviceStateFor(manager serviceManager) string {
	switch manager {
	case serviceManagerSystemd:
		out, err := exec.Command("systemctl", "is-active", systemdServiceName).CombinedOutput()
		state := strings.TrimSpace(string(out))
		if state != "" {
			return state
		}
		if err != nil {
			return "unknown"
		}
		return state
	case serviceManagerOpenRC:
		out, err := exec.Command("rc-service", serviceName, "status").CombinedOutput()
		return normalizeOpenRCState(string(out), err)
	default:
		return "unknown"
	}
}

func normalizeOpenRCState(output string, err error) string {
	state := strings.ToLower(strings.TrimSpace(output))
	switch {
	case strings.Contains(state, "status: started"):
		return "active"
	case strings.Contains(state, "status: crashed") || strings.Contains(state, "status: unsupervised"):
		return "failed"
	case strings.Contains(state, "status: stopped") || strings.Contains(state, "status: inactive"):
		return "inactive"
	case err != nil:
		return "unknown"
	default:
		return "active"
	}
}

func serviceDefinitionFor(manager serviceManager) (string, []byte, os.FileMode, error) {
	switch manager {
	case serviceManagerSystemd:
		unit := fmt.Sprintf(`[Unit]
Description=Xboard Node Backend
Documentation=https://github.com/P0me1oo/YZboard-Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=%s
EnvironmentFile=-%s
ExecStart=%s -c %s
Restart=always
RestartSec=5
LimitNOFILE=1048576
NoNewPrivileges=true
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, defaultInstallRoot, defaultCredentialsPath, defaultBinaryPath, defaultConfigPath)
		return systemdServiceFilePath, []byte(unit), 0o644, nil
	case serviceManagerOpenRC:
		script := fmt.Sprintf(`#!/sbin/openrc-run

name="Xboard Node Backend"
description="YZboard node backend"
supervisor=supervise-daemon
command="%s"
command_args="-c %s"
directory="%s"
pidfile="/run/%s.pid"
output_log="%s"
error_log="%s"
respawn_delay=5
respawn_max=0
no_new_privs=true
rc_ulimit="-n 1048576"

depend() {
    need net
}

start_pre() {
    local line key value
    if [ -r "%s" ]; then
        while IFS= read -r line || [ -n "$line" ]; do
            case "$line" in
                ""|\#*) continue ;;
            esac
            key=${line%%%%=*}
            if [ "$key" = "$line" ]; then
                eerror "Invalid credentials entry: missing '='"
                return 1
            fi
            case "$key" in
                ""|[0-9]*|*[!A-Za-z0-9_]*)
                    eerror "Invalid credentials key: $key"
                    return 1
                    ;;
            esac
            value=${line#*=}
            export "${key}=${value}"
        done < "%s"
    fi
    checkpath -f -m 0640 -o root:root "%s"
}
`, defaultBinaryPath, defaultConfigPath, defaultInstallRoot, serviceName, openRCLogPath, openRCLogPath,
			defaultCredentialsPath, defaultCredentialsPath, openRCLogPath)
		return openRCServiceFilePath, []byte(script), 0o755, nil
	default:
		return "", nil, 0, fmt.Errorf("unsupported service manager: %s", manager)
	}
}

func regenerateServiceFile() error {
	manager, err := detectServiceManager()
	if err != nil {
		return err
	}
	path, content, mode, err := serviceDefinitionFor(manager)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, mode)
}
