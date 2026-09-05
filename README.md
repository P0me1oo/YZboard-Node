# xboard-node

Node backend for [YZboard](https://github.com/P0me1oo/YZboard). Supports `sing-box` / `xray-core` dual kernels.

> **Disclaimer**: This project is for educational and learning purposes only.

## Features

- Protocols: V2Ray family, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS
- Sync: WebSocket push + REST polling dual channel
- User controls: speed limit, device limit, alive-IP tracking, hot update
- Deploy modes: node mode, machine mode, standalone mode
- Multi-instance: single process binding multiple panels / nodes
- Relay: VLESS entry with Shadowsocks or VLESS landings, including VLESS Encryption
- Machine mode: each panel node can select Xray or sing-box independently; Xray is the default
- 时间校准：为依赖时间戳的 Shadowsocks 2022 链路提供进程内 NTP 校准
- sing-box：`yz.17` 以官方 `v1.14.0` 为基线，固定使用 `P0me1oo/YZ-sing-box v1.14.0-yz.1`，保留用户和路由热更新、Mieru；发布状态见 [兼容矩阵](YZ_COMPATIBILITY.md)，运行和安装验收方法见 [升级验证](docs/singbox-v1.14-validation.md)

## Install

### Docker

```bash
docker run -d --restart=always --network=host \
  -e apiHost=https://panel.com -e apiKey=TOKEN -e nodeID=1 \
  ghcr.io/p0me1oo/yzboard-node:latest
```

### Installer（Linux systemd / OpenRC）

安装器会自动识别正在运行的服务管理器。Debian、Ubuntu 等 systemd 系统使用
`xboard-node.service`；Alpine Linux 使用 OpenRC 的 `xboard-node` 服务，并由
`supervise-daemon` 在进程异常退出后自动拉起。

```bash
# Node mode
curl -fsSL https://github.com/P0me1oo/YZboard-Node/releases/latest/download/install.sh | \
  sudo bash -s -- --mode node --panel https://panel.example.com --token TOKEN --node-id 1 --kernel xray --version latest

# Machine mode
curl -fsSL https://github.com/P0me1oo/YZboard-Node/releases/latest/download/install.sh | \
  sudo bash -s -- --mode machine --panel https://panel.example.com --token TOKEN --machine-id 1 --kernel xray --version latest
```

### Upgrade

从 cedar2025 原版或早期 YZ fork 首次迁移到本 fork：

```bash
curl -fsSL https://github.com/P0me1oo/YZboard-Node/releases/latest/download/install.sh | \
  sudo bash -s -- upgrade --version latest
```

已经安装带 `xbctl` 的 YZ fork 版本后，可直接使用：

```bash
sudo xbctl upgrade --version latest
```

`latest` 只解析 GitHub 最新正式 Release。需要回滚时显式传入旧 Tag，例如
`sudo xbctl upgrade --version v1.13-yz.2`。

## xbctl

Run `xbctl` after installation for help. Common commands:

```bash
xbctl list                          # list all instances
xbctl status                        # running status
xbctl bind add-node --panel URL --token TOKEN --node-id 1
xbctl bind add-machine --panel URL --token TOKEN --machine-id 1
xbctl bind remove-node --panel URL --node-id 1
xbctl service restart
xbctl doctor time                  # 检查 NTP 来源与协议时间偏移
```

`xbctl service status|start|stop|restart|enable|disable|logs` 会使用当前系统的服务管理器。
systemd 日志由 journal 提供；OpenRC 日志写入 `/var/log/xboard-node.log`。安装、升级、
卸载和节点绑定变更不需要手工改用 `rc-service` 或 `systemctl`。

安装器和 `xbctl upgrade` 都会从 `P0me1oo/YZboard-Node` 的同一个 GitHub Release 下载与当前架构匹配的 `xboard-node`、`xbctl`，并使用 Release 中的 `SHA256SUMS` 校验文件完整性。面板使用 `releases/latest/download/install.sh` 获取最新正式安装器，不跟随 `master` 或 `dev` 分支。

## Configuration

Legacy single-panel config is fully compatible. Appending bindings auto-migrates to `instances` format. See `config.yml.example`.

## Extensions

- Xray REALITY 最低客户端版本: [docs-xray-reality.md](docs-xray-reality.md)
- 中转节点与 VLESS 落地: [docs-relay.md](docs-relay.md)
- Custom routes: [docs-custom-routes.md](docs-custom-routes.md)
- Custom outbounds: [docs-custom-outbounds.md](docs-custom-outbounds.md)
- DNS providers (ACME DNS-01): [docs-dns-providers.md](docs-dns-providers.md)

## License

MPL-2.0.
