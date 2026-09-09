# HY2 ECH 中转验证记录

日期：2026-09-08。对应 Node `v1.13-yz.22` 和面板 `1.12.0`。
本页记录本地源码和开发构建验证；正式发布来源、产物和后续 CI 结果以兼容矩阵中的发布记录为准。

## 源码与依赖

Node 修改基线为 `0066db507d5fe26698528175d69e30522fa2f4ce`，面板修改基线为
`f5d5075baf9592080a9bd0a1dba8c492f3684cc3`。两边都在现有分支工作区修改，没有引入新的核心补丁。

实际解析的固定模块保持如下：

- Xray：`github.com/P0me1oo/YZ-Xray-core v0.0.0-20260907200713-b4caa82d6414`，对应 `v26.7.11-yz.6`。
- sing-box：`github.com/P0me1oo/YZ-sing-box v1.14.0-yz.2`，对应 `09615a105e219076330d9d2a25ea1e2e733d5427`。
- AnyTLS 与 SS2022 继续使用随 Node 固定的现有兼容模块。

## 修复与运行验证

Xray 的普通 TLS 入站原本会提取 ECH 密钥，但 HY2 单独生成 TLS 配置时遗漏了该字段。
本次补齐 `echServerKeys`，并在序列化配置前检查实际写入的密钥、Base64 编码和密钥列表。
密钥文件只读取一次，校验和运行配置使用同一份内容，避免文件轮换期间二次读取造成不一致。
中转校验改为要求 `key` 或 `key_path`。sing-box 已有 HY2 ECH 写入路径，继续由其原生 TLS 实现校验和加载密钥。

`internal/kernel/xray/hysteria_ech_test.go` 验证内联 PEM、Base64 和文件来源的真实 QUIC 握手，
读取 `TLS.ECHAccepted` 确认服务端接受 ECH，并验证缺失、不可读、公钥误作私钥及损坏编码被拒绝。

`internal/kernel/singbox/hysteria_ech_test.go` 在回环地址实际启动两种内核，覆盖以下八组：

| 入口 | 落地内核 | 入口混淆 | 结果 |
| --- | --- | --- | --- |
| Xray | Xray | 无／Salamander | 通过 |
| Xray | sing-box | 无／Salamander | 通过 |
| sing-box | Xray | 无／Salamander | 通过 |
| sing-box | sing-box | 无／Salamander | 通过 |

每组分别使用直接出口、SS2022 和 VLESS 落地完成 TCP/UDP 应答，以不同返回标记核对实际选路。
同时检查错误 ECH 公共配置拒绝、用户增删与重复调用、同配置重载、ECH 密钥及端口变更、停止恢复。
用户和用户-线路流量按有效载荷精确累计，落地不再次计入用户套餐流量；落地总量沿用现有核心出站口径。

测试使用正常证书验证，证书、ECH 密钥和身份均在运行时生成。开发时曾因探针向 Windows 的 UDP4
套接字传入 IPv4 映射地址，导致 Salamander 用例报 `An address incompatible with the requested protocol was used`；
探针改为 `To4()` 后，八组专项测试和后续全量测试均通过，产品代码及核心无需地址兼容补丁。

## Mihomo 与面板联测

专项测试可通过 `YZ_HY2_ECH_MIHOMO` 加入官方 Mihomo 客户端。
`internal/kernel/singbox/hysteria_ech_mihomo_test.go` 调用面板实际的 `ClashMeta::buildHysteria`，
只补充本次测试证书的信任信息，再启动回环 SOCKS 客户端。八组组合均验证 SS2022／VLESS 的 TCP/UDP，
错误 ECH 公钥时不能继续得到出口应答。

本次使用官方 `v1.19.9` 的 Windows amd64 compatible 发行文件，程序版本输出为 Go `1.24.3`、`with_gvisor`。

| 文件 | SHA256 |
| --- | --- |
| `mihomo-windows-amd64-compatible-v1.19.9.zip` | `1dd10bddfaf210b3888c766ac1d96e646ec9bae89663243a859c4ae5778ce983` |
| `mihomo-windows-amd64-compatible.exe` | `3c06883f5cec0375d0ee59ff3e73969028cdf0983d128d3501c39f22e68c1c85` |

