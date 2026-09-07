# sing-box 双内核中转验证

本文对应 Node `v1.13-yz.21`、面板 `1.11.0`；当前修复结果与此前开发记录分别列出。
Node 起点为 `94e2a76e42c1f059126b588b2d02f49e54fd8246`，面板起点为
`eff2fa22531f2e15168d3e7e96d8ab45639b1969`，包含工作区中已有的 HY2 修改。

当前 Xray 固定为 `v26.7.11-yz.6` / `b4caa82d6414196565599c19ebc1b53e331349b6`，sing-box 固定为
`v1.14.0-yz.2` / `09615a105e219076330d9d2a25ea1e2e733d5427`。SS2022 关闭补丁位于 Node 的
`compat/sing-shadowsocks`，上游基线为 `v0.2.8` / `e0612494bafdd1429e9632bc52fd278585d28690`。

## 发布前 HY2 补充修复（2026-09-08）

Node 提交 `9047546147f2028e756a33b3a5fabb8a1a67d778` 的完整 `make test` 在
[run 34155848942](https://github.com/P0me1oo/YZboard-Node/actions/runs/34155848942) 报告另一类数据竞争：
`TestHysteria2RelayRuntime/salamander=false` 中，Xray `udpSessionManager.clean` 在锁外读取关闭状态，
同时 `run` 在锁内写入。流水线在测试阶段失败，未构建或发布安装包、镜像。

同一文件的 `InterConn.Write` 与会话关闭也共享未同步的关闭状态。新增用例在 `yz.5` 上分别复现两处竞争；
`yz.6` 使用管理器读锁与单会话原子状态修复，关闭不等待正在进行的写入。5 项回归在 Linux/amd64、Go 1.26.4 下
连续 10 轮 `-race` 共 50 次通过，耗时 31.045 秒，无竞争报告。覆盖清理退出、空闲会话回收、关闭后拒绝新会话、
并发写入、阻塞写入、读取唤醒、底层错误、缓存排空和重复关闭。

Windows/amd64 的核心 `transport/internet/hysteria`、`common/singbridge` 和 `core` 普通测试通过。
Windows 并发检测器受工具链及运行时地址分配限制未能有效运行，该补充并发结论来自 Linux。
本次 Linux 单元测试没有使用网络监听或认证信息；测试结束后独立目录已删除并复查不存在，本次 Netcatty 会话已关闭。

最终 Xray replacement 为 `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260907200713-b4caa82d6414`，
构建信息和 CI 显式回归包同步更新。最终完整 Node 并发检测、安装包与镜像验证以这个依赖对应的发布 CI 为准。

## 三类并发修复验证（Xray yz.5 阶段，2026-09-08）

本节 Xray 使用 `v26.7.11-yz.5` / `dcb690846b525851f0ee8dc47388e110d4600042`，属于补充 HY2 修复之前的阶段结果。

| 位置 | 修复与保持的行为 |
| --- | --- |
| Xray UDP 缓冲 | 读取顺序与缓存所有权分别同步；关闭不等待阻塞读取，迟到数据会释放；保留包顺序、目标地址和底层错误 |
| SS2022 关闭 | 原模块及 sing-box 多用户副本均直接关闭底层连接一次，不再读取首次应答正在初始化的写入器 |
| sing-box gRPC | 读取等待初始化发布；初始化与关闭共享状态锁；提前关闭唤醒读取，迟到响应被释放，重复关闭只执行一次 |

新增测试先在旧代码上复现三个竞争位置，再验证修复。Go 1.26.4、Linux/amd64、`GOMAXPROCS=2`，四个相关包各连续 10 轮 `-race` 全部通过，共 180 项测试及子测试执行。

| 验证 | 结果 |
| --- | --- |
| Node 完整普通测试 | Windows/amd64，17 个有测试的包，609 项测试及子测试通过；无失败、无测试跳过 |
| Node 完整 sing-box 包并发检测 | Linux/amd64，126.042 秒，142 项测试及子测试通过；无失败、无跳过、无数据竞争 |
| 面板完整测试 | 57 tests / 549 assertions 通过 |
| 两机转发 | YT-HK 入口与 DGN-HK 落地，3.297 秒，60 次成功转发和生命周期检查 |
| Xray 固定提交 CI | 三平台完整测试、多平台构建及 Windows 7 打包均通过 |

完整 sing-box 包保留此前失败的 `TestSingBoxRelayRuntime/vless/xray` 和
`TestSingBoxRelayVLESSTransports/grpc/singbox`，没有删除、跳过或弱化这些用例。
两机测试覆盖 sing-box VLESS/HY2 入口到两种内核的 Shadowsocks/VLESS 四类落地，以及 Xray VLESS/HY2 入口到 sing-box 两类落地；分别为 19、19、11、11 次成功请求。
内部 Shadowsocks 使用 SS2022 AES-128；包含 TCP/UDP、用户增删、重复同步、重载、停止恢复与流量归属检查，落地用户计数为空。

```sh
go test -json -mod=readonly -p 2 -count=1 \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./... -timeout 10m
GOMAXPROCS=2 ./singbox-race-linux-amd64 -test.v -test.run . -test.count 1 -test.timeout 10m
```

本地和双机验证使用 `GOWORK=off` 与空 `GOFLAGS`，没有本地核心路径替换。sing-box 在创建 Tag 前固定到同一远程提交的
`v1.14.0-yz.1.0.20260907182817-09615a105e21`；正式 `go.mod` 改用已固定的 `v1.14.0-yz.2`，来源提交不变。
Xray replacement 为 `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260907183145-dcb690846b52`。
Node 自有 AnyTLS、SS2022 兼容模块随 Node 固定源码构建，来源与删除条件分别记录在兼容目录中。

本轮两机测试结束后，已确认测试进程退出、测试端口释放，既有 Node 进程及监听保持不变。
两端本轮实例目录、配置、证书、密钥和日志已删除，并通过 SFTP 复查目录不存在；本次创建的两个 Netcatty 会话已关闭。
本地配置传输副本和原始普通测试日志已删除，仅保留无凭据的验证摘要、开发产物、备份及编译缓存。

| 验证产物 | SHA256 |
| --- | --- |
| `singbox-race-linux-amd64` | `47830be842b6edb73004936a2719aa1dbeadfae30c30d7649018e3b75ad9bc99` |
| `relay-probe-linux-amd64` | `f763073c6a464dd5ac211de7c7ec142f6140137f1281a16a8257de9edbe27fdc` |

这些产物来自待发布工作区，正式安装包与镜像由 CI 使用固定 Node 提交重新构建。
Xray 固定提交验证：[三平台测试](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34152008337)、
[多平台构建](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34152008334)、
[Windows 7 打包](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34152008353)。

## Xray yz.4 接入验证（2026-09-08，修复前历史记录）

本节使用 Xray `v26.7.11-yz.4` / `4c8f533bce3258e1c03469d5a6013152988aac04`、sing-box `v1.14.0-yz.1` /
`f47d4d565a4371cf46b6c462612fc085f634a6af`。下文的并发失败属于修复前记录，当前修复结果见上节。

接入前重新核对 Node 分支、起点和未提交文件，并备份两仓库工作区；四文件依赖补丁检查通过后，
更新 `go.mod`、`go.sum`、`internal/buildinfo/buildinfo.go` 及其测试。
未修改已有 HY2、sing-box 中转或面板功能实现，也未修改核心仓库。

`go list -mod=readonly -m -json github.com/xtls/xray-core` 确认实际使用上述最终远程提交。
模块校验值为 `h1:TzlUVJkpfU3uZy3+qJUI/2hq2NF38Mu4RfgYbuf5pBY=`，
go.mod 校验值为 `h1:Jts8yHqPCpvsdL5CW5xMd8H9d2fkg1cILeBNqEwRXNw=`。

Windows/amd64、Go 1.26.4、`GOMAXPROCS=2`，构建标识与 Xray 两个包的 129 项测试及子测试通过。
Xray 包耗时 4.692 秒，`TestHysteria2RelayRuntime` 的无混淆和 Salamander 两个子用例均通过，
此前 VLESS 出站上传为零的断言失败已消失。测试断言和运行测试代码未修改。

```sh
go test -json -mod=readonly -p 2 -count=1 \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' \
  ./internal/buildinfo ./internal/kernel/xray -timeout 180s
```

固定依赖下的完整 Node 普通测试通过：17 个有测试的包、609 项测试及子测试，失败和测试跳过均为 0。
另有 4 个包没有测试文件，Go JSON 输出将这些包标为 `skip`，不计入测试跳过数。
其中 sing-box 包耗时 166.169 秒，Xray 包耗时 4.948 秒；中转、认证、用户生命周期、重载恢复、
流量统计、配置校验、控制通道与上报链路均使用当前工作区实现。

```sh
go test -json -mod=readonly -p 2 -count=1 \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./... -timeout 10m
go test -mod=readonly -p 2 -count=1 github.com/xtls/xray-core/common/buf -timeout 120s
```

核心 `common/buf` 回归通过，耗时 0.362 秒，覆盖实际写入计数、部分写入、失败和重复刷新。
面板完整测试再次通过：57 项测试、549 个断言。上述结果均为普通测试，并发检测单独记录。

### 两机转发复测

在用户已授权的 YT-HK 与 DGN-HK 上使用新目录、新认证值及新证书，重新运行当前工作区构建的探针。
YT-HK 运行回环客户端与独立入口，DGN-HK 运行 sing-box/Xray 的 Shadowsocks/VLESS 四个独立落地。
探针 SHA256 为 `92619e8284a1f8fb0fa13ab513218ab952129633ce967b58be5de12a867d7e13`，两端均与本地一致。

| 入口 | 落地范围 | 成功转发次数 |
| --- | --- | --- |
| sing-box VLESS | 两种内核 × Shadowsocks/VLESS | 19 |
| sing-box HY2 | 两种内核 × Shadowsocks/VLESS | 19 |
| Xray VLESS | sing-box Shadowsocks/VLESS | 11 |
| Xray HY2 | sing-box Shadowsocks/VLESS | 11 |

共 60 次成功转发，探针运行耗时 3.166 秒，覆盖 TCP/UDP、用户新增与删除、重复同步、重载及停止恢复。
用户与用户-线路明细按有效载荷匹配；sing-box 线路总量按有效载荷匹配，Xray 内部出站检查双向非零。
内部 VLESS 使用 TLS；无 TLS 场景的上传漏计修复由本轮本地 `TestHysteria2RelayRuntime` 验证。

落地收到停止标记后正常退出，确认落地用户计数为空，没有重复扣费。两台机器上的本次进程、监听、
实例目录、配置、证书、密钥和日志均已清理，两个 Netcatty 会话已关闭；本地配置传输副本和原始测试日志也已删除。
YT-HK 既有 Node 的进程号 `175514`、运行状态及原有监听保持不变，DGN-HK 原有监听也保持不变。
本地保留工作区备份、验证摘要、构建信息、无凭据的开发产物、测试工具和无敏感编译缓存。

### Linux 并发复查

使用同一工作区和固定依赖重新编译完整 sing-box 包的 Linux/amd64 `-race` 测试程序，
工具链为 Go 1.26.4、Zig 0.14.1，C 编译目标为 `x86_64-linux-gnu.2.31`。
构建元数据确认 `CGO_ENABLED=1`、`-race=true`、`GOOS=linux`、`GOARCH=amd64`、
`vcs.revision=94e2a76e42c1f059126b588b2d02f49e54fd8246`、`vcs.modified=true`。
测试程序 SHA256 为 `0eac46e330a500d6e5306b3bfe57c2370049cc3710d84bf4b644581f149e69a4`，上传至 YT-HK 后校验一致。

```sh
GOMAXPROCS=2 ./singbox-race-linux-amd64 -test.v -test.run . -test.count 1 -test.timeout 10m
```

完整运行耗时 125.307 秒，退出码为 1；138 项测试及子测试通过，没有跳过，报告 4 次数据竞争。
失败分支为 `TestSingBoxRelayRuntime/vless/xray` 与 `TestSingBoxRelayVLESSTransports/grpc/singbox`，
它们的父测试也随之失败。竞争仍指向以下三类依赖位置：

| 固定依赖 | 竞争位置 | 根因 |
| --- | --- | --- |
| Xray `yz.4` / `4c8f533bce32` | `common/singbridge/packet.go:54,105` | UDP 读取更新缓存，同时关闭连接释放该缓存，缺少同步 |
| sing-shadowsocks `v0.2.8` | `shadowaead_2022/service.go:305,361` | 首次应答发布写入器，同时关闭连接读取写入器，缺少同步 |
| sing-box `v1.14.0-yz.1` | `transport/v2raygrpclite/conn.go:52,64` | gRPC 读取器初始化与读取路径检查并发发生，尚未建立同步关系 |

堆栈中的模块路径使用 `require` 版本；实际 replacement 为本页开头的固定 fork，不能仅按堆栈路径误判为官方旧核心。
`yz.4` 相对 `yz.3` 没有修改上述竞争位置；本次接入修复了上传漏计，完整并发验收仍未通过。
后续需在相应依赖中协调读写、初始化和关闭，再保留这些混合内核及 gRPC 用例复测。

### 当前开发构建

Node 与 xbctl 的 `linux/amd64`、`linux/arm64` 四个开发构建均成功，版本为 `v1.13-yz.21-dev`，
构建时间为 `2026-09-07T16:09:23Z`。构建与测试均使用 `GOWORK=off`、空 `GOFLAGS` 和正式 `go.mod`。
产物元数据确认 Go 1.26.4、`CGO_ENABLED=0`、对应目标架构、
`vcs.revision=94e2a76e42c1f059126b588b2d02f49e54fd8246`、`vcs.modified=true`。
两个 Node 产物实际使用 Xray `v0.0.0-20260907151131-4c8f533bce32` 和 sing-box `v1.14.0-yz.1`；
AnyTLS 保持 `./compat/sing-anytls`。xbctl 自身不链接内核，版本输出中的内核模块为 `not-linked`。

| 开发构建 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `17424f8b1811937ab9a2e256531a6244c39c0eb231dbd34233e5f8fbc68d61e5` |
| `xbctl-linux-amd64` | `0626bc095ad64503a84df8ee52df89f7c7e71658936ce66f5524b3302fec055b` |
| `xboard-node-linux-arm64` | `51ee9a974f9ba215ed39a79e5f7ede02d97c5093425812cb2d5d1075bf635e00` |
| `xbctl-linux-arm64` | `857b2aa9c3d105ac6455d9f239a723ddf9197b3d25664b49ce412ca78b65aec4` |

amd64 两个程序在 YT-HK 独立目录实际执行版本命令，输出与上述源码、版本及依赖一致；没有安装或替换现有服务。
arm64 仅完成交叉编译及元数据核对，未进行实机运行。当前工作区尚未提交、打 Tag 或发布。

## 首轮本地回归（2026-09-07，Xray yz.3）

以下首轮记录使用 Xray `v26.7.11-yz.3` / `601226e180d3684a5eabb8bc901c99f499398db1`，
保留原始失败和构建结果；不能将其替代为接入 `yz.4` 后的验证。

工具链为 Go 1.26.4，进程环境设置 `GOMAXPROCS=2`。运行命令：

```sh
go test -mod=readonly -p 2 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./...
go test -mod=readonly -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./internal/kernel/singbox -run '^Test(SingBoxRelay|Relay)' -count=1
```

2026-09-07，Windows/amd64 下完整 sing-box 包通过，耗时 209.235 秒；Node 其余包也通过，
唯一失败包为 `internal/kernel/xray`，其 `TestHysteria2RelayRuntime` 两个混淆子用例均因
VLESS 出站上传计数为零失败。此问题在本轮开始前已记录，根因是固定 Xray 的
`BufferToBytesWriter` 普通字节写入绕过出站计数。用户套餐和用户-线路明细走不同计数路径。
不能据此把全量 Node 测试记为通过。

中转回归实际启动内核，用各出口不同的应答内容检查选路，覆盖：

- VLESS/HY2 sing-box 入口，分别连接 sing-box/Xray 的 Shadowsocks 与 VLESS 落地；TCP、UDP 都验证。
- 用户新增、重复新增、删除、凭据轮换、清空及恢复。删除和轮换后，新认证与旧 HY2 会话的新请求都必须拒绝。
- 重复配置同步、非法配置、删除及恢复落地、停止后重启，检查用户与线路累计值恰好连续，落地用户流量保持为空。
- 限速与设备查询使用真实 UUID；管理员覆盖出口为直连时，流量不计入原计划的落地。
- RAW/TCP、WebSocket、gRPC、HTTPUpgrade 四种 VLESS 内部传输，分别连接两种落地内核，均完成 TCP/UDP 实际转发。

拒绝用例以“收到任意出口应答即失败”为准，不把应答内容不同当作认证失败；网络监控变化会直接导致用例失败。
HY2 UDP 测试使用 `SetReadDeadline`，因为其底层 `SetDeadline` 不受支持。

配套面板完整测试通过：57 项、549 个断言。

```sh
php -d extension=pdo_sqlite -d extension=sqlite3 vendor/bin/phpunit --no-progress --stop-on-error
```

PHP 的 SQLite 扩展通过当前命令启用，未修改系统 PHP 配置。管理端补丁分别验证原始资源、
旧补丁增量升级与重复应用；最终 JavaScript 语法检查通过，重复执行保持文件校验值不变。

## 首轮两机运行验证（Xray yz.3）

使用用户指定的 Netcatty 主机，测试程序为 `tests/relay-probe`，不安装服务或读取现有 Node 配置。

| 测试机 | 用途 |
| --- | --- |
| YT-HK | 回环客户端与独立入口实例 |
| DGN-HK | sing-box/Xray × Shadowsocks/VLESS，共四个独立落地 |

内部 Shadowsocks 使用 SS2022；内部 VLESS 使用本次生成的证书完成 TLS 验证。
落地仅把保留测试地址的请求转到本机应答程序。证书信任只作用于当前测试进程。

测试程序 SHA256：`b114f965632199b5186aa804da9c8021a9499b56a96c4642b408fc8bdd34b5eb`，
上传前与两台机器上的文件均已核对一致。

sing-box VLESS 与 HY2 入口均通过，分别完成 19 次跨机请求，共 38 次。
覆盖四种落地、TCP/UDP、用户增删、重复更新、配置重载及停止恢复；用户、线路与用户-线路计数均按有效载荷精确匹配。

Xray 的 VLESS 与 HY2 入口反向连接两种 sing-box 落地也通过，各完成 11 次请求。
两机共完成 60 次成功请求。Xray 用户与用户-线路明细按有效载荷精确匹配，内部出站检查双向计数非零；
本次 TLS 链路通过这一检查，不能据此认为本地无 TLS 用例暴露的上传漏计已修复。

本轮实测之后，探针关闭了 Xray 原始访问日志，避免拒绝请求时输出测试认证值；此调整不改变转发与断言。

## 首轮 Linux 数据竞争检测（Xray yz.3）

Windows 使用 Zig 交叉编译带 `-race` 的 Linux/amd64 测试程序，在 YT-HK 的独立目录运行
完整 sing-box 包。测试文件 SHA256 为
`f2acd3c9177a6d9991981e771abd9acdb1d1946de648afe43631344122184bec`，两端一致；
构建元数据确认 `GOOS=linux`、`GOARCH=amd64`、`CGO_ENABLED=1`、`-race=true`。

完整检测失败，报告四次竞争，涉及固定依赖中的三类位置：

| 依赖及位置 | 并发访问 |
| --- | --- |
| Xray `common/singbridge/packet.go:54,105` | UDP 读取更新缓冲状态，同时连接关闭读取该状态 |
| sing-shadowsocks `shadowaead_2022/service.go:305,361` | SS2022 首次应答初始化写入器，同时关闭连接读取写入器 |
| sing-box `transport/v2raygrpclite/conn.go:52,64` | gRPC 客户端设置读取器，同时 TCP/UDP 读取路径访问该读取器 |

这些栈均位于固定内核或其依赖，新增的混合内核与 gRPC 用例使它们在本轮检测中暴露。
不能把这一结果写成“完整 race 通过”。本轮保留固定依赖及失败用例，没有修改核心仓库。

随后仅为定位范围，使用同一测试程序对纯 sing-box 的 VLESS/HY2 中转、身份碰撞与线路计数用例
连续运行三轮，均通过且没有竞争报告。该范围不包括 Xray 落地和 gRPC，不替代上面的完整检测结果。

```sh
./singbox-race -test.run '^Test(SingBoxRelayRuntime|RelayIdentityMappingAndCollision|RelayPacketCountersUseActualOutbound)$/(vless|hysteria2)/singbox$' -test.count=3 -test.v -test.timeout=5m
```

## 两机测试复现

先按任务约定确认两台测试机及四个连续空闲端口，使用新的私有目录与凭据。构建：

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -trimpath \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' \
  -o relay-probe ./tests/relay-probe
```

以下参数中的路径、地址和端口需替换为本次确认的测试范围：

```sh
./relay-probe -mode prepare -config <独立配置路径> -address <落地测试机IP> -base-port <起始端口>
./relay-probe -mode landing -config <独立配置路径>
./relay-probe -mode probe -entry-kernel singbox -config <独立配置路径>
./relay-probe -mode probe -entry-kernel xray -config <独立配置路径>
```

`prepare` 仅允许新建配置，生成随机认证值和有效期两小时的测试证书；两端通过私有文件传递配置，
不打印内容。落地创建 `<配置路径>.ready` 后可运行入口测试；创建 `<配置路径>.stop` 会停止落地，
并检查落地没有重复上报用户流量。落地最长运行二十分钟，超时退出视为测试未正常完成。
每轮复测都使用新目录，不复用已经运行过的实例目录。

测试结束后确认进程和监听退出，删除本轮配置、证书、密钥、实例目录与本地传输副本。
已有服务与目录不在清理范围内。

本轮已完成清理：DGN-HK 落地正常退出且测试端口释放，两台机器的本次目录均已删除，
本地配置传输副本与中断测试遗留目录也已清理。YT-HK 原有 Node 服务状态和进程号保持不变，
两个测试会话已关闭。

## 配置生成成本

```sh
go test -mod=readonly ./internal/kernel/singbox -run '^$' \
  -bench '^BenchmarkSingBoxRelayConfig$' -benchmem -benchtime=5x -count=1
```

Windows/amd64，Intel Core i5-8265U，`GOMAXPROCS=2`；1000 用户、20 个落地加入口自身，
共 21000 个认证项。五次构建的平均结果为：

| 项目 | 结果 |
| --- | --- |
| 单次配置构造耗时 | 94.3743 毫秒 |
| 单次累计内存分配 | 16,430,129 字节 |
| 单次分配次数 | 329,582 |

该测量只包含配置对象构造，不包含 JSON 编码、内核启动、网络收发或完整热更新，
分配字节也不等于进程常驻内存。热更新还会构造过渡规则和解析配置，不能用此数据直接推算线上容量。

## 首轮构建与发布状态（Xray yz.3）

本轮构建用于验证未提交工作区，版本标记为 `v1.13-yz.21-dev`，不是正式 Release。
当时尚未接入 Xray 计数修复；后续依赖接入见本文开头。正式发布前需要固定提交，重新运行验证并记录不可变产物及回滚版本。

Node 与 xbctl 的 `linux/amd64`、`linux/arm64` 四个构建均成功。`go version -m` 确认
Go 1.26.4、目标操作系统和架构、`CGO_ENABLED=0`，以及
`vcs.revision=94e2a76e42c1f059126b588b2d02f49e54fd8246`、`vcs.modified=true`。
两个 Node 产物实际解析到首轮 Xray `v0.0.0-20260903142229-601226e180d3` 与 sing-box `v1.14.0-yz.1`；AnyTLS 继续使用既有的
`./compat/sing-anytls`。未设置额外 Go 工作区或源码覆盖。

| 开发构建 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `58b4a3426afd1c6841eed8ac79176ee2f6dd4dd3e04fe8dd4e692e11f2ca0389` |
| `xbctl-linux-amd64` | `699f48c27f15315a0608c40f2608cb0fd9a45571aef9409fbe5ccefbc5aa1f8c` |
| `xboard-node-linux-arm64` | `793ebfea4012739325e27fc981f7fed21a6b78e7ee98828ac14482269d5c5c34` |
| `xbctl-linux-arm64` | `2bd48694215d6ace0b0b064f5b85c00fac4f9fb9d6f5e41adaea664e45bf652e` |

本轮 Linux 运行测试使用 amd64 测试程序；arm64 只完成交叉编译和元数据校验，没有进行实机运行。
