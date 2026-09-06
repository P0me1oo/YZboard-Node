# YZboard-Node 兼容矩阵

本文件记录可发布的 Node 构建与内嵌内核之间的固定关系。构建上线时必须使用明确的 Node Release Tag 和固定的 Xray fork commit，不能依赖 `main` 或其他移动分支。

## 当前源码与已发布构建

| 项目 | 标识 |
| --- | --- |
| Node 源码目标版本 | `v1.13-yz.19`（尚未发布） |
| Node 发布版本 | `v1.13-yz.16` |
| Node 适用分支 | `upgrade/singbox-v1.14.0` |
| Node 上游发布基线 | `v1.13` |
| Node 上游基线 commit | `0a29338e1f102a462363ce3527417029f89bab28` |
| Node Release Tag 对应 commit | `v1.13-yz.19` 尚未发布 |
| Node `yz.19` 修复基线 | `89c2753390356f51df3d8fc133ae8064fa8ed669` |
| Node `yz.17` 验证构建 commit | `ada7bb60b18bf14b80e171030b82bc0f3412beb6` |
| Node Release 构建工具链 | `Go 1.26.4`（`go.mod` 要求 `go 1.26`） |
| Node Release 构建 | 最近已发布 `v1.13-yz.16`；`v1.13-yz.19-test` 的 Node、xbctl 双架构开发构建已完成，尚未发布 |
| Node Docker 标签 | 本次尚未发布 `v1.13-yz.19` 镜像 |
| 最近已发布 Docker manifest（yz.16） | OCI index `sha256:f3e0895ebc04ac603158a5b96413e7695e5c7a1abd596ee864d7d4852b0b4665`；包含 `linux/amd64` 与 `linux/arm64` |
| Node Docker OCI 标识 | 待发布 |
| YZboard 兼容版本 | `1.9.0`（待发布；用户-落地流量归属需与 Node `v1.13-yz.16` 成套使用） |
| 最近已发布 YZboard 兼容代码 | `cf698392cd0b0623876b5166ab31b10fea2cb889`（面板 `v1.4.0`） |
| Xray 官方仓库 | `XTLS/Xray-core` |
| Xray 上游预发布 Tag | `v26.7.11` |
| Xray 上游 Tag commit | `50231eaff98ccc31b5cbd247a721c16e97fe5ec1` |
| YZ-Xray-core 源码目标版本 | `v26.7.11-yz.3`（待发布） |
| YZ-Xray-core 当前已固定版本 / commit | `v26.7.11-yz.3` / `601226e180d3684a5eabb8bc901c99f499398db1` |
| Node 当前 Xray replace | `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260903142229-601226e180d3` |
| YZboard 兼容标识 | `xray-v26.7.11-yz.3`（面板版本 `1.9.0`） |
| sing-box `require` 版本 | `v1.14.0` |
| sing-box 实际 replacement | `github.com/P0me1oo/YZ-sing-box v1.14.0-yz.1` |
| sing-box 官方基线 | `v1.14.0` / `0b8995879f29a9b98ee027bc17b75e101445b238` |
| sing-box 兼容仓库 | [P0me1oo/YZ-sing-box](https://github.com/P0me1oo/YZ-sing-box)；保留用户、路由热更新和 Mieru |
| sing-box 兼容 Tag / commit | `v1.14.0-yz.1` / `f47d4d565a4371cf46b6c462612fc085f634a6af` |
| sing-box 模块校验值 | `h1:oyWPL6yHrnYmtLhvM54ygF0fuRbIxIKf5K42rPafUeU=` |
| sing-box 验证依赖 | 正式 `go.mod` 使用远程固定 Tag；前期隔离实测使用同一兼容源码的本地 replacement |
| AnyTLS 上游基线 | `anytls/sing-anytls v0.0.11` / `130d2e61b8895727bfed4942c535e91b246a9603` |
| AnyTLS 实际 replacement | `./compat/sing-anytls`，随 Node 固定提交构建；仅修复流关闭状态和回调的并发访问，来源与移除条件见 [补丁说明](compat/sing-anytls/README.yz.md) |

Node 自身版本保持独立，不伪装成 Xray 版本。Node 延续上游 `v1.13` 版本线；`yz.5` 支持首版 Shadowsocks 中转，`yz.6` 至 `yz.9` 延续既有安装、出站和用户同步修订，`yz.10` 新增 VLESS 落地、VLESS Encryption 和当前 Xray 传输矩阵，`yz.11` 为安装器与 `xbctl` 增加 Alpine Linux/OpenRC 生命周期支持，`yz.12` 修复机器模式首个用户同步与失败回滚，`yz.13` 增加 REST/WS 双通道对账、ETag 事务回滚和权威设备快照，`yz.14` 增加 SS2022 进程内时间校准、健康状态和主动诊断，`yz.15` 增加机器模式节点级内核选择，`yz.16` 增加用户-落地节点流量归属上报，`yz.19` 明确失败配置停止和健康失败状态，不自动恢复旧配置。Xray 的上游版本、YZ fork patch 版本和 Node 发布版本分别记录，便于升级、回滚和定位构建来源。

先前的 `v0.1.0-yz.1` Tag 保留用于审计，但其版本低于上游 `v1.13`，不作为部署或升级目标，也不创建对应 Release。

`yz.17` 将 sing-box 官方基线更新为 `v1.14.0`，修复 UDP 统计、用户更新并发、稳定认证身份、路由更新回滚和 Mieru 监听生命周期。兼容源码与 Node 的 Xray 依赖独立；本次没有修改 YZboard 或 YZ-Xray-core。

正式 `go.mod` 与 `go.sum` 已固定兼容核心的远程 Tag 和校验值。Node 当前目标 `v1.13-yz.19` 尚未发布，源码修复和构建验证不改变已发布版本或服务器安装版本。

## `yz.19` 失败状态修复（2026-09-07）

配置、协议、端口、出站或用户应用失败时，Node 直接记录包含操作、内核和底层原因的错误，停止当前内核并将节点健康状态标记为失败。面板最新的失败配置和用户快照会保留为待修正状态；相同失败快照不会被定时采样、REST 或 WebSocket 反复启动，只有配置或用户实际变化后才允许重新尝试。首次启动失败的进程继续保持控制通道和健康端点，等待修正后的面板配置。

本次只修改 YZboard-Node，未修改 YZboard 或 YZ-Xray-core。详细行为和测试命令见 [Node 修复验证](docs/node-reliability-validation.md)。

配置校验在节点服务内统一执行，REST、WebSocket 和首次同步均保留无效快照并停止对应内核；用户增删不能绕过配置或证书错误。删除问题用户会尝试启动剩余用户，单个节点或实例初始化失败不会取消同一进程的其他节点。整体健康端点返回 503 表示存在失败项，其他节点仍可继续转发。

首次发布前 CI 在 AnyTLS 用户删除测试中捕获上游 `dieErr`、`dieHook` 数据竞争，因此未发布该构建。Node 内增加固定上游源码的最小兼容补丁，保留完整生命周期测试，并单独执行依赖包的并发测试。两个主内核的固定版本未变；AnyTLS 本地 replacement 的身份由 Node 提交及原始文件校验清单共同记录。

`yz.19` 最终开发验证已完成：Windows amd64 通过 18 个包、586 项测试及子测试；YT-HK Linux amd64 `-race` 通过 5 个包、331 项测试及子测试，失败、跳过和数据竞争均为 0。AnyTLS 底层关闭回归重复 50 轮、真实用户生命周期重复 20 轮全部通过。12 个实际进程场景通过，包括两种多节点模式下的端口冲突、无效出站与初始 HTTP 失败隔离，以及等待 18 秒的退出报告和失败重试。Node、xbctl 双架构构建和 amd64 运行时版本检查通过；arm64 未进行实机运行。验收包为 `runtime-validation.tar.xz`，大小 130250888 字节，SHA256 `aa3f59474867e7c727beb793af1186044b44ab97dee2298da40f49bcbd67f424`。原有 7 个监听未变化，临时目录、测试进程和上传包已清理，Netcatty 会话已关闭。

进程退出等待两分钟，安装器的 systemd/OpenRC 模板等待 150 秒。已有部署通过 xbctl 单独替换二进制时还需同步服务停止等待设置；Docker/Compose 也应设为 150 秒。发布 CI 核对完整来源提交、干净源码标识与双架构元数据，Docker 版本检查通过后才更新正式标签。

## `yz.18` 可靠性修复（2026-09-06）

本次只修改 Node，沿用上表中的 sing-box、Xray 固定依赖和既有面板接口。修复范围包括内核监听退役、出站应用与失败恢复、空用户同步、REST 用户重试、跨实例流量累计和退出上报。详细不变量、执行命令和验证状态见 [修复验证](docs/node-reliability-validation.md)。

`yz.18` 构建来自上表基线的未提交修复，必须保留实际的 `vcs.modified=true` 标识。发布前需要提交、固定 Tag、重新构建并更新本表；下文 `yz.17` 的历史校验值不能用于本次修复产物。

Windows amd64 全量测试通过：17 个包、555 项测试及子测试；`go vet` 与安装器脚本检查通过。YT-HK 上实际执行四个 Linux amd64 `-race` 测试包，共 301 项测试及子测试，全部通过、无跳过、无数据竞争。Node、xbctl 双架构构建和 amd64 运行时版本检查通过；arm64 尚未进行实机运行。测试产物和结果按 [修复验证](docs/node-reliability-validation.md) 归档，本次远程测试文件已清理。

验证归档为 `yznode-v1.13-yz.18-test-linux-89c2753.tar.gz`，大小为 206497020 字节，SHA256 为 `7ced3892a85b142ff873efea731149123f3c61afb2287f0794b1c69d06a719a3`。归档包含八个二进制及其校验值、源码清单、构建元数据和本次 Linux 实测结果；逐文件校验值以包内 `SHA256SUMS` 为准。

## `yz.17` 固定依赖验证（2026-09-05）

远程 Tag 指向上表中的完整兼容提交。下载的 Go 模块中，1187 份源码、模块文件及第三方来源文件与该提交的原始 Git 内容逐字节一致；657 项模块版本与隔离实测使用的依赖列表一致。`go mod verify` 通过，Xray replacement 保持原有固定提交。

使用正式 `go.mod` 在 Windows amd64 执行 `go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./...`，全部 17 个测试包通过，共 522 项测试及子测试，没有失败或跳过。本轮为普通测试；相同源码和依赖版本的 Linux race 结果见下文。

以下产物从 `ada7bb60b18bf14b80e171030b82bc0f3412beb6` 的干净源码构建，版本为 `v1.13-yz.17`，构建时间为 `2026-09-05T14:14:13Z`，工具链为 `Go 1.26.4`，`CGO_ENABLED=0`。`go version -m` 已确认两个 Linux 目标架构、远程固定的 sing-box 与 Xray 模块，以及一致的源码提交和 `vcs.modified=false`。完整构建参数和校验值由产物附带的 `build-metadata.json` 与 `SHA256SUMS` 保存。

| 固定依赖构建产物 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `9e71a840ab5716eb005c7ad4d8ff7fbb8f5c42a335dd709cb32e6f618d656ddf` |
| `xboard-node-linux-arm64` | `0f7884d02c2902a7df1198da3ffc8882af444b62e212f83a41ac5fe985a011dc` |
| `xbctl-linux-amd64` | `c7ba1724e0852168bb0795d09fe7944d6bc6c9daa4b349dddaba589040d961cb` |
| `xbctl-linux-arm64` | `5b0086dde3af4a69f960e77962d1af7a8f40040921e9177be4888b573fb13bba` |
| `install.sh` | `8e7c5c21210020f283ebc793e7c6deb8b389648060dc689b88653293c6c6d8dc` |

## `yz.17` 开发验证（2026-09-05）

本次结果对应 `v1.13-yz.17-test`。Node 使用升级分支的未提交改动及上表中的本地兼容核心；不代表 `v1.13-yz.17` 已发布。测试方法和未覆盖的场景见 [sing-box 升级验证](docs/singbox-v1.14-validation.md)。

| 检查 | 结果 |
| --- | --- |
| 核心普通测试 | `./route ./route/rule` 通过；包含规则集事务、并发匹配、初始网络通知和连接转交 |
| Linux amd64 race | Node 全部 17 个测试包及核心 `route` 包通过；未跳过测试，数据竞争报告为 0 |
| 协议与用户生命周期 | 13 个 TCP 场景、10 个 UDP 场景通过；覆盖重复增删、已有连接、新流、重载及恢复 |
| 空用户重启 | SOCKS、HTTP、普通 Shadowsocks、SS2022 AES-128 均拒绝未授权连接，重新添加用户后恢复 |
| Mieru 监听生命周期 | TCP/UDP 停止后重启通过，端口被占用时明确返回启动错误 |
| 安装器检查 | 服务管理器测试、Bash 语法和 Python 语法检查通过 |
| YT-HK 隔离安装 | 使用最终 amd64 产物完成安装、重复安装、模拟面板用户同步、路由回滚及恢复、流量与状态上报、systemd 重启 |
| 原服务与清理 | 原服务 PID 和程序校验值未变化；一次性实例、凭据、上传程序和测试日志均已清理，测试 SSH 会话已关闭 |
| 目标架构 | `linux/amd64`、`linux/arm64` 的 Node 和 xbctl 构建通过；arm64 仅完成构建与元数据检查 |

构建时间为 `2026-09-05T13:13:22Z`，工具链为 `Go 1.26.4`，可安装产物使用 `CGO_ENABLED=0`。`go version -m` 已确认两个目标架构、sing-box `v1.14.0` 的本地 replacement、固定的 YZ-Xray-core replacement，以及 Node 的 `vcs.revision=7802e87136e62ebfc79048207b39323556c7cabc`、`vcs.modified=true`。这些是开发构建的实际标识；正式依赖固定后必须重新提交并构建发布产物。

| 开发产物 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `8734c3e69d773ad8167545a5566d1eb9126fe7bb93b0a3cf142c17f1a44e9ae6` |
| `xboard-node-linux-arm64` | `191840244cbb1d4509e9ae5e12642349e029e3f7957f45716fa8974fb3424242` |
| `xbctl-linux-amd64` | `a46b5e529b319be5b85a92c78b77a8bb078dadb57574d0799b48b78a1e183272` |
| `xbctl-linux-arm64` | `b8442e4a1d267c381353718e3049a761289cadd1300a94e6b10b74b8cf9bba3f` |

真实安装使用的开发包 SHA-256 为 `369862973dcc38a9551fc8d04f161bc4084a93ea0da76d2171f94d6df0fbf245`；包含全部 18 个测试二进制的 race 包 SHA-256 为 `eedf0cda1bf48f5b8b2c91985f0482c1f06322e79ab5875e024465bea4fd3ca0`。上传后已在服务器重新校验。

上述校验值对应使用本地 replacement 的开发产物。使用远程固定依赖重新构建时，应以该次产物附带的 `build-metadata.json` 和 `SHA256SUMS` 为准，不沿用开发产物的校验值。

## 兼容约束

- Hysteria2 用户转换使用 Xray v26.7.11 的 `hysteria/account.MemoryAccount{Auth: ...}`，同时保留 `MemoryUser.Email` 的 `user@<id>` 映射。
- Xray fork 提供的 Dispatcher、用户级限速、统计计数器和在线 IP/连接状态能力继续由 Node 使用。
- Node 的流量方向保持 `[upload, download]`，由内核累计计数器交给 tracker 计算增量，再由面板客户端上报。
- 中转入口的 `relay_user_traffic` 形状为 `user_id => logical_node_id => [upload, download]`，只用于用户-落地归属分析，不参与套餐扣除；`relay_traffic` 继续负责落地节点总量。
- 每次刷出的报告批次带有进程启动标识和递增序号组成的 `report_id`；HTTP 失败时保留完整批次并复用 ID，避免面板重复累计。
- Xray REALITY 入站的 `realitySettings.minClientVer` 由 Node 显式写入，默认 `0.0.0`，可通过 `kernel.reality_min_client_ver` 覆盖。缺省该字段时 v26.7.11 会使用内置下限 `26.3.27`，低于该版本的客户端握手会被拒绝。
- v26.7.11 已移除未加密 Shadowsocks。历史配置中的 `none`/`plain` 会显式返回错误，不会静默转换成其他加密算法。
- `go.mod` 的 Xray `require` 版本只用于保持模块路径兼容；实际代码由 `replace` 固定到上表中的 fork pseudo-version。提交前应使用 `go list -m -json github.com/xtls/xray-core` 复核替换路径和版本。
- 中转拓扑依赖 Xray 的 VLESS 路由值能力：认证前清零 UUID 第 7、8 字节，认证后按原始字节还原，并由路由规则的 `vlessRoute` 匹配。该能力来自上游 `v26.7.11`，sing-box 不具备，因此入口和落地节点都要求 xray 内核。
- 面板 `relay` 段与 `relay_traffic` 上报字段属于 YZboard `1.1.0` 起的接口；旧面板不下发该字段时 Node 行为不变。
- 安装器从 `yz.6` 起默认写入 `kernel.type: xray`；`yz.15` 起代码层缺省也按 Xray 处理空值。机器模式下节点的面板 `kernel_type` 优先于机器级默认值；独立实例仍可显式执行 `xbctl config kernel <xray|singbox>` 切换。
- xray 可承载的入站协议为 vmess、vless、trojan、shadowsocks、hysteria；tuic、naive、anytls、mieru、socks、http 只能由 sing-box 承载。安装器和 `xbctl config kernel` 都会在未显式确认时拒绝把这些节点切到 xray。
- 自定义出站从 `yz.7` 起接受 `direct`/`freedom` 与 `block`/`blackhole`，由 Node 翻译成目标内核的原生名；`settings.send_through` 在 xray 下提升为 outbound 级的 `sendThrough`。`settings` 内其余字段原样透传，需按目标内核的字段名填写，跨内核切换时要同步调整。
- 从 `yz.8` 起，同一用户 ID 的 UUID 变化会被视为凭据替换，Xray `UserManager` 必须先删除旧凭据再添加新凭据；任一步失败都不得推进 Node 内部用户状态，并由 Service 尝试使用完整用户集重建内核。
- 从 `yz.9` 起，Xray 的 Shadowsocks 2022 动态用户密钥按面板约定从 UUID 前 16 或 32 字节生成标准 Base64；静态启动配置和运行时增删用户必须得到同一密钥。
- 从 `yz.10` 起，中转 child/landing 同时接受 Shadowsocks 和 VLESS。VLESS 的入口客户端参数放在 `relay.children[].vless`，落地内部身份放在 `relay.vless`；服务端顶层继续承载 `decryption`、Reality 私钥和证书配置。
- VLESS relay 的传输矩阵固定为 RAW/TCP、WS、gRPC、XHTTP、HTTPUpgrade、mKCP、Hysteria；Reality 只允许 RAW/TCP、gRPC、XHTTP，Hysteria 必须使用 TLS，H2/HTTP 和 mKCP header/seed 会在启动前拒绝。
- `yz.10` 继续使用当前 `go.mod` 固定的 YZ-Xray-core pseudo-version，不需要核心补丁。入口和落地 JSON 由该核心自带解析器覆盖验证。
- `yz.11` 的安装器自动识别正在运行的 systemd 或 OpenRC。OpenRC 路径固定使用 `/etc/init.d/xboard-node`、`supervise-daemon` 和 `default` runlevel，日志写入 `/var/log/xboard-node.log`；凭据仍保存在权限为 `0600` 的 `/etc/xboard-node/credentials.env`，启动脚本只按 `KEY=VALUE` 解析，不执行其中内容。
- `yz.14` 的 SS2022 时间校准只在 Node 进程内提供可选时间函数，不修改系统时间。Xray 需要 `v26.7.11-yz.2` 的上下文时间服务补丁；sing-box 使用相同服务，避免每个实例重复查询 NTP。
- `yz.15` 的机器模式按节点创建独立内核服务；只有发现到 sing-box 节点时才创建 sing-box 服务。节点内核变化只重启目标节点，Xray-only 机器不会启动空的 sing-box。

## `yz.14` 时间校准兼容约束

- 普通 SS2022 入站、VLESS 前置中的 SS2022 出站和落地 SS2022 入站共享同一校准结果。VLESS 客户端入口本身不依赖该时间戳。
- 校准器默认并行查询三个 NTP 源并使用有效偏移中位数；查询失败不会猜测时间，最近成功结果超过三个查询周期后回退系统时间。
- `/healthz` 的时钟降级保持 HTTP 200；只有节点组件启动中或失败继续返回 HTTP 503。`xbctl doctor time` 的异常状态返回非零退出码。
- 当前 `go.mod` 已固定 YZ-Xray-core `v0.0.0-20260903142229-601226e180d3`，对应 fork `v26.7.11-yz.3` 和 commit `601226e180d3684a5eabb8bc901c99f499398db1`。正式构建前仍需确认该核心提交可回滚，并不得改回本地路径 replace 或移动分支。

## `yz.13` 同步兼容约束

- WebSocket 仍用于即时推送，但 Node 在连接正常时至少每 5 分钟执行一次 REST ETag 对账；面板推送丢失不会再让配置或用户状态长期停留在旧版本。
- 一次 REST 对账只有在配置、用户和配置规范化全部成功后才提交 ETag。内核应用失败时 Node 保留失败快照和失败状态，不会用旧配置继续运行；相同快照不自动重试，面板下发新配置或新用户状态后才重新应用。
- Node 的 `alive` 和 `online` 都是权威全量快照。空对象表示没有在线设备或用户；YZboard `1.7.0` 会据此清理旧缓存并把在线人数写为 0。
- 设备快照同时通过周期 HTTP 报告和 WebSocket 上报，两条路径读取同一份不可变快照，不再互相消耗。sing-box 的跨节点设备状态按 2 分钟判断过期。
- 流量报告继续复用 `report_id`。YZboard `1.7.0` 会先持久化报告并在单个数据库事务中结算；升级面板时必须执行新增迁移，否则 Node 会持续保留并重试未被接受的批次。

## 构建与版本检查

发布构建示例：

```bash
VERSION=v1.13-yz.11 make build-linux
```

两个二进制的 `-v`/`version` 输出都包含：

- Node 自身版本、构建时间和提交短 SHA；
- Xray 上游 Tag/commit、YZ fork 版本/commit，以及实际模块替换版本；
- sing-box 请求版本和实际 replacement 版本。

`v1.13-yz.15` Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `9f10f11ed8d449b63893cdd7cdb150c7d3fa37239a12968b5e6f71c1c7d7e41d` |
| `xboard-node-linux-arm64` | `f1719755cab857bbbb4961adb1c17e53e76bc6cd428b2e029898be8a209d273c` |
| `xbctl-linux-amd64` | `e10813146d9a643928d039ce2bff222e521cb5b5b0d8205d1684584da2e9aae9` |
| `xbctl-linux-arm64` | `df8b5b69bf58b05acb80a61df40f97f437eae48d25d21982e9bef90a2e4a0531` |
| `install.sh` | `9b685f508ad44ca179914175595fd3bd1c32f1c40ba617177e7807e41317e7ba` |
| `SHA256SUMS` | `838546fd99216921fdbdabdc8bbb1dfba96d5b603bd0c268eb04ba136b3cc61d` |

`v1.13-yz.11` Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `a6228fdd6e41f3753934635165a221405ad841cfabf1c3d6558f120df131b92b` |
| `xboard-node-linux-arm64` | `4874ba28d26cf5f12a0c18cbbab02f440b32218ff173c5a0ec8696fa4bb7e6bf` |
| `xbctl-linux-amd64` | `a1aa15df6f2f23692227d09e2b2bd17665feea0147f7f6157a83477422bb5fb0` |
| `xbctl-linux-arm64` | `cb506202d724a55929e7b9ecbbf31e0ad29e008649c0850d158bb08823e1de08` |
| `install.sh` | `9b685f508ad44ca179914175595fd3bd1c32f1c40ba617177e7807e41317e7ba` |
| `SHA256SUMS` | `9a657fd90efb1d0ab4122d1562902f2aa0a99e57138d86c25630d35823b38170` |

`v1.13-yz.10` Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `ef103c4de2ec4d5e50785491897ccbf0b6c77be5b85c9011f8703aa2d9df333d` |
| `xboard-node-linux-arm64` | `7c0bb626d775eac127ca5e0fce8a7d7381417df61af6fb0471cc2b60a1f54a36` |
| `xbctl-linux-amd64` | `368ce32546c3e4cd431bf788744cb1ebbf997f57f4123e63671aaf3c5a51a14d` |
| `xbctl-linux-arm64` | `cdcbc9a3c811592c546c2761c07bff0e600d0304aec4829c296fdb2858b3a54f` |
| `install.sh` | `d9e6df2cf7b1cd0441c1d2a74d55a2149ed2f25a18120c530fbb650f89bab431` |
| `SHA256SUMS` | `e3c87d67623b787f6f08ff0372d6aa1201cab3954ee60c1cd1dfee2b20c24bdc` |

`v1.13-yz.9` Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `1f2d6c170aed2479ac365089bc185a1d6baa7714a053f701cbb074e578b14340` |
| `xboard-node-linux-arm64` | `95d0b2ce6810ba326ad1e1b7860a8e9a479039431eee84d29b033e2e91501d01` |
| `xbctl-linux-amd64` | `b6f10695cf20c1db407025233853d41da692c3c896c407aac723a11517dad9d4` |
| `xbctl-linux-arm64` | `47fb158e462c5289bdd02f49ac01248f3bea45c30cca02acf1497c1587ac121f` |
| `install.sh` | `d9e6df2cf7b1cd0441c1d2a74d55a2149ed2f25a18120c530fbb650f89bab431` |
| `SHA256SUMS` | `ce25e451979d2275ed9a9ccc10e15613f6b380afbf8674179dcbd8a7552bc770` |

`v1.13-yz.8` Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `9fef57ea43c0fefc9df516863d6753d7435a4f28a33dc3e9eec0d6c5d5d091f0` |
| `xboard-node-linux-arm64` | `65edb812c6893ca3953927f94635c0d865b57298a9f0405ed93563b26d94eed8` |
| `xbctl-linux-amd64` | `a38964e43a5a6ad76d160f6de955fc9d641e3fef48821c92fa0f53decbe5f705` |
| `xbctl-linux-arm64` | `7986dc8386bb82b49b359afbc5d6a99cf8b5631d6eda9a8531e9f757e16470f0` |
| `install.sh` | `d9e6df2cf7b1cd0441c1d2a74d55a2149ed2f25a18120c530fbb650f89bab431` |
| `SHA256SUMS` | `cd57fa28b929b5512f12cfba35e4eae01348c604ca170d13fd3c4deacddf0194` |

`v1.13-yz.7` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `df82755f05292e989a47fa6b7047586f2ae96c00dec8275c13b45e404f3a1a63` |
| `xboard-node-linux-arm64` | `b40510fe856c6c998ccf0a904e59cbda5093550e21c7874c9d324eed80662822` |
| `xbctl-linux-amd64` | `b127af9b59ed2ae5ee6401398158eb0d930f71f9642066dd18fa8eaa8eb05392` |
| `xbctl-linux-arm64` | `b9509d4be3e84d700f17416c13ca23a33c0d0aaa443481673d96abf07bc0d911` |
| `install.sh` | `d9e6df2cf7b1cd0441c1d2a74d55a2149ed2f25a18120c530fbb650f89bab431` |
| `SHA256SUMS` | `cabf9ed0b3af7bd6e5a12f3bcaec67d493dd63329942c812ff302995ed144c23` |

`v1.13-yz.6` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `97c1603121ed6564b45432098eb2d3470be4f330e236b377aa81f039e6e1aae0` |
| `xboard-node-linux-arm64` | `021c0eb0e1ce4fc2b39be18cb397c35c288bf0657097aefe1bc87c28a4efb84d` |
| `xbctl-linux-amd64` | `18bb034e8c880c3aacc5b214ba5e055b45953c7d073c5e87778b7ee88e0a48d6` |
| `xbctl-linux-arm64` | `d4f5a6771a039e5be02174c21addf936bb2a76c00c8ea24636a42d09e684c3a7` |
| `install.sh` | `d9e6df2cf7b1cd0441c1d2a74d55a2149ed2f25a18120c530fbb650f89bab431` |
| `SHA256SUMS` | `c4955fb132756ef9309dc8ba6fd10b3563b35a5255f479d624272eb302d5ce33` |

`v1.13-yz.5` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `336c1efae66987d32be24c59033abc45f1a0e444679abdcf2948c17ea819495f` |
| `xboard-node-linux-arm64` | `d68001dae1eefbdbc00a99e435debf3316570753a334873b81fe8918b157a766` |
| `xbctl-linux-amd64` | `202eadca74a18995189f9c33e51feddeede921ba985c9eecdafb252591e8ebaa` |
| `xbctl-linux-arm64` | `b05e4e90e9a46dfcbcf4f8386b85e4ba9354e9928bdcc92d6b427330890bbe2e` |
| `install.sh` | `32b0317588421622f4ea24d97ab8a5b813a1c767c0c0e43d9e20fb5f8f977f8e` |
| `SHA256SUMS` | `cab369760d4d299b6a5570fbfb5361c941af40e0b95ad45b58e45b2aa4c81bb2` |

`v1.13-yz.4` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `ba23e86f7e7e331d6d5e21d59c59b3eebd1e521ea82eff2832a5faf74c990a8a` |
| `xboard-node-linux-arm64` | `c2d4db57ae6171d58c7fa1b36f2af9458202333b73c6809a0d9d405735cea86d` |
| `xbctl-linux-amd64` | `5daac44d10a074ab3b12237dc37d2b2a9b5d86f578cfe317f171f86e583c639d` |
| `xbctl-linux-arm64` | `72b84b04961c7344aca16c549f61a8cfe652df44871e29a71fcabb71fec75243` |
| `install.sh` | `32b0317588421622f4ea24d97ab8a5b813a1c767c0c0e43d9e20fb5f8f977f8e` |
| `SHA256SUMS` | `a96778a437a3d20de673c84e63b6e4ad84df3ef610f8f6ed684a5eacd0c36eb0` |

`v1.13-yz.3` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `898bfa76a81bfb71f01a5964d8bef8a8032874507086ccbf00e60975a0a0715e` |
| `xboard-node-linux-arm64` | `ccc53f19466e2c9fcf8afeb8ebe3a2ffc70bff199d928243bc2b078df677fde4` |
| `xbctl-linux-amd64` | `e017b653baf8819ab9cdec416daa04c9a35e9ae03a35f36468c019ec524f02bf` |
| `xbctl-linux-arm64` | `c1b62e0846d49fe0e20527127f772aedb73213990e6352748a2ed53476b84eb5` |
| `install.sh` | `32b0317588421622f4ea24d97ab8a5b813a1c767c0c0e43d9e20fb5f8f977f8e` |
| `SHA256SUMS` | `e1082c8c53d4111709683217d187cb6186e04fc900cfa4da7fc972ecba4d33e1` |

`v1.13-yz.2` 历史 Release 资产校验值：

| 资产 | SHA-256 |
| --- | --- |
| `xboard-node-linux-amd64` | `26ceefd8d190abf46eae64c254fb7a8cda5737f46cabe2980b53c62812aed7ca` |
| `xboard-node-linux-arm64` | `9f4b9b5e5178f36c9708f35399e53857a69a1f2a5908d4184775cbd56a306c88` |
| `xbctl-linux-amd64` | `13b477631bae112134588a184422cb91345bda168fee97987c44a542ece2298f` |
| `xbctl-linux-arm64` | `e0c5eb94288a3d2c4d813fa1bc227d76000e20dae74aac6c6324c0b9ce44f19a` |

发布前至少执行：

```bash
go list -m -json github.com/xtls/xray-core
go test -v -race -count=1 ./...
go build -ldflags "-X main.version=v1.13-yz.11" ./cmd/xboard-node
go build -ldflags "-X main.version=v1.13-yz.11" ./cmd/xbctl
```

安装器和升级器从同一 Node Release 下载 `xboard-node` 和 `xbctl`，并使用该 Release 的 `SHA256SUMS` 校验。面板通过 `releases/latest/download/install.sh` 获取最新正式安装器，安装器再通过 `latest` 解析同一正式 Release；需要回滚时必须传入明确的旧 Node Tag。`.github/workflows/ci.yml` 对固定 `v*` Tag 执行测试、双架构构建、Release 资产上传和多架构镜像发布。只有 Release 记录与六个资产完整、校验值一致且 Docker manifest 包含 `linux/amd64` 和 `linux/arm64` 后，才能把 Tag 视为已发布。Xray fork 的回滚边界由 Node `go.mod` 中记录的 pseudo-version 和对应 fork commit 确定。

## 后续上游同步

同步新的 Xray 预发布 Tag 时：

1. 先记录官方 Tag 和对应 commit，再合并到 YZ fork；
2. 解决冲突时保留 Hysteria2 用户识别、统计、Dispatcher 和限速补丁及其测试；
3. 新上游版本的 fork patch 序列从 `yz.1` 重新开始；
4. 同步更新本文档、`go.mod/go.sum`、Node Release Tag、构建信息和变更说明。
