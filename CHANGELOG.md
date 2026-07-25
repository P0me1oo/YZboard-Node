# 变更记录

## v1.13-yz.2 - 2026-07-25

- 安装器和 `xbctl upgrade` 的默认 Release 来源改为 `P0me1oo/YZboard-Node`。
- 安装器、systemd 单元和发布工作流统一使用 YZ fork 仓库与 GHCR 地址。
- Release 增加 `SHA256SUMS`，安装和升级时校验两个下载二进制。
- Release 使用与 `go.mod` 一致的 Go `1.26.4` 构建，后续 CI 工具链同步为 `1.26.4`。
- 保留 `v1.13-yz.1` 的 Xray v26.7.11、Hysteria2 用户识别、限速、Dispatcher、统计和上报兼容修复。

## v1.13-yz.1 - 2026-07-25

- 固定 YZ-Xray-core `v26.7.11-yz.1` fork commit。
- 修复 Hysteria2 用户账户转换和流量批次幂等上报。
