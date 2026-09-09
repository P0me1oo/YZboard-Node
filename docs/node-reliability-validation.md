# Node v1.13-yz.19 修复验证

本次修复从 `89c2753390356f51df3d8fc133ae8064fa8ed669` 开始，只修改 YZboard-Node。正式发布的固定提交、镜像摘要和校验值见 [兼容矩阵](../YZ_COMPATIBILITY.md) 与 GitHub Release。

sing-box 继续固定为 `github.com/P0me1oo/YZ-sing-box v1.14.0-yz.1`；Xray 继续固定为 `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260903142229-601226e180d3`。面板配置、用户、报告格式和 Xray 中转计数器名称保持兼容。

首次发布前 [CI](https://github.com/P0me1oo/YZboard-Node/actions/runs/34061908602) 在 AnyTLS 用户删除测试中发现上游 `v0.0.11` 的关闭状态数据竞争，阻止了发布。`v0.0.13` 仍保留同样的未保护访问，因此 Node 使用 [固定上游源码的兼容补丁](../compat/sing-anytls/README.yz.md)，只修改流关闭状态和回调访问的两个源文件；原许可证、作者和原始文件校验值保留。原有测试要求没有放宽。

## 运行行为

| 场景 | 行为 |
| --- | --- |
| 配置、协议、端口、出站或证书应用失败 | 记录操作、内核和底层原因，停止对应内核，保留最新配置和用户快照 |
| 初始、REST、WebSocket 配置校验失败 | 无效快照交给节点服务统一处理，控制通道继续接收修正，旧配置不再运行 |
| 重复失败快照 | 定时采样、REST 和 WebSocket 不反复启动；配置、用户或证书材料变化后才有新的尝试机会 |
| 用户变更后的恢复 | 全量更新、增量增删和 UUID 替换都会重新检查配置及证书，不能绕过未修正错误 |
| 删除问题用户 | 立即尝试启动剩余用户；重复删除不存在的用户不会触发重复启动 |
| 删除最后一个用户 | 保留明确的空集合，清空认证列表；落地节点按设计允许空用户启动 |
| 多节点或多实例 | 单个节点端口冲突、配置错误或初始化失败不取消其他节点；`/healthz` 的 503 表示至少一项失败 |
| sing-box 运行配置变化 | 协议、端口、出站等变化完整重建；纯路由规则和独立用户更新继续热更新 |
| Xray 重载 | 先关闭旧监听，已有连接继续按旧用户及旧中转映射结算，最多排空五分钟，进程退出时缩短为五秒 |
| Xray 连接及在线 IP 快照 | 持锁复制嵌套 map，避免与连接增删并发读写 |
| 流量累计 | 同一 Kernel 对象的 Start、Reload、Stop 不清空累计基线；退出和旧实例释放时补采尾部流量 |
| 退出上报 | 等待在途报告，失败批次复用原 report_id 和完整内容，成功后发送独立的最终增量 |
| 证书续期 | REST 对账成功后才消费续期标记，请求失败不会丢失后续重载 |

sing-box 完整重建最多等待已有连接五秒，剩余连接可能中断；失败不会恢复旧监听。流量连续性限于同一进程，本次没有增加进程崩溃后的磁盘持久化。初始化时无法取得面板快照的网络错误属于服务初始化失败；多节点进程中的其他服务仍继续运行。

## 验证结果（2026-09-07）

| 检查 | 结果 |
| --- | --- |
| Windows amd64 全量测试 | 18 个测试包、586 项测试及子测试通过，无失败、无跳过，包含 AnyTLS 兼容包 |
| 静态及脚本检查 | Go vet、安装器服务模板测试、Bash/Python 语法、CI actionlint 通过 |
| GitHub Linux 全量并发测试 | 18 个测试包、586 项测试及子测试通过；发布前及正式发布 CI 均通过 |
| YT-HK Linux amd64 并发检测 | 5 个包、331 项测试及子测试通过，无失败、无跳过、无数据竞争 |
| AnyTLS 重复并发检测 | 底层关闭测试 50 轮、真实用户生命周期 20 轮全部通过，无跳过、无数据竞争 |
| 首次端口冲突 | 进程与健康端点存活；相同快照只启动一次，修正端口后恢复实际传输 |
| 证书与无效出站 | 坏证书不能被用户更新绕过；初始和运行中无效出站均停止监听，修正后恢复实际传输 |
| 慢报告正常返回 | SIGTERM 后等待约 18 秒，退出码 0，发送两个批次并保留尾部流量 |
| 慢报告失败重试 | SIGTERM 后等待约 18 秒，退出码 0，发送三个请求；前两次 ID 和内容相同，最终增量使用新 ID |
| 两节点故障隔离 | 机器模式和传统 nodes 模式各验证端口冲突、无效初始出站、初始 HTTP 失败；6 个场景中健康节点均继续传输 |
| 双架构构建 | Node、xbctl 的 linux/amd64 与 linux/arm64 构建及元数据检查通过；amd64 版本命令在测试机运行通过 |
| 正式发布 | `v1.13-yz.19` 的 10 个附件摘要通过核对，双架构镜像实际执行版本命令通过，OCI 来源与 latest 指向通过核对；固定摘要见兼容矩阵 |

并发检测包分别通过 controlplane 18 项、service 59 项、singbox 126 项、xray 123 项、anytls 5 项。AnyTLS 底层重复测试共通过 250 项测试及子测试，真实用户生命周期重复测试共通过 40 项测试及子测试。额外的实际进程验收共 12 个场景全部通过。配置与用户恢复测试还覆盖 WebSocket 转换、机器邮箱、REST ETag 对账、增量删除和 UUID 变更。

本轮使用 Go 1.26.4。最终验收应用标识为 `v1.13-yz.19-test`、`d257b4db9d4d74ad6ac6994fea78d7302eb126ce-dirty`；`go version -m` 保留真实的 `vcs.modified=true`。五个 Linux 并发检测程序使用 `-race`、Zig 0.14.1 和 glibc 2.31 目标。验收包 `runtime-validation.tar.xz` 为 130250888 字节，SHA256 为 `aa3f59474867e7c727beb793af1186044b44ab97dee2298da40f49bcbd67f424`，不作为正式安装包。

每个验收二进制的实际模块、目标架构、构建参数和 SHA256 记录在 `build-metadata.json`；149 份 Go 源码及模块文件记录在 `source-manifest.json`，打包前已逐项核对。Node 二进制同时确认使用 `./compat/sing-anytls`。正式发布从固定 Git 提交重新构建，CI 检查 `vcs.modified=false`、来源提交及目标架构，Release 附带 `.buildinfo.txt` 和 `SHA256SUMS`。Docker 使用相同工具链和五个功能标签，版本检查通过后才更新正式版本与 latest 标签。

镜像运行验证从固定 OCI index 中分别解析 amd64、arm64 的唯一 manifest 摘要，再按对应摘要执行版本命令，避免 Docker 本地镜像存储在同一 index 摘要下切换架构时发生冲突；arm64 使用 QEMU 执行。

## 复现命令

```bash
go test -mod=readonly -p=2 -count=1 -timeout=5m \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./...
go test -mod=readonly -race -count=1 github.com/anytls/sing-anytls/session
go vet -mod=readonly -p=2 \
  -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./...
bash tests/install_service_manager_test.sh
bash -n install.sh tests/install_service_manager_test.sh tests/singbox_isolated_install.sh
```

Linux 验收需要一台 amd64 测试机，具备 Python 3.11、iproute2、unshare、sha256sum。使用预编译包无需在服务器安装 Go 或 C 编译器；先核对包内 SHA256SUMS，再运行：

```bash
sha256sum --check --strict SHA256SUMS
unshare --net --fork python3 run_linux_regressions.py --artifacts "$PWD"
unshare --net --fork python3 runtime_failure_probe.py --artifacts "$PWD"
unshare --net --fork python3 multi_node_failure_probe.py --artifacts "$PWD"
unshare --net --fork python3 repeat_anytls.py
```

`run_linux_regressions.py` 需要五个 `*-linux-amd64.test` 程序，分别为 controlplane、service、singbox、xray、anytls；后两个脚本需要 `xboard-node-linux-amd64`。这三个脚本都拒绝在主机原网络空间运行，使用独立测试身份、临时目录、回环监听和模拟面板。报告只保留结果与计数，原始日志和凭据不落入报告。

验收包内的 `repeat_anytls.py` 在相同隔离环境中额外执行 50 轮 Stream 关闭回归和 20 轮 `TestSingBoxRuntimeUserLifecycle/anytls`，仅保存状态和计数。

## 部署停止等待时间

进程退出上限为两分钟，用于覆盖在途报告、失败重试、最终报告各 30 秒及内核排空；第二次退出信号仍会立即强制退出。systemd/OpenRC 模板等待 150 秒，Docker 应设置 `--stop-timeout=150`，Compose 应设置 `stop_grace_period: 150s`。已有服务通过 xbctl 单独升级二进制时，不会自动重写服务文件，需要同步等待设置或使用安装器升级，详见 [README](../README.md#upgrade)。

## 环境与范围

YT-HK 测试结束后原有 7 个监听的快照未变化，没有遗留测试进程；本次上传包、解压目录、一次性配置和凭据均已清理，Netcatty 会话已关闭。内核监听、用户认证和累计流量由现有回归覆盖，额外多节点实测使用 sing-box SOCKS/HTTP 与 Xray VLESS。

本轮未覆盖真实面板数据库、生产证书、外部代理链路或系统服务安装。arm64 完成构建、元数据检查及 CI 中的 QEMU 镜像版本命令验证，未在 arm64 实机测试代理流量。
