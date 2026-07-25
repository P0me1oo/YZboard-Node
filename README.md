# xboard-node

Node backend for [YZboard](https://github.com/P0me1oo/YZboard). Supports `sing-box` / `xray-core` dual kernels.

> **Disclaimer**: This project is for educational and learning purposes only.

## Features

- Protocols: V2Ray family, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS
- Sync: WebSocket push + REST polling dual channel
- User controls: speed limit, device limit, alive-IP tracking, hot update
- Deploy modes: node mode, machine mode, standalone mode
- Multi-instance: single process binding multiple panels / nodes

## Install

### Docker

```bash
docker run -d --restart=always --network=host \
  -e apiHost=https://panel.com -e apiKey=TOKEN -e nodeID=1 \
  ghcr.io/p0me1oo/yzboard-node:v1.13-yz.2
```

### Installer (Linux systemd)

```bash
# Node mode
curl -fsSL https://raw.githubusercontent.com/P0me1oo/YZboard-Node/v1.13-yz.2/install.sh | \
  sudo bash -s -- --mode node --panel https://panel.example.com --token TOKEN --node-id 1 --version v1.13-yz.2

# Machine mode
curl -fsSL https://raw.githubusercontent.com/P0me1oo/YZboard-Node/v1.13-yz.2/install.sh | \
  sudo bash -s -- --mode machine --panel https://panel.example.com --token TOKEN --machine-id 1 --version v1.13-yz.2
```

### Upgrade

从 cedar2025 原版或早期 YZ fork 首次迁移到本 fork：

```bash
curl -fsSL https://raw.githubusercontent.com/P0me1oo/YZboard-Node/v1.13-yz.2/install.sh | \
  sudo bash -s -- upgrade --version v1.13-yz.2
```

已经安装 `v1.13-yz.2` 或更高 YZ fork 版本后，可直接使用：

```bash
sudo xbctl upgrade --version v1.13-yz.2
```

## xbctl

Run `xbctl` after installation for help. Common commands:

```bash
xbctl list                          # list all instances
xbctl status                        # running status
xbctl bind add-node --panel URL --token TOKEN --node-id 1
xbctl bind add-machine --panel URL --token TOKEN --machine-id 1
xbctl bind remove-node --panel URL --node-id 1
xbctl service restart
```

安装器和 `xbctl upgrade` 都会从 `P0me1oo/YZboard-Node` 的 GitHub Release 下载与当前架构匹配的 `xboard-node`、`xbctl`，并使用 Release 中的 `SHA256SUMS` 校验文件完整性。

## Configuration

Legacy single-panel config is fully compatible. Appending bindings auto-migrates to `instances` format. See `config.yml.example`.

## Extensions

- Custom routes: [docs-custom-routes.md](docs-custom-routes.md)
- Custom outbounds: [docs-custom-outbounds.md](docs-custom-outbounds.md)
- DNS providers (ACME DNS-01): [docs-dns-providers.md](docs-dns-providers.md)

## License

MPL-2.0.
