# 变更记录

## v1.13-yz.5 - 2026-07-26

- 支持面板下发的 `relay` 段，实现“单入口、多逻辑节点、多落地出口”的中转拓扑。
- 入口节点在保持唯一 VLESS + Reality 入站的前提下，为每个逻辑节点生成独立的 Shadowsocks 出站（标签 `relay-<节点 ID>`），并按 VLESS 路由编号生成 `vlessRoute` 路由规则；入口自身的编号指向 `direct`。
- 中转路由规则排在自定义规则和内网拦截之后、面板路由组之前，保证选中逻辑节点时出口稳定。
- 落地节点生成只含内部凭据的 Shadowsocks 入站（标签 `relay-in`，`tcp,udp`），不包含任何面板用户；用户列表为空时内核保持监听，不再按普通节点的规则停止。
- 入口读取各中转出站的统计计数器，通过上报中的 `relay_traffic` 上传逻辑节点维度的落地流量；用户流量仍只在入口按真实用户身份统计一次。
- 增加中转配置校验：路由编号范围与唯一性、出站标签唯一且不与自定义出站冲突、中转协议与加密算法白名单；入口和落地都要求 xray 内核，`vlessRoute` 能力 sing-box 不具备。
- 新增简体中文文档 `docs-relay.md`。

## v1.13-yz.4 - 2026-07-26

- 新增 `kernel.reality_min_client_ver` 配置，生成 Xray REALITY 入站时显式注入 `realitySettings.minClientVer`，默认 `0.0.0`；未填写、留空或没有 `kernel` 段时同样使用该默认值，避免 xray-core 回退到内置下限 `26.3.27` 拒绝旧客户端。
- 允许填写其他合法的三段版本号（每段 `0-255`）抬高客户端版本门槛；格式非法时在启动校验阶段直接报错，不带错误配置启动内核。
- `xbctl` 重写配置文件时保留 `kernel.reality_min_client_ver`，改绑节点不会丢失已有取值。
- 新增简体中文文档 `docs-xray-reality.md`，说明取值规则、多实例继承和验证方式。

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
