# 自动防火墙与端口跳跃验证

本轮目标版本为 Node `v1.13-yz.24`、面板 `1.13.0`，均为待发布状态。没有修改 Xray、sing-box 的固定依赖，没有进行生产部署。正式版本和历史回滚引用见 [兼容矩阵](../YZ_COMPATIBILITY.md)。

## 环境与范围

本地使用 Windows、Go 1.26.4 和 PHP 8.4.21。经授权在 YT-HK 执行 Linux 验证，防火墙工具和测试程序位于专用 rootfs，使用私有挂载、网络和进程命名空间；客户端位于第二个网络命名空间。未安装宿主机软件，也未修改宿主机防火墙、路由或现有服务。

| 工具 | 版本 |
| --- | --- |
| UFW | `0.36.2` |
| firewalld | `1.3.3` |
| nftables | `1.0.6` |
| iptables / ip6tables | `1.8.9`，`nf_tables` 后端 |
| 官方 Hysteria 客户端 | `app/v2.6.2`，Linux amd64 |
| Hysteria 客户端 SHA256 | `966ebeb1ff2d878b4a4bcf1e6e588d7c3c239ade9115b6ef14d0f6b7314fd2bd` |

完整 Node 测试使用临时 Python 面板接口替身，核对真实机器发现、配置获取、用户获取、状态和流量回传；Laravel 面板配置接口另由本地 PHP 功能测试验证。本次没有在服务器部署完整 Laravel 面板。

测试身份在每次执行时随机生成，配置、证书、日志和归属记录位于私有 `/run`，进程退出后删除。失败日志会隐藏临时认证值，不记录真实节点身份。

## 已完成的本地检查

```bash
# Node 仓库
go test -mod=readonly -tags 'with_quic with_utls with_wireguard with_acme with_clash_api' ./...
# 面板仓库
php -d extension=pdo_sqlite -d extension=sqlite3 vendor/bin/phpunit --colors=never
```

Node 全部普通 Go 测试通过；新防火墙覆盖共享引用、重复同步、范围冲突、手工规则归属、失败回收和并发退出。面板完整回归通过，共 112 项测试、1762 个断言，覆盖端口集合、实际配置接口、停用节点发现、订阅输出和保存校验。SQLite 扩展通过 PHP 进程参数加载，没有修改全局 PHP 配置。

## Linux 原生命令检查

| 放行后端 | 转发后端 | 结果 |
| --- | --- | --- |
| UFW | nftables | 通过 |
| UFW | iptables | 通过 |
| firewalld | nftables | 通过 |
| firewalld | iptables | 通过 |

四组均使用真实 TCP/UDP 和 IPv4/IPv6 数据包，检查连续范围与离散端口、指定监听地址、共享引用、重复应用、范围重载、手工规则保留、缺失规则恢复、正常清理和异常退出后的归属恢复；UFW 另外检查停用状态下的持久规则清理。

firewalld 自身 `reload` 会删除测试中的 xtables 运行时手工链，已独立复现。测试先验证 Node 操作保留对照规则，再单独执行外部重载并验证 Node 恢复；不将 firewalld 的重载行为归为 Node 删除手工规则。

## 完整 Node 和官方客户端检查

| Node 内核 | 放行 / 转发后端 | 结果 |
| --- | --- | --- |
| Xray | UFW / nftables | 通过 |
| sing-box | firewalld / iptables | 通过 |

测试使用真实监听端口 `28443`，客户端初始端口为 `29440-29441,29500`，重载为 `29600-29601,29700`。抓包检查至少观察到两个不同 UDP 目的端口，同时持续校验完整 HTTP 响应。范围变化前后比较监听 socket 的 inode，确认仅改变跳跃范围不会重建内核监听。

生命周期检查包括 IPv6 实际代理请求、面板停用和重新启用、监听端口被占用时的失败清理、修正配置后恢复、状态和流量回传、正常退出清理，以及 `SIGKILL` 后重启回收遗留规则。

