# 变更记录

## v1.13-yz.17 - 2026-09-05

- sing-box 升级到官方 `v1.14.0` 基线，依赖固定为 `P0me1oo/YZ-sing-box v1.14.0-yz.1`，保留 Node 所需的用户热更新、路由热更新和 Mieru 兼容层；发布状态见兼容矩阵。
- 适配 1.14 的流量跟踪接口，修复 UDP 优化转发绕过下载计数、上传下载回调方向相反的问题；热更新后流量继续按用户累计一次。
- 路由更新先完成解析和启动，失败保留原规则并返回错误；兼容 1.14 的本地、远程规则集多标签格式。
- 修复 Hysteria2/TUIC 已认证会话在用户列表缩短后读取失效下标，以及首次网卡通知关闭新建 QUIC 会话的问题。
- Mieru 用户更新通过原有监听器应用；sing-box 入站使用面板指定的 `listen_ip`。
- 修复 Mieru 停止返回后 UDP 端口尚未释放而导致重新启动失败的问题，并向 Node 返回监听失败。
- 修复用户热更新与协议认证、UDP 收包之间的数据竞争；使用稳定的认证身份，SOCKS/HTTP 清空用户后不会退回匿名代理。
- 显式保留 SOCKS/HTTP 认证要求和 Shadowsocks 多用户模式，清空用户后重启仍拒绝未授权连接；空闲 DNS 连接及入站转交不阻塞路由热更新。
- 流量统计器在监听启动前注册，首个连接也纳入统计。
- 新增 TCP、UDP、两种 SS2022、用户增删、重载、重启和路由回滚测试；CI 测试启用与 Linux 发布一致的功能标签。
- 增加独立网络命名空间下的安装验收脚本，验证安装、重复安装、面板同步、流量上报及 systemd 重启，并清理测试实例。

## v1.13-yz.16 - 2026-09-03

- 中转入口在保留原有 `traffic` 和 `relay_traffic` 的同时，新增 `relay_user_traffic`，按用户和 VLESS 路由拆分落地节点原始上下行流量。
- 节点报告、失败重试和机器模式上报链路完整传递新字段；旧面板和非中转节点继续使用原有报告结构。
- 配套 YZ-Xray-core `v26.7.11-yz.3` 的用户-路由计数器，并记录报告批次兼容约束。

## v1.13-yz.15 - 2026-09-02

- 机器模式支持按面板节点分别选择 Xray 或 sing-box，同一台机器可以混合运行两种内核；未设置的历史节点默认使用 Xray。
- 节点内核选择变化只重启对应节点；机器只包含 Xray 节点时不会创建或启动空的 sing-box 实例。
- 管理端新增位于协议选择左侧的内核下拉，并通过节点同步接口下发 `kernel_type`；中转入口和落地节点仍强制使用 Xray。
- 修复时间校准服务首次检查与停止操作之间的生命周期竞态，确保退出时等待首次 NTP 检查完成。

## v1.13-yz.14 - 2026-09-02

- 新增进程内 NTP 时间校准，默认并行查询多个时间源、校验响应并使用有效偏移中位数；只影响 Node 内嵌协议栈，不修改系统时间，也不管理 systemd-timesyncd、chrony 或同类服务。
- Xray 与 sing-box 共享同一时间服务，覆盖普通 SS2022 入站、VLESS 前置中的 SS2022 中转出站和落地 SS2022 入站；传统 Shadowsocks、VLESS 中转和其他协议保持原行为。
- 新增偏移分级、最近有效结果过期回退、仅在实际使用 SS2022 时触发的状态变化日志，以及 `/healthz` 的 `clock` 详情。时钟降级保持 HTTP 200，避免服务管理器反复重启。
- 新增 `time_sync` 配置、跨实例一致性校验和 `xbctl doctor time` 主动诊断；诊断不依赖面板凭据，且不会修改系统时间。
- 配套 YZ-Xray-core 源码目标版本升级为 `v26.7.11-yz.2`。正式发布前仍需先固定并推送核心提交，再用真实 pseudo-version 更新 Node 的 `go.mod` 和构建信息。

## v1.13-yz.13 - 2026-08-31

