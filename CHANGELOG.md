# 变更记录

## v1.13-yz.3 - 2026-07-26

- 修复 Xray Hysteria2 入站缺少协议层 `settings.version` 导致的 `version != 2` 启动失败，并使用 Xray 自身 JSON 解析器增加回归测试。
- 进程级 SIGINT/SIGTERM 监听跨配置热重载持续生效，避免重载窗口丢失 systemd 停止信号。
- `/healthz` 根据节点服务的启动、运行和失败状态返回 200 或 503；机器模式失败节点可在后续重新发现时重试并恢复健康。
- `sync.devices` 同时兼容字符串数组、PHP 数字键对象和空数组，保持在线设备与连接状态同步。
- GitHub Release 增加 `install.sh` 资产，支持稳定的 `releases/latest/download/install.sh` 安装入口；具体二进制仍由 Release Tag 和 `SHA256SUMS` 固定。
- 继续使用 YZ-Xray-core `v26.7.11-yz.1`，保留 Hysteria2 用户识别、限速、Dispatcher、统计和上报能力。

## v1.13-yz.2 - 2026-07-25

- 安装器和 `xbctl upgrade` 的默认 Release 来源改为 `P0me1oo/YZboard-Node`。
- 安装器、systemd 单元和发布工作流统一使用 YZ fork 仓库与 GHCR 地址。
- Release 增加 `SHA256SUMS`，安装和升级时校验两个下载二进制。
- Release 使用与 `go.mod` 一致的 Go `1.26.4` 构建，后续 CI 工具链同步为 `1.26.4`。
- 保留 `v1.13-yz.1` 的 Xray v26.7.11、Hysteria2 用户识别、限速、Dispatcher、统计和上报兼容修复。

## v1.13-yz.1 - 2026-07-25

- 固定 YZ-Xray-core `v26.7.11-yz.1` fork commit。
- 修复 Hysteria2 用户账户转换和流量批次幂等上报。
