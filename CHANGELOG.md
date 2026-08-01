# 变更记录

## v1.13-yz.9 - 2026-08-02

- 修复 SS2022 用户 UUID 轮换后，Xray `UserManager` 热更新路径直接使用原始 UUID、导致动态加入的用户密钥与订阅不一致的问题。
- SS2022 完整启动和动态用户更新现在复用同一套固定长度截取与 Base64 编码规则，并增加 AES-128/AES-256 两种用户密钥的一致性回归测试。
- 本次只修复 Node 的 SS2022 用户同步，不修改订阅链接生成、缓存或失效逻辑，也不修改 `YZ-Xray-core`。

## v1.13-yz.8 - 2026-08-01

- 修复用户 UUID 轮换时，Node 只尝试添加新凭据、未将旧凭据作为待删除用户的问题；Xray 动态更新现在先删除旧凭据，再添加新凭据。
- Xray `UserManager` 的用户构建、添加或删除失败时不再吞掉错误，也不会提前推进内部用户状态；动态更新失败后会尝试使用完整用户集重建内核。
- 机器模式收到“同一用户 ID、不同 UUID”的增量消息时，改用完整用户集替换，并增加 UUID 替换、元数据变更和错误回退的回归测试。
- 本版本不修改面板的订阅 Token、订阅链接、缓存或旧链接更新逻辑，也不修改 YZ-Xray-core。

## v1.13-yz.7 - 2026-07-27

- 自定义出站支持直连与拦截：面板可以填 `direct`/`freedom` 与 `block`/`blackhole`，由 Node 翻译成目标内核的原生名（xray 为 `freedom`/`blackhole`，sing-box 为 `direct`/`block`）。此前两个内核的白名单都不含这些协议，配了就会让节点启动失败。
- 支持 `settings.send_through` 绑定出站源地址。xray 的 `sendThrough` 是 outbound 级字段，无法写在 `settings` 内，Node 会在生成配置时提升上来，并且不修改调用方持有的 settings。
- 直连与拦截出站不再强制要求 `settings`，无参数的拦截出站可以直接使用。
- `settings` 内的其余字段保持原样透传，不做跨内核翻译：两个内核的绑定与解析语义并不等价（sing-box 的 `prefer_ipv6` + `fallback_delay` 是连接级回落且可同时绑定 v4/v6 源地址，xray 的 `domainStrategy` 只影响解析顺序且 `sendThrough` 只能填一个地址），自动翻译会掩盖差异。
- 以上均使用 Xray 自带能力，不改动 YZ-Xray-core。

## v1.13-yz.6 - 2026-07-27

- 安装器默认内核由 `singbox` 改为 `xray`，一键安装命令不再需要显式传 `--kernel xray`；用新安装器覆盖旧安装时，该实例的内核会一并切换到 xray。
- 未显式指定 `--kernel` 且 `--node-type` 属于 xray 无法承载的入站协议（`tuic`、`naive`、`anytls`、`mieru`、`socks`、`http`）时，安装器回退到 sing-box 并给出提示，避免装完节点直接起不来。显式传 `--kernel xray` 可覆盖该回退。
- 新增 `xbctl config kernel <xray|singbox>`，用于切换已安装节点的内核。默认作用于全部实例，`--instance` 可只改一个；切到 xray 时会拒绝 xray 无法承载的入站协议，`--force` 可强制执行；重复执行保持幂等，并保留配置文件中原有的实例 ID。

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
