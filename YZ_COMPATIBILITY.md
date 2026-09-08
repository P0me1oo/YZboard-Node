# YZboard-Node 兼容矩阵

本文件记录可发布的 Node 构建与内嵌内核之间的固定关系。构建上线时必须使用明确的 Node Release Tag 和固定的 Xray fork commit，不能依赖 `main` 或其他移动分支。

## 自定义程序目录（v1.13-yz.23）

| 项目 | 标识 |
| --- | --- |
| 当前 Node 版本 | `v1.13-yz.23`；Tag、Release、双架构安装包及镜像已发布并校验 |
| 正式来源 commit | `1944c8eb7982c4f6156d6adff8e8a734cdc1813b` |
| 本次修改基线 | `1d74d19cbb738e6fcde466361c24a186b8deec60` |
| 修改范围 | 安装器、xbctl 的程序目录与升级回滚、systemd/OpenRC 服务路径及相关测试 |
| 配置与主控兼容 | 保留 `/etc/xboard-node` 和原日志位置；主控通信、配置格式及两个内核的依赖均未变更 |
| 旧安装 | 无目录记录时使用 `/usr/local/bin`；可使用新安装器的 `upgrade --bin-dir` 迁移 |
| 旧版本回退 | 自定义目录下拒绝缺少目录支持的旧 xbctl；先用新安装器迁回默认目录，再回退到 `v1.13-yz.22` 等旧版本 |
| 验证范围 | 本地和 Linux CI 的安装、迁移、错误与中断恢复；正式产物验证见下文，过程与限制见 [自定义安装目录说明](docs/custom-install-directory.md) |

## HY2 ECH 前置入口（v1.13-yz.22）

| 项目 | 标识 |
| --- | --- |
| 该版 Node 版本 | `v1.13-yz.22`；Tag、Release、双架构安装包及镜像已发布 |
| 本次 Node 修改基线 | `0066db507d5fe26698528175d69e30522fa2f4ce` |
| 配套面板版本 | `1.12.0`，正式来源 `c2d6873ec055dbb8d48184eb296d50db6c85e529` |
| Xray 固定依赖 | `v26.7.11-yz.6` / `b4caa82d6414196565599c19ebc1b53e331349b6`，本次没有修改核心 |
| sing-box 固定依赖 | `v1.14.0-yz.2` / `09615a105e219076330d9d2a25ea1e2e733d5427`，本次没有修改依赖 |
| 修改范围与验证 | HY2 ECH 密钥写入、运行配置与中转校验、实际握手及中转；见 [验证记录](docs/hy2-ech-validation.md) |

以下记录 `v1.13-yz.23` 的正式发布结果；Node 回滚基线为 `v1.13-yz.22`，面板沿用 `1.12.0`。`yz.22` 的验证保留在历史发布记录中。

## 最近正式发布引用与回滚基线

