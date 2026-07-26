# YZboard-Node 兼容矩阵

本文件记录可发布的 Node 构建与内嵌内核之间的固定关系。构建上线时必须使用明确的 Node Release Tag 和固定的 Xray fork commit，不能依赖 `main` 或其他移动分支。

## 当前构建

| 项目 | 标识 |
| --- | --- |
| Node 发布版本 | `v1.13-yz.7` |
| Node 适用分支 | `upgrade/xray-v26.7.11-yz.1` |
| Node 上游发布基线 | `v1.13` |
| Node 上游基线 commit | `0a29338e1f102a462363ce3527417029f89bab28` |
| Node Release Tag 对应 commit | 发布后回填 |
| Node Release 构建工具链 | `Go 1.26.4`（`go.mod` 要求 `go 1.26`） |
| Node Release 构建 | 发布后回填 |
| Node Docker 标签 | 发布后回填 |
| Node Docker manifest | 发布后回填 |
| Node Docker OCI 标识 | 发布后回填 |
| YZboard 兼容代码 commit | `90c11685eab03a68e167a3c0c969bd774a89e362`（面板版本 `1.2.1`） |
| Xray 官方仓库 | `XTLS/Xray-core` |
| Xray 上游预发布 Tag | `v26.7.11` |
| Xray 上游 Tag commit | `50231eaff98ccc31b5cbd247a721c16e97fe5ec1` |
| YZ-Xray-core fork 版本 | `v26.7.11-yz.1` |
| YZ-Xray-core fork commit | `620bee93867095f73880056cdfb08bc54a15f69e` |
| Node 中的 Xray replace | `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260724203739-620bee938670` |
| YZboard 兼容标识 | `xray-v26.7.11-yz.1`（面板版本 `1.2.1`） |
| sing-box `require` 版本 | `v1.13.2` |
| sing-box 实际 replacement | `github.com/cedar2025/sing-box v1.14.0-alpha.2.0.20260316103356-2e665cb7e295` |

Node 自身版本保持独立，不伪装成 Xray 版本。Node 延续上游 `v1.13` 版本线；`yz.5` 支持面板下发的 `relay` 段实现单入口多落地中转，`yz.6` 把安装器默认内核改为 xray 并新增 `xbctl config kernel` 切换命令，`yz.7` 让自定义出站支持直连、拦截与源地址绑定，均不改变 Xray fork 基线。Xray 的上游版本、YZ fork patch 版本和 Node 发布版本分别记录，便于升级、回滚和定位构建来源。

先前的 `v0.1.0-yz.1` Tag 保留用于审计，但其版本低于上游 `v1.13`，不作为部署或升级目标，也不创建对应 Release。

## 兼容约束

- Hysteria2 用户转换使用 Xray v26.7.11 的 `hysteria/account.MemoryAccount{Auth: ...}`，同时保留 `MemoryUser.Email` 的 `user@<id>` 映射。
- Xray fork 提供的 Dispatcher、用户级限速、统计计数器和在线 IP/连接状态能力继续由 Node 使用。
- Node 的流量方向保持 `[upload, download]`，由内核累计计数器交给 tracker 计算增量，再由面板客户端上报。
- 每次刷出的报告批次带有进程启动标识和递增序号组成的 `report_id`；HTTP 失败时保留完整批次并复用 ID，避免面板重复累计。
- Xray REALITY 入站的 `realitySettings.minClientVer` 由 Node 显式写入，默认 `0.0.0`，可通过 `kernel.reality_min_client_ver` 覆盖。缺省该字段时 v26.7.11 会使用内置下限 `26.3.27`，低于该版本的客户端握手会被拒绝。
- v26.7.11 已移除未加密 Shadowsocks。历史配置中的 `none`/`plain` 会显式返回错误，不会静默转换成其他加密算法。
- `go.mod` 的 Xray `require` 版本只用于保持模块路径兼容；实际代码由 `replace` 固定到上表中的 fork pseudo-version。提交前应使用 `go list -m -json github.com/xtls/xray-core` 复核替换路径和版本。
- 中转拓扑依赖 Xray 的 VLESS 路由值能力：认证前清零 UUID 第 7、8 字节，认证后按原始字节还原，并由路由规则的 `vlessRoute` 匹配。该能力来自上游 `v26.7.11`，sing-box 不具备，因此入口和落地节点都要求 xray 内核。
- 面板 `relay` 段与 `relay_traffic` 上报字段属于 YZboard `1.1.0` 起的接口；旧面板不下发该字段时 Node 行为不变。
- 安装器从 `yz.6` 起默认写入 `kernel.type: xray`。代码层缺省仍为 `singbox`，因此已有配置不会因升级二进制而被动改内核；切换需要显式执行 `xbctl config kernel <xray|singbox>` 或用新安装器覆盖安装。
- xray 可承载的入站协议为 vmess、vless、trojan、shadowsocks、hysteria；tuic、naive、anytls、mieru、socks、http 只能由 sing-box 承载。安装器和 `xbctl config kernel` 都会在未显式确认时拒绝把这些节点切到 xray。
- 自定义出站从 `yz.7` 起接受 `direct`/`freedom` 与 `block`/`blackhole`，由 Node 翻译成目标内核的原生名；`settings.send_through` 在 xray 下提升为 outbound 级的 `sendThrough`。`settings` 内其余字段原样透传，需按目标内核的字段名填写，跨内核切换时要同步调整。

## 构建与版本检查

发布构建示例：

```bash
VERSION=v1.13-yz.7 make build-linux
```

两个二进制的 `-v`/`version` 输出都包含：

- Node 自身版本、构建时间和提交短 SHA；
- Xray 上游 Tag/commit、YZ fork 版本/commit，以及实际模块替换版本；
- sing-box 请求版本和实际 replacement 版本。

`v1.13-yz.7` 尚未构建和发布，Release 资产校验值发布后回填。

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
go build -ldflags "-X main.version=v1.13-yz.7" ./cmd/xboard-node
go build -ldflags "-X main.version=v1.13-yz.7" ./cmd/xbctl
```

安装器和升级器从同一 Node Release 下载 `xboard-node` 和 `xbctl`，并使用该 Release 的 `SHA256SUMS` 校验。面板通过 `releases/latest/download/install.sh` 获取最新正式安装器，安装器再通过 `latest` 解析同一正式 Release；需要回滚时必须传入明确的旧 Node Tag。`.github/workflows/ci.yml` 已配置 `v*` Tag 推送触发，但截至 `v1.13-yz.4`，推送 Tag 实际未产生任何 workflow run（GitHub 侧未创建记录），`yz.2`、`yz.3`、`yz.4` 均通过 `workflow_dispatch` 传入 `release_tag` 发布。发布来源仍然固定：workflow 会 checkout 该 Tag 并校验 commit 一致后才继续构建。根因待查，暂按手动触发执行。在 Release 记录和六个资产出现前不能把 Tag 视为已发布。Xray fork 的回滚边界由 Node `go.mod` 中记录的 pseudo-version 和对应 fork commit 确定。

## 后续上游同步

同步新的 Xray 预发布 Tag 时：

1. 先记录官方 Tag 和对应 commit，再合并到 YZ fork；
2. 解决冲突时保留 Hysteria2 用户识别、统计、Dispatcher 和限速补丁及其测试；
3. 新上游版本的 fork patch 序列从 `yz.1` 重新开始；
4. 同步更新本文档、`go.mod/go.sum`、Node Release Tag、构建信息和变更说明。