- WebSocket 正常连接时保留至少每 5 分钟一次的 REST ETag 全量对账，修复 Redis Pub/Sub、Workerman 重启或短暂网络异常造成的配置和用户推送丢失。
- 配置与用户 REST 拉取改为整次成功后才提交 ETag；用户拉取、配置规范化或初始同步失败时恢复旧 ETag，下一轮会重新获取完整快照。
- 配置哈希只在内核 Reload/Start 成功后提交。配置应用失败会恢复旧配置与哈希并重置 REST 条件请求，相同配置可以继续重试；运行中的内核异常停止后也会用已确认配置自动恢复。
- 在线设备改为每次发送权威全量快照，包括稳定不变和空快照；HTTP 报告与 WebSocket 设备报告不再竞争一次性 Flush，面板设备 TTL 能持续续期并正确清理离线用户。
- REST 报告始终携带空 `alive` 和 `online` 对象，面板可以把旧设备与在线人数明确清零；sing-box 跨节点设备状态有效期由 60 秒调整为 2 分钟，覆盖正常上报与推送抖动。
- 本版本不修改 YZ-Xray-core，继续固定 `v26.7.11-yz.1` 对应 pseudo-version；持久化流量批次需要配套 YZboard `1.7.0` 数据库迁移。

## v1.13-yz.12 - 2026-08-31

- 修复机器模式节点在内核尚未启动时收到首个用户同步后丢弃更新的问题；复制机器或调整节点绑定后，用户配置现在会立即写入并启动对应内核，不再依赖手工重启服务。
- 修复机器模式增量用户同步提前更新状态后导致 UUID 变更误判的问题，并增加停止内核接收首个用户的回归测试。
- 增加用户同步启动失败、增量添加失败和 UUID 替换失败时的状态回滚，避免失败后的相同事件被哈希去重而无法重试。

## v1.13-yz.11 - 2026-08-14

- 安装器新增 Alpine Linux/OpenRC 支持，自动识别正在运行的 systemd 或 OpenRC，并分别生成 systemd unit 或 `/etc/init.d/xboard-node` 服务脚本。
- OpenRC 服务使用 `supervise-daemon` 保持前台进程托管，异常退出后等待 5 秒自动拉起，开机启动加入 `default` runlevel，日志写入 `/var/log/xboard-node.log`。
- OpenRC 启动时逐行解析 `credentials.env` 的 `KEY=VALUE`，不把凭据文件当作 Shell 脚本执行；安装、升级、健康检查、失败回滚、状态查询和卸载统一通过服务管理器抽象执行。
- `xbctl service`、绑定变更、升级、卸载和状态列表同时支持 systemd 与 OpenRC；在 Alpine 的 root 会话中不再依赖默认未安装的 `sudo`。
- 新增服务管理器识别、命令映射、状态归一化、服务文件渲染与安装器 Shell 语法回归测试；本版本不修改 YZ-Xray-core，继续固定 `v26.7.11-yz.1` 对应 pseudo-version。

## v1.13-yz.10 - 2026-08-09

- 中转内部协议在保留 Shadowsocks 的基础上新增 VLESS；入口按逻辑节点生成独立 VLESS 出站，落地生成只含内部 UUID 的 VLESS 入站，用户流量仍只在客户端入口统计一次。
- VLESS 中转支持 RAW/TCP、WebSocket、gRPC、XHTTP、HTTPUpgrade、mKCP 和 Hysteria 传输，并在启动前二次校验 TLS/Reality 组合、唯一路由编号、内部身份和 Hysteria transport auth。
- 支持面板下发 VLESS Encryption：入口出站使用 `encryption`，落地入站使用 `decryption`；两者不写入 Node 本地配置，普通 Shadowsocks 中转结构保持兼容。
- 抽取统一传输构建逻辑供普通 VLESS 入站、VLESS 中转入站和出站复用，补齐 mKCP、Hysteria、RAW/TCP 传输设置并移除 Node 侧 H2 兼容生成路径。
- 新增当前 Xray 有效矩阵、VLESS Encryption、错误组合和 Shadowsocks 回归测试；前置入口与内部链路各 16 组有效组合生成的 JSON 均通过固定 YZ-Xray-core 自带解析器。
- 本版本不修改 YZ-Xray-core，继续固定 `v26.7.11-yz.1` 对应 pseudo-version。

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