来源：[官方固定 Release](https://github.com/MetaCubeX/mihomo/releases/tag/v1.19.9)。
参数支持也与该 Tag 的 `adapter/outbound/hysteria2.go` 核对，包括 `ech-opts` 和 `ca-str`。

## 测试命令与结果

本机使用 Go `1.26.4 windows/amd64`。以下普通测试均通过：

```powershell
go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./... -timeout 10m

go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./internal/kernel/xray -run 'TestHysteria2ECH' -timeout 2m

go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./internal/kernel/singbox -run 'TestHysteria2ECHRelayRuntime' -timeout 4m

go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' github.com/anytls/sing-anytls/session github.com/sagernet/sing-shadowsocks/shadowaead_2022 github.com/xtls/xray-core/common/singbridge github.com/xtls/xray-core/transport/internet/hysteria github.com/sagernet/sing-box/transport/v2raygrpclite github.com/sagernet/sing-box/third_party/sing-shadowsocks/shadowaead_2022 -timeout 4m

bash tests/install_service_manager_test.sh
```

Mihomo 增强联测需要本地已安装 PHP 和面板 Composer 依赖。将两个路径设为本次使用的实际目录后执行：

```powershell
$env:YZ_HY2_ECH_MIHOMO = '<固定版本 Mihomo 可执行文件的绝对路径>'
$env:YZ_HY2_ECH_PANEL_ROOT = '<YZboard 仓库的绝对路径>'
go test -mod=readonly -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./internal/kernel/singbox -run 'TestHysteria2ECHRelayRuntime' -timeout 5m
```

本次增强联测通过。未设置 `YZ_HY2_ECH_MIHOMO` 时，专项测试仍执行完整 sing-box 客户端矩阵；
不会把未执行的外部客户端测试计为已验证。

## Linux 开发构建

Node 与 xbctl 的 `linux/amd64`、`linux/arm64` 构建均通过，使用 `CGO_ENABLED=0`、`-mod=readonly`、
`-trimpath`、`-buildvcs=true`，Node 使用上面五个功能标签。运行版本注入 `v1.13-yz.22-dev`。
构建中的完整 VCS revision 为 `0066db507d5fe26698528175d69e30522fa2f4ce`，`vcs.modified=true`，
表示包含本次未提交修改，不能当作该基线 commit 的干净发布产物。两种 Node 产物实际包含的 Xray／sing-box
替换模块与本页固定依赖一致，目标架构均已通过 `go version -m` 核对。

| 开发产物 | SHA256 |
| --- | --- |
| `xboard-node-linux-amd64` | `7fafd72874d52e64caccc0f82644f6ca4cde7ae6a361ed6741312e35edf9e3e1` |
| `xboard-node-linux-arm64` | `b498c8804fc31a0a2b10e5f175c7959f16885d632f7ac08b62f6fe9fdbfb92ae` |
| `xbctl-linux-amd64` | `ccad5fb8c9cb1fe7a8001de988d7dc2e41b2f2d94b2b1ea760e229690ce0bd38` |
| `xbctl-linux-arm64` | `d0039af5fc2bb99456cc0565f29ec2df8c7a146ce3c6a78b8ec135d1ea4bab26` |

## 正式发布验证

Node `v1.13-yz.22` 固定来源为 `2aa021b65481a14b1d34ff9f594939387c5f0f05`。
[正式发布 CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34172378951) 在 Linux runner 上完成
`make test`，包括主模块和六组兼容模块的数据竞争检测，共 661 项测试及子测试通过，失败、跳过和竞争报告均为 0。
其中 HY2 ECH 相关测试及子测试为 21 项，包含实际握手和八组中转组合。
本节的 Linux 结果补充前文 Windows 开发阶段的验证，外部 Mihomo 客户端仍以本地联测结果为准。

[Release](https://github.com/P0me1oo/YZboard-Node/releases/tag/v1.13-yz.22) 的 10 个附件已下载，
逐个核对 GitHub 附件摘要、`SHA256SUMS` 和实际二进制模块信息。四个程序均为 Go `1.26.4`、
对应 Linux 架构、`CGO_ENABLED=0`、上述来源和 `vcs.modified=false`，运行版本为 `v1.13-yz.22`。
两个 Node 程序解析出的 Xray、sing-box 和本地兼容模块与本页固定依赖一致。

镜像版本标签、完整提交标签和 `latest` 已核对为同一 manifest：
`sha256:8c831eca80ebc66680055e0f6f443a2c0e1830a28fc6bab638fc7a0487a54235`。
两架构的 OCI 来源与版本一致，并通过镜像内运行版本检查；arm64 版本检查使用 QEMU。
正式产物校验值与面板配套发布信息见 [兼容矩阵](../YZ_COMPATIBILITY.md)。

## 环境限制与清理

`go test -mod=readonly -race -count=1 -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./internal/kernel/xray ./internal/kernel/singbox -run 'TestHysteria2ECH' -timeout 4m`
在默认环境提示 `-race requires cgo`；本次仅在测试进程设置 `CGO_ENABLED=1` 后，两个内核测试包仍因
`cgo: C compiler "gcc" not found` 无法构建。该限制来自本机工具链；后续正式 Linux CI 已完成数据竞争检测。
Windows 开发阶段未执行 Linux 转发，正式 CI 已在 Linux amd64 runner 上执行实际回环转发。
本次没有进行 arm64 实际转发或真实服务器测试，也没有验证外部 DNS HTTPS 记录提供 ECH 配置的部署。

所有测试实例监听回环地址，临时证书、身份、配置和 ECH 密钥由测试目录自动清理。
外部 Mihomo 进程已停止。下载文件、管理端资源副本和一次性开发构建保留在
仓库内的 `.hy2-ech-validation-20260908064733/`，其中不包含测试证书、身份或 ECH 密钥。
该临时目录不纳入源码提交或正式发布产物。
正式发布必须从已提交的固定源码重新构建并记录不可变引用；上一正式版本为 Node `v1.13-yz.21` 和面板 `1.11.0`。