调试时确认了两项测试环境要求：Xray 默认拒绝保留地址，测试仅用已有自定义出站配置允许隔离的 HTTP 对照地址和端口；机器发现的间隔需至少 30 秒，更小的值会按既有逻辑回退到 60 秒。没有为适配测试修改核心默认策略或机器发现周期。

## Linux 竞争检测

在专用 rootfs 内使用 Go 1.26.4、GCC 12.2.0 和 `CGO_ENABLED=1` 编译 `internal/firewall` 的 `-race` 测试二进制，再在独立网络中运行。管理器、归属状态和解析测试全部通过，没有数据竞争报告，包含并发申请、释放与退出的回归。

本次竞争检测没有开启原生命令的环境开关，两项原生入口测试按设计跳过；它们已在上面的四组独立原生测试中执行。此次没有运行全仓库 `make test` 的全部竞争检测，不将普通 Go 回归与单包竞争检测写成全量竞争检测。

## 构建与验证限制

合并固定安装器提交 `b5ce51f5dea000545d8a4ea6adde0074a9ee7146` 后，上述完整普通 Go 回归再次通过；`cmd/xbctl`、`internal/config`、`internal/buildinfo`、`internal/firewall` 的定向测试及 xbctl、防火墙的 `go vet` 通过。安装器脚本语法、服务文件测试和 20 个安装目录场景通过；迁移测试使用 C、F 两个文件系统，所有操作位于测试临时目录。

Node 功能提交为 `ade17c59c021fdeeec7a19c4ebb5e5cb9ccce17f`，合并后的构建来源为 `7b3a7b434790ecf7238b4977d02db5610b4db7e5`，配套面板功能提交为 `cee871e25a54b180e384346689664f538c8c2887`。合并只增加已验证的安装器和 xbctl 改动，Node 运行代码、内核依赖与实机测试时一致。

从干净提交生成 Node 和 xbctl 的 Linux amd64、arm64 二进制，使用版本 `v1.13-yz.24-dev`、Go 1.26.4、`CGO_ENABLED=0`、`-mod=readonly`、`-trimpath` 和 `-buildvcs=true`。四份构建信息均确认完整来源、目标系统和架构、`vcs.modified=false`；两个 Node 二进制实际解析到兼容矩阵中的 Xray 和 sing-box 固定版本。

Go 在 `-trimpath` 构建信息中不记录 `ldflags`，因此另外核对二进制里的版本值。最终 amd64 两个程序上传后重新核对摘要，并在隔离 rootfs 中执行版本命令，运行输出与上述版本、来源一致。arm64 完成编译、版本值和元数据检查，未进行 arm64 运行或实际转发。

| 程序 | 架构 | 字节数 | SHA256 |
| --- | --- | --- | --- |
| Node | Linux amd64 | `74907810` | `9be64a99154b338c8c045a46fa0bfdf6f202170fea670ff079705a19a3950aba` |
| xbctl | Linux amd64 | `7278754` | `77132c1b084f73155dc9f06e4f01e3274c03a9afe63083061056fdb4242f265a` |
| Node | Linux arm64 | `69402786` | `5ec1cbda7ba0e73c598a841cf5d9d20fd31554eddaa25c40d0cb800d43c92e62` |
| xbctl | Linux arm64 | `6750370` | `51cdbf55f3dd7d7d739e2369c6350d3e58c51c881e16dbd2321e9cb40daf4475` |

本轮是独立功能分支的测试构建，没有创建正式 Tag、Release 或生产镜像。源码记录与二进制元数据用于后续合并和发布审核，不替代正式发布验证。

## 测试清理和运行限制

YT-HK 本次 rootfs、临时程序、实例、认证值、证书和日志均已清理；清理前检查专用目录没有剩余挂载或以该 rootfs 运行的进程。本次打开的 SSH 会话已关闭。仅保留本地四个构建产物、构建元数据和校验清单。

运行中的 Node 每 30 秒恢复缺失规则。正常退出会回收规则；`SIGKILL` 或断电无法立即执行清理，需以后再次使用同一配置路径启动回收。已有手工放行不由 Node 删除，因此停用节点后仍可能保留管理员主动开放的端口。