| 项目 | 标识 |
| --- | --- |
| Node 本版版本 | `v1.13-yz.23`；固定来源使用同名 Git Tag |
| Node 本版来源 commit | `1944c8eb7982c4f6156d6adff8e8a734cdc1813b` |
| 上一正式 Node 版本（回滚） | `v1.13-yz.22`；自定义目录安装需先用新版安装器迁回 `/usr/local/bin` |
| Node 适用分支 | `upgrade/singbox-v1.14.0` |
| Node 上游发布基线 | `v1.13` |
| Node 上游基线 commit | `0a29338e1f102a462363ce3527417029f89bab28` |
| Node 本版 Release | [v1.13-yz.23](https://github.com/P0me1oo/YZboard-Node/releases/tag/v1.13-yz.23)；完整来源、架构和校验值由附件中的构建信息与 SHA256SUMS 固定 |
| Node 回滚 Tag / commit | `v1.13-yz.22` / `2aa021b65481a14b1d34ff9f594939387c5f0f05` |
| Node `yz.19` 修复基线 | `89c2753390356f51df3d8fc133ae8064fa8ed669` |
| Node `yz.17` 验证构建 commit | `ada7bb60b18bf14b80e171030b82bc0f3412beb6` |
| Node Release 构建工具链 | `Go 1.26.4`（`go.mod` 要求 `go 1.26`） |
| Node Release 构建 | Node、xbctl 的 `linux/amd64` 与 `linux/arm64` 安装包、安装器、四份构建元数据和 SHA256SUMS，共 10 个附件 |
| Node 本版 Docker 标签 | `ghcr.io/p0me1oo/yzboard-node:v1.13-yz.23`；完整提交标签和 `latest` 已同步并核对同一 digest |
| Node 本版 Docker manifest | OCI index `sha256:e041bab08ea982bd205cbe72e93509175de79cbb2c827f73e80d7c7ea76810c2`；包含 `linux/amd64` 与 `linux/arm64` |
| Node 回滚 Docker manifest（yz.22） | OCI index `sha256:8c831eca80ebc66680055e0f6f443a2c0e1830a28fc6bab638fc7a0487a54235`；包含 `linux/amd64` 与 `linux/arm64` |
| 历史 Docker manifest（yz.16 记录） | OCI index `sha256:f3e0895ebc04ac603158a5b96413e7695e5c7a1abd596ee864d7d4852b0b4665`；用于历史版本审计 |
| Node 本版 Docker OCI 标识 | 两架构均为 `revision=1944c8eb7982c4f6156d6adff8e8a734cdc1813b`、`version=v1.13-yz.23` |
| YZboard 兼容版本 | `1.12.0`；沿用 `yz.22` 的通信与内核配置，本次安装目录修改不要求更新面板 |
| YZboard 本版来源 commit | `c2d6873ec055dbb8d48184eb296d50db6c85e529`，Tag `v1.12.0` |
| YZboard 本版镜像 | `ghcr.io/p0me1oo/yzboard:1.12.0-c2d6873`；manifest `sha256:39065c1b1fb66537e62f8b18f8c44c21a8ae0e64b3d120944573a17cfd471aa6` |
| 历史 YZboard 兼容代码 | `f91568d72ffb55205cbcd9b15a8476283a017683`（面板 `v1.11.0`） |
| 历史 YZboard 镜像（1.11.0） | `ghcr.io/p0me1oo/yzboard:1.11.0-f91568d`；manifest `sha256:9ec52732a2f93f77e1ae6f34e314cf9399b26a8c4febf4a82db2e32cd7e651b4` |
| Xray 官方仓库 | `XTLS/Xray-core` |
| Xray 上游预发布 Tag | `v26.7.11` |
| Xray 上游 Tag commit | `50231eaff98ccc31b5cbd247a721c16e97fe5ec1` |
| YZ-Xray-core 源码 Tag | `v26.7.11-yz.6` |
| YZ-Xray-core 当前已固定版本 / commit | `v26.7.11-yz.6` / `b4caa82d6414196565599c19ebc1b53e331349b6` |
| Node 当前 Xray replace | `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260907200713-b4caa82d6414` |
| Xray 模块校验值 | `h1:s4BnktK25n8oj8+sQmfD2n86e+JmekassicYU4F+YFs=` |
| YZboard 兼容标识 | `xray-v26.7.11-yz.6`（面板 `1.12.0`） |
| sing-box `require` 版本 | `v1.14.0` |
| sing-box 实际 replacement | `github.com/P0me1oo/YZ-sing-box v1.14.0-yz.2` |
| sing-box 官方基线 | `v1.14.0` / `0b8995879f29a9b98ee027bc17b75e101445b238` |
| sing-box 兼容仓库 | [P0me1oo/YZ-sing-box](https://github.com/P0me1oo/YZ-sing-box)；保留用户、路由热更新和 Mieru |
| sing-box 兼容 Tag / commit | `v1.14.0-yz.2` / `09615a105e219076330d9d2a25ea1e2e733d5427` |
| sing-box 模块校验值 | `h1:KL0agFYXpL1qEUsa+toI8VhlUVuQser2AVYL3unVkKk=` |
| sing-box 验证依赖 | 正式 `go.mod` 使用远程固定 Tag；前期隔离实测使用同一兼容源码的本地 replacement |
| AnyTLS 上游基线 | `anytls/sing-anytls v0.0.11` / `130d2e61b8895727bfed4942c535e91b246a9603` |
| AnyTLS 实际 replacement | `./compat/sing-anytls`，随 Node 固定提交构建；仅修复流关闭状态和回调的并发访问，来源与移除条件见 [补丁说明](compat/sing-anytls/README.yz.md) |
| SS2022 原模块基线 | `sing-shadowsocks v0.2.8` / `e0612494bafdd1429e9632bc52fd278585d28690` |
| SS2022 实际 replacement | `./compat/sing-shadowsocks`，关闭补丁随 Node 固定提交构建，来源与移除条件见 [补丁说明](compat/sing-shadowsocks/README.yz.md) |

Node 自身版本保持独立，不伪装成 Xray 版本。Node 延续上游 `v1.13` 版本线；`yz.5` 支持首版 Shadowsocks 中转，`yz.6` 至 `yz.9` 延续既有安装、出站和用户同步修订，`yz.10` 新增 VLESS 落地、VLESS Encryption 和当前 Xray 传输矩阵，`yz.11` 为安装器与 `xbctl` 增加 Alpine Linux/OpenRC 生命周期支持，`yz.12` 修复机器模式首个用户同步与失败回滚，`yz.13` 增加 REST/WS 双通道对账、ETag 事务回滚和权威设备快照，`yz.14` 增加 SS2022 进程内时间校准、健康状态和主动诊断，`yz.15` 增加机器模式节点级内核选择，`yz.16` 增加用户-落地节点流量归属上报，`yz.19` 明确失败配置停止和健康失败状态，不自动恢复旧配置。Xray 的上游版本、YZ fork patch 版本和 Node 发布版本分别记录，便于升级、回滚和定位构建来源。

先前的 `v0.1.0-yz.1` Tag 保留用于审计，但其版本低于上游 `v1.13`，不作为部署或升级目标，也不创建对应 Release。

`yz.17` 将 sing-box 官方基线更新为 `v1.14.0`，修复 UDP 统计、用户更新并发、稳定认证身份、路由更新回滚和 Mieru 监听生命周期。兼容源码与 Node 的 Xray 依赖独立；本次没有修改 YZboard 或 YZ-Xray-core。

正式 `go.mod` 与 `go.sum` 固定两个主内核的远程版本和校验值；AnyTLS 与 SS2022 兼容源码随 Node 提交固定。服务器安装版本由用户执行升级后改变。

## 本版发布验证（2026-09-09）

Node 的 [发布前 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34269113262/attempts/2) 与
[正式发布 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34270886003) 均从
`1944c8eb7982c4f6156d6adff8e8a734cdc1813b` 执行完整 Linux `make test`。
各完成 690 项 Go 测试及子测试、20 个新增安装器场景和已有服务文件测试，无失败、无跳过、无数据竞争报告。
发布前发现并修复了 Linux Bash 信号退出后重复回滚的问题；另一次既有 HTTPUpgrade 用例的 `EOF`、同提交重跑和本地复验结果见 [过程记录](docs/custom-install-directory.md)。

正式 Release 发布时间为 `2026-09-08T19:57:07Z`，对应新加坡时间 2026-09-09。
10 个附件已下载，逐个核对 GitHub 附件摘要与 `SHA256SUMS`；安装器内容与固定提交一致。
四个二进制的实际构建信息与附件逐项一致，均为 Go 1.26.4、对应 Linux 架构、`CGO_ENABLED=0`、
上表来源和 `vcs.modified=false`。Node 程序实际解析到上表固定内核、模块校验值和本地兼容模块；xbctl 不链接内核。
CI 已运行 amd64 的 Node、xbctl，以及两个架构的 Node 镜像，确认版本与来源；arm64 镜像版本检查通过 QEMU 执行。

| 正式产物 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `849ac862547850481d4d49c1697858088c4d31be2924a33d611e5bf1b4197992` |
| `xboard-node-linux-arm64` | `803cc7886843d60efb48dbc3ecaec7b27aa2bb63dde90b49ca0fa4e2ee1101a6` |
| `xbctl-linux-amd64` | `6f11c35e2bc4141c512640f3eed24f12999fe4a962c686d52eedaa132f675190` |
| `xbctl-linux-arm64` | `c6d00ae4f429718b08656829b34b1e3f001c9d6da51b2e57cd54b29a8735ffee` |
| `install.sh` | `12ff414f6004a5a4b886055904c39cc75ee7c772cc3a6717546c4079824a7ef1` |
| `SHA256SUMS` | `dd4f56e44742445ff5a564d1170b5b3cf9df1dbe4dc62457752eae949b6fdeb9` |
| `xboard-node-linux-amd64.buildinfo.txt` | `3e04725b37960b68eb4a737b7545200cc712fc4f96e237dcab25bfa75dd26d05` |
| `xboard-node-linux-arm64.buildinfo.txt` | `eebc37582244cd595736fd7e1e62970635d92b0929644191d9e7540b2f409e71` |
| `xbctl-linux-amd64.buildinfo.txt` | `3191f6954cee9a529680efb6bafcdd4cc50f7c7105defe21dad08a76394dca68` |
| `xbctl-linux-arm64.buildinfo.txt` | `ddfaf1f40728080b68581bafb392df45e3c4c95b0e750e832a4bf6a8c838ca60` |

版本标签、完整提交标签与 `latest` 均指向上表 OCI index，已通过匿名读取确认；两个架构的 OCI revision 与 version 一致。

| 镜像平台 | 平台 manifest |
| --- | --- |
| `linux/amd64` | `sha256:1e363a6a2e0d6e83eb827bf2dd076625acc50be7669437c9a8c3541c902e2ad2` |
| `linux/arm64` | `sha256:22168dbb80120bb5919c23399624e027ff91bcdf071eca71c77bdeb72014ff2a` |

安装器测试使用隔离目录和模拟服务；本次未进行生产迁移、Linux 实机服务与系统重启挂载验收或 arm64 实际转发测试。
回退至 `v1.13-yz.22` 前须使用新版安装器将自定义目录迁回默认目录，并保证默认分区具备所需空间。
发布结果通过单独的文档提交补充，已发布 Tag 与产物来源保持固定。

## 历史发布验证：v1.13-yz.22（2026-09-08）

Node 的 [正式发布 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34172378951) 从 `2aa021b65481a14b1d34ff9f594939387c5f0f05` 执行
Linux `make test`，主模块和六组兼容模块共 661 项测试及子测试通过，无失败、无跳过、无数据竞争报告。
HY2 ECH 测试及子测试为 21 项，包含正常握手、错误公钥拒绝和八组中转组合；外部 Mihomo 客户端的八组联测在 Windows 本地完成。

两个架构的安装包与镜像构建、来源检查和镜像版本运行均通过。10 个 Release 附件已下载核对
GitHub 附件摘要、SHA256SUMS 和实际模块信息，四个二进制均为 Go 1.26.4、对应 Linux 架构、
`CGO_ENABLED=0`、上述 `yz.22` 来源和 `vcs.modified=false`，运行版本为 `v1.13-yz.22`。
两个 Node 程序实际使用上表固定内核与兼容模块；xbctl 不链接内核。

| 正式产物 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `172eb498e07d421021f9fdc026a4556ff46bb3f86e4722740703ecc53ff2f51a` |
| `xboard-node-linux-arm64` | `151a6da01a1b3faea98d9202a8a8787deccb938fbe3a63834bd8626044298f83` |
| `xbctl-linux-amd64` | `74e273bae6890f07f0ebff28bc812b3b1b708bbf8c0fd1e708b9f12868e95b81` |
| `xbctl-linux-arm64` | `dfabeced808e2dbd954849635235c06803f05b3ffa164ba754d2947f6ff82813` |
| `install.sh` | `19d5556a52da021209f10ad88d06d26e7050747df2ba00a778f08e66a1c2372a` |
| `SHA256SUMS` | `276e4870dea001170f6245317096000d9b278e0c57b7eee2634563b14bc51f39` |

该版发布时，版本标签、完整提交标签与 `latest` 指向 `sha256:8c831eca80ebc66680055e0f6f443a2c0e1830a28fc6bab638fc7a0487a54235`，两个架构的 OCI 来源和版本均一致，支持匿名拉取。
Linux amd64 runner 已执行实际回环转发；arm64 通过 QEMU 运行镜像版本检查，未进行 arm64 实际转发。
本次没有进行真实服务器测试或外部 DNS ECH 配置部署验证。

配套面板 [v1.12.0](https://github.com/P0me1oo/YZboard/releases/tag/v1.12.0) 已发布，面板测试为 60 项、813 个断言；
镜像固定来源、两个架构和 `latest` 已核对。先升级相关 Node，再更新面板和订阅。
回退到 Node `v1.13-yz.21` 或面板 `1.11.0` 前，应先关闭 HY2 入口的 ECH 并刷新订阅。
发布结果以单独文档提交补充，已发布 Tag 与产物来源保持不变。

## 历史发布验证：v1.13-yz.21（2026-09-08）

Node 的 [发布前 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34158600477) 与
[正式发布 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34159504349/attempts/2) 均从 `2f08f4134d352e127828e1e15aeaa4cfd479864c` 执行完整 `make test`，
各完成 637 项测试及子测试，无跳过、无竞争报告。两个架构的安装包与镜像构建、构建来源检查和镜像版本运行均通过。
正式 CI 第一次因回环临时端口被占用而失败；新 runner 使用同一源码、Tag 和全部检查重跑后通过，首次记录保留在验证文档中。

10 个 Release 附件已下载，逐个核对 `SHA256SUMS` 与 GitHub 附件摘要。四个二进制的实际构建信息与附件一致，
均为 Go 1.26.4、对应 Linux 架构、`CGO_ENABLED=0`、上述 `yz.21` 来源和 `vcs.modified=false`。
Node 二进制实际使用上表两个远程内核及随源码固定的 AnyTLS、SS2022 兼容模块；xbctl 不链接内核。

| 正式产物 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `5ef02f9b680c847758cf6c820f1780f089013d704602b1ee198733ee8f5d6db0` |
| `xboard-node-linux-arm64` | `2b2d604a509446ac17d6cdd58c27807c864cdcf0e19b976d0e4dd602b4cb960d` |
| `xbctl-linux-amd64` | `acc594ec027646aef34d70437572cacf2e8f86935630c5199a918cac3e50daf9` |
| `xbctl-linux-arm64` | `21a141e023c071b689295ddd3d39082eee658f61004b6cd591053a03af14032e` |
| `install.sh` | `19d5556a52da021209f10ad88d06d26e7050747df2ba00a778f08e66a1c2372a` |

完整附件校验清单见 [SHA256SUMS](https://github.com/P0me1oo/YZboard-Node/releases/download/v1.13-yz.21/SHA256SUMS)。
该版发布时，Node 三个镜像标签均解析到 `sha256:500bd8ac445a38ae76550bc2d66c7fd9700515255ab92e9276020bc6136984a0`；面板的不可变标签、`1.11.0` 与 `latest` 也已核对一致，
面板来源构建见 [run 34161072639](https://github.com/P0me1oo/YZboard/actions/runs/34161072639)。两仓库的两个 Linux 架构及 OCI 来源、版本均已逐项核对。

Xray `v26.7.11-yz.6` 的 [三平台测试](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34158811734)、
[多平台构建](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34158811725) 和
[Windows 7 打包](https://github.com/P0me1oo/YZ-Xray-core/actions/runs/34158811778) 均通过。
核心依赖以源码 Tag 固定并内嵌在 Node 安装包、镜像中；本轮未另行创建独立核心二进制 Release。

发布后的补充记录使用单独的文档提交，已发布 Tag 与产物来源保持不变。生产服务器的安装版本由用户执行更新后改变。

## `yz.21` sing-box 中转与并发修复

本版配套面板为 `1.11.0`。sing-box 支持
VLESS/HY2 入口和 Shadowsocks/VLESS 落地，可与 Xray 混用。
2026-09-08 固定 Xray `v26.7.11-yz.6` 和 sing-box `v1.14.0-yz.2`，更新 `go.mod`、`go.sum` 和构建标识。
用户身份按「用户 × 线路」展开，以原生 `auth_user` 规则选路，再映射回真实用户进行限速、设备限制和计费。

认证、用户映射和路由更新按过渡规则协调；删除或轮换用户后，旧 HY2 会话的新请求不能退回默认出站。
未知线路不通过认证。用户和线路计数在内核重建、停止恢复时继续累计，落地不加载面板用户。
sing-box 的线路总量按实际出站上的用户有效载荷计算，Xray 保持原有内部出站口径。

任一端使用 sing-box 时，VLESS 内部链路支持 RAW/TCP、WebSocket、gRPC、HTTPUpgrade，
不支持 VLESS Encryption、TCP 头部伪装及无效 Vision 组合。使用步骤见 [中转说明](docs-relay.md)，
本轮运行、两机测试、性能测量和构建记录见 [sing-box 中转验证](docs/singbox-relay-validation.md)。

Xray `yz.6` 保留 VLESS 首批缓冲上传计数修复，并同步 UDP 缓存及 HY2 会话的关闭状态。sing-box `yz.2` 同步 gRPC 初始化与关闭，修复 SS2022 多用户兼容副本的关闭逻辑；原 SS2022 模块的同一修复随 Node 源码固定。用户套餐和用户-线路明细继续使用原有计数路径。

原三类修复在 Xray `yz.5` 阶段分别由新增用例复现，修复后四个相关包各连续 10 轮 `-race` 通过，共 180 项测试及子测试执行。该阶段 Windows/amd64 的 Node 完整普通测试通过：17 个有测试的包、609 项测试及子测试，无失败或测试跳过；面板 57 项测试、549 个断言通过。

Linux/amd64 完整 sing-box 包并发检测耗时 126.042 秒，142 项测试及子测试全部通过，无跳过、无竞争报告；此前失败的混合内核与 gRPC 分支全部保留。YT-HK、DGN-HK 使用修复后依赖完成 60 次转发及生命周期请求，覆盖 TCP/UDP、用户增删、重复同步、重载和停止恢复。原 Xray `yz.4` 下的失败结果保留在验证文档的历史部分。

上述为发布前工作区的本地及双机验证。正式安装包和镜像已从上表固定 Node 提交构建、发布并核验；结果见本版发布验证。升级时先更新相关 Node，再更新面板并启用新拓扑；上一正式版本用于回滚。

发布前完整 Node 并发检测 [run 34155848942](https://github.com/P0me1oo/YZboard-Node/actions/runs/34155848942) 曾在 HY2 会话关闭处发现另一类竞争，该次未进入构建和发布。补充修复固定为 Xray `yz.6`，5 项会话回归连续 10 轮 Linux 并发检测共 50 次通过，无竞争报告；新依赖随后通过了上述完整 Node 发布验收。

## `yz.20` HY2 前置入口基线（并入 yz.21 发布）

本次修改起点为 Node `94e2a76e42c1f059126b588b2d02f49e54fd8246`，配套面板起点为
`eff2fa22531f2e15168d3e7e96d8ab45639b1969`。源码目标为 Node `v1.13-yz.20` 与面板 `1.10.0`，
该轮没有单独创建 `yz.20` Tag、Release 或镜像，相关功能最终随上文 `yz.21` 一同发布。

HY2 前置入口复用 YZ-Xray-core `601226e180d3684a5eabb8bc901c99f499398db1` 的认证 UUID 路由能力，
入口和落地均使用 Xray，内部协议仍为 Shadowsocks/VLESS。该轮固定依赖未变，联调发现 VLESS 出站上传漏计；后续修复与依赖接入见上文 `yz.21` 记录。
同时修正用户热更新成功后未更新配置指纹的问题，避免随后相同配置同步误触发重建、中断 HY2 会话。

`relay` 配置和三类流量报告的结构不变。应先升级 Node，再启用面板中的 HY2 前置入口。
使用方式、证书和混淆要求见 [中转说明](docs-relay.md)。

### 本地验证状态（2026-09-07，Xray yz.3 历史结果）

当时固定核心下，`TestHysteria2RelayRuntime` 的实际 TCP/UDP 路由、Salamander、用户变更、重载、恢复及用户流量断言通过，
但最终 VLESS 落地出站上传计数为零，运行测试因此失败。
根因是 Xray `BufferToBytesWriter` 的普通字节写入绕过计数器，导致 VLESS 首批缓冲写入漏计；用户套餐和用户-落地明细的计数路径不受此问题影响。

用本地临时覆盖补上字节写入计数后，Node 完整 Go 测试及核心 `common/buf` 测试通过。
独立的核心回归用例还覆盖缓冲刷新、部分写入及失败时的实际字节计数，确认多缓冲区写入不会重复累计。
临时覆盖未修改核心仓库和正式依赖，不能作为正式发布验收；需在修复纳入、固定到远程不可变提交后重新验证。

安装器服务文件测试和 AnyTLS 兼容模块测试通过。Windows 环境下 `go test -race` 因未启用 CGO 无法执行；完整 race 检查仍需具备 C 编译器的环境。

使用正式 `go.mod`、关闭临时工作区及覆盖后，`xboard-node` 和 `xbctl` 的 `linux/amd64`、`linux/arm64` 交叉编译通过。
产物已核对目标架构、`CGO_ENABLED=0`、Xray/sing-box 实际 replacement 和 VCS 信息：
`vcs.revision=94e2a76e42c1f059126b588b2d02f49e54fd8246`、`vcs.modified=true`。
这些产物来自未提交工作区，仅用于编译验证，包含的 Xray 仍是上述未修复统计问题的固定依赖；未执行 Linux 运行验收，也未作为 Release 发布。

## `yz.19` 正式发布（2026-09-07）

发布提交为 `d22037477a7e97825990eb35e41d12926c117680`。[发布前 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34066413890) 与 [正式发布 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34066869570) 均通过，两个架构的镜像版本命令与来源提交一致；arm64 镜像运行检查使用 QEMU。

[GitHub Release](https://github.com/P0me1oo/YZboard-Node/releases/tag/v1.13-yz.19) 包含 10 个附件。安装包的 Go 1.26.4、Linux 架构、`CGO_ENABLED=0`、五个 Node 功能标签、实际模块和 `vcs.modified=false` 已核对；发布附件摘要与 SHA256SUMS 一致。完整构建信息和校验值以 Release 附件为准。

| 安装产物 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `4b6334e15854da2b9ddbc5557a52bc43bc12e986b56c6a3d50516dcb760dd078` |
| `xboard-node-linux-arm64` | `f755ee4c76dd30fd57da40a0f3753373e4eef609d682387ab7327fb64c2f06f7` |
| `xbctl-linux-amd64` | `b3a4e0805ffbb6edd4262b576129083d39971a3707fc4c911839c797e71db1b6` |
| `xbctl-linux-arm64` | `9ab77eff9cc7f6c2bfb9139c07599d9211038a35a3d85aeed392722f3609193c` |
| `install.sh` | `19d5556a52da021209f10ad88d06d26e7050747df2ba00a778f08e66a1c2372a` |

固定镜像引用为 `ghcr.io/p0me1oo/yzboard-node@sha256:8b65c52c0c0f59a24c56ab48ced7a1dda9a07c6948edc3f454b41c140d999818`。`v1.13-yz.19` 与 `latest` 均指向该 index；amd64 manifest 为 `sha256:826ac2d52b00bef1510080bb79f9b76dc12bd8844ad3856f38cf3a5fa5c0cb54`，arm64 manifest 为 `sha256:d93cee9e308cfa77c448031f992ade9e45042647b0edc313ee6041dffc09f064`。两份镜像配置中的 OCI revision 和 version 均与本次发布一致。

上一正式 Node 版本为 `v1.13-yz.16`。本次未执行生产服务器升级；升级时同步服务停止等待时间为 150 秒，具体要求见下文。

## `yz.19` 失败状态修复（2026-09-07）

配置、协议、端口、出站或用户应用失败时，Node 直接记录包含操作、内核和底层原因的错误，停止当前内核并将节点健康状态标记为失败。面板最新的失败配置和用户快照会保留为待修正状态；相同失败快照不会被定时采样、REST 或 WebSocket 反复启动，只有配置或用户实际变化后才允许重新尝试。首次启动失败的进程继续保持控制通道和健康端点，等待修正后的面板配置。

本次只修改 YZboard-Node，未修改 YZboard 或 YZ-Xray-core。详细行为和测试命令见 [Node 修复验证](docs/node-reliability-validation.md)。

配置校验在节点服务内统一执行，REST、WebSocket 和首次同步均保留无效快照并停止对应内核；用户增删不能绕过配置或证书错误。删除问题用户会尝试启动剩余用户，单个节点或实例初始化失败不会取消同一进程的其他节点。整体健康端点返回 503 表示存在失败项，其他节点仍可继续转发。

首次发布前 CI 在 AnyTLS 用户删除测试中捕获上游 `dieErr`、`dieHook` 数据竞争，因此未发布该构建。Node 内增加固定上游源码的最小兼容补丁，保留完整生命周期测试，并单独执行依赖包的并发测试。两个主内核的固定版本未变；AnyTLS 本地 replacement 的身份由 Node 提交及原始文件校验清单共同记录。

`yz.19` 最终开发验证已完成：Windows amd64 通过 18 个包、586 项测试及子测试；YT-HK Linux amd64 `-race` 通过 5 个包、331 项测试及子测试，失败、跳过和数据竞争均为 0。AnyTLS 底层关闭回归重复 50 轮、真实用户生命周期重复 20 轮全部通过。12 个实际进程场景通过，包括两种多节点模式下的端口冲突、无效出站与初始 HTTP 失败隔离，以及等待 18 秒的退出报告和失败重试。Node、xbctl 双架构构建和 amd64 运行时版本检查通过；arm64 未进行实机运行。验收包为 `runtime-validation.tar.xz`，大小 130250888 字节，SHA256 `aa3f59474867e7c727beb793af1186044b44ab97dee2298da40f49bcbd67f424`。原有 7 个监听未变化，临时目录、测试进程和上传包已清理，Netcatty 会话已关闭。

进程退出等待两分钟，安装器的 systemd/OpenRC 模板等待 150 秒。已有部署通过 xbctl 单独替换二进制时还需同步服务停止等待设置；Docker/Compose 也应设为 150 秒。发布 CI 核对完整来源提交、干净源码标识与双架构元数据，Docker 版本检查通过后才更新正式标签。

## `yz.18` 可靠性修复（2026-09-06）

该轮只修改 Node，沿用当时固定的 sing-box、Xray 依赖和既有面板接口。修复范围包括内核监听退役、出站应用与失败恢复、空用户同步、REST 用户重试、跨实例流量累计和退出上报。详细不变量、执行命令和验证状态见 [修复验证](docs/node-reliability-validation.md)。

`yz.18` 开发构建来自当时未提交的修复，保留实际的 `vcs.modified=true` 标识。该轮正式发布前需要提交、固定 Tag、重新构建并更新记录；下文 `yz.17` 的历史校验值不能用于 `yz.18` 修复产物。

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

该轮结果对应 `v1.13-yz.17-test`。Node 使用升级分支的未提交改动及当时的本地兼容核心；这些开发结果不代表正式发布。测试方法和未覆盖的场景见 [sing-box 升级验证](docs/singbox-v1.14-validation.md)。

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

- Xray 的 Hysteria2 用户转换使用 v26.7.11 的 `hysteria/account.MemoryAccount{Auth: ...}`，同时保留 `MemoryUser.Email` 的 `user@<id>` 映射。
- Xray fork 提供的 Dispatcher、用户级限速、统计计数器和在线 IP/连接状态能力继续由 Node 使用。
- Node 的流量方向保持 `[upload, download]`，由内核累计计数器交给 tracker 计算增量，再由面板客户端上报。
- 中转入口的 `relay_user_traffic` 形状为 `user_id => logical_node_id => [upload, download]`，只用于用户-落地归属分析，不参与套餐扣除；`relay_traffic` 继续负责落地节点总量。
- 每次刷出的报告批次带有进程启动标识和递增序号组成的 `report_id`；HTTP 失败时保留完整批次并复用 ID，避免面板重复累计。
- Xray REALITY 入站的 `realitySettings.minClientVer` 由 Node 显式写入，默认 `0.0.0`，可通过 `kernel.reality_min_client_ver` 覆盖。缺省该字段时 v26.7.11 会使用内置下限 `26.3.27`，低于该版本的客户端握手会被拒绝。
- v26.7.11 已移除未加密 Shadowsocks。历史配置中的 `none`/`plain` 会显式返回错误，不会静默转换成其他加密算法。
- `go.mod` 的 Xray `require` 版本只用于保持模块路径兼容；实际代码由 `replace` 固定到上表中的 fork pseudo-version。提交前应使用 `go list -m -json github.com/xtls/xray-core` 复核替换路径和版本。
- Xray 中转通过携带线路编号的认证身份和 `vlessRoute` 规则选路；sing-box 中转按用户与线路生成认证身份，并使用原生 `auth_user` 规则。Node `yz.21` 起入口与落地可混用两种内核，内部传输按两端共同能力校验。
- 面板 `relay` 段与 `relay_traffic` 上报字段属于 YZboard `1.1.0` 起的接口；旧面板不下发该字段时 Node 行为不变。
- 安装器从 `yz.6` 起默认写入 `kernel.type: xray`；`yz.15` 起代码层缺省也按 Xray 处理空值。机器模式下节点的面板 `kernel_type` 优先于机器级默认值；独立实例仍可显式执行 `xbctl config kernel <xray|singbox>` 切换。
- xray 可承载的入站协议为 vmess、vless、trojan、shadowsocks、hysteria；tuic、naive、anytls、mieru、socks、http 只能由 sing-box 承载。安装器和 `xbctl config kernel` 都会在未显式确认时拒绝把这些节点切到 xray。
- 自定义出站从 `yz.7` 起接受 `direct`/`freedom` 与 `block`/`blackhole`，由 Node 翻译成目标内核的原生名；`settings.send_through` 在 xray 下提升为 outbound 级的 `sendThrough`。`settings` 内其余字段原样透传，需按目标内核的字段名填写，跨内核切换时要同步调整。
- 从 `yz.8` 起，同一用户 ID 的 UUID 变化会被视为凭据替换，Xray `UserManager` 必须先删除旧凭据再添加新凭据；任一步失败都不得推进 Node 内部用户状态，并由 Service 尝试使用完整用户集重建内核。
- 从 `yz.9` 起，Xray 的 Shadowsocks 2022 动态用户密钥按面板约定从 UUID 前 16 或 32 字节生成标准 Base64；静态启动配置和运行时增删用户必须得到同一密钥。
- 从 `yz.10` 起，中转 child/landing 同时接受 Shadowsocks 和 VLESS。VLESS 的入口客户端参数放在 `relay.children[].vless`，落地内部身份放在 `relay.vless`；服务端顶层继续承载 `decryption`、Reality 私钥和证书配置。
- 两端均为 Xray 时，VLESS relay 支持 RAW/TCP、WS、gRPC、XHTTP、HTTPUpgrade、mKCP、Hysteria；Reality 只允许 RAW/TCP、gRPC、XHTTP，Hysteria 必须使用 TLS，H2/HTTP 和 mKCP header/seed 会在启动前拒绝。任一端使用 sing-box 时，内部传输限于 RAW/TCP、WS、gRPC、HTTPUpgrade，并按传输校验安全组合。
- `yz.10` 当时沿用已固定的 YZ-Xray-core pseudo-version，本项协议扩展未新增核心补丁。入口和落地 JSON 由该核心自带解析器覆盖验证。
- `yz.11` 的安装器自动识别正在运行的 systemd 或 OpenRC。OpenRC 路径固定使用 `/etc/init.d/xboard-node`、`supervise-daemon` 和 `default` runlevel，日志写入 `/var/log/xboard-node.log`；凭据仍保存在权限为 `0600` 的 `/etc/xboard-node/credentials.env`，启动脚本只按 `KEY=VALUE` 解析，不执行其中内容。
- `yz.14` 的 SS2022 时间校准只在 Node 进程内提供可选时间函数，不修改系统时间。Xray 需要 `v26.7.11-yz.2` 的上下文时间服务补丁；sing-box 使用相同服务，避免每个实例重复查询 NTP。
- `yz.15` 的机器模式按节点创建独立内核服务；只有发现到 sing-box 节点时才创建 sing-box 服务。节点内核变化只重启目标节点，Xray-only 机器不会启动空的 sing-box。

## `yz.14` 时间校准兼容约束

- 普通 SS2022 入站、VLESS 前置中的 SS2022 出站和落地 SS2022 入站共享同一校准结果。VLESS 客户端入口本身不依赖该时间戳。
- 校准器默认并行查询三个 NTP 源并使用有效偏移中位数；查询失败不会猜测时间，最近成功结果超过三个查询周期后回退系统时间。
- `/healthz` 的时钟降级保持 HTTP 200；只有节点组件启动中或失败继续返回 HTTP 503。`xbctl doctor time` 的异常状态返回非零退出码。
- 当前 `go.mod` 已固定 YZ-Xray-core `v0.0.0-20260907151131-4c8f533bce32`，对应 fork `v26.7.11-yz.4` 和 commit `4c8f533bce3258e1c03469d5a6013152988aac04`。正式构建前仍需确认该核心提交可回滚，并不得改回本地路径 replace 或移动分支。

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
