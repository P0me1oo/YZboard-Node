# sing-box v1.14.0 升级验证

## 源码关系

Node 源码目标版本为 `v1.13-yz.17`。升级前 Node 提交为 `7802e87136e62ebfc79048207b39323556c7cabc`，原 sing-box 实际来源为 `cedar2025/sing-box` 的 `2e665cb7e295949ba7c5536f9b7754f94ab78cee`，基线是 `v1.14.0-alpha.2`。

新核心以官方 `v1.14.0` / `0b8995879f29a9b98ee027bc17b75e101445b238` 为固定基线。Node 仍需要用户热更新、路由热更新和 Mieru 兼容代码，兼容版本为 `v1.14.0-yz.1`。正式 `go.mod` 使用 `github.com/P0me1oo/YZ-sing-box v1.14.0-yz.1`；开发时的本地模块替换不属于正式发布配置。

兼容提交为 `f47d4d565a4371cf46b6c462612fc085f634a6af`，已推送至 [P0me1oo/YZ-sing-box](https://github.com/P0me1oo/YZ-sing-box/tree/v1.14.0-yz.1)。远程模块内容和依赖版本列表已与隔离实测使用的源码核对一致，校验值由 `go.sum` 固定。Node `v1.13-yz.17` 尚未发布。

YZboard 和 YZ-Xray-core 不需要配合修改。Xray replacement 保持 `github.com/P0me1oo/YZ-Xray-core v0.0.0-20260903142229-601226e180d3`。

## 测试范围

- TCP：VMess、VLESS、Trojan、普通 Shadowsocks、SS2022 AES-128/AES-256、SOCKS、HTTP、AnyTLS、Mieru 的 TCP/UDP 承载、Hysteria2、TUIC，共 13 个场景。
- UDP：VMess、VLESS、Trojan、普通 Shadowsocks、SS2022 AES-128/AES-256、Mieru 的 TCP/UDP 承载、Hysteria2、TUIC，共 10 个场景。
- 用户生命周期：初始连接、添加和重复添加、删除和重复删除、旧凭据拒绝、保留用户的已有连接、QUIC 会话新流、重复重载和停止后重启；清空用户后拒绝旧凭据及匿名 SOCKS/HTTP，随后可重新添加用户。
- 监听生命周期：Mieru 停止后立即释放 UDP 端口，端口被占用时 TCP/UDP 启动明确失败。
- 空用户重启：SOCKS/HTTP 保留认证要求，普通 Shadowsocks 和 SS2022 保留多用户模式，拒绝匿名代理或仅持有服务器密钥的客户端；重新添加用户后恢复通信。
- 流量：按用户累计，重复重载不重复注册统计器；单独验证上传、下载长度不同时的方向和次数。
- 路由：成功替换、相同配置去重、缺失或重复规则集失败、失败后旧规则继续生效、同名规则集更新、多标签本地规则集、并发匹配；空闲 DNS 连接和 TCP/UDP 入站转交不阻塞热更新。
- 安装：正式安装器的配置生成、程序安装、服务渲染和启动流程；重复安装保持单个实例；模拟面板下发用户、路由和错误配置；流量与状态上报；systemd 重启恢复。

Naive 客户端、Hysteria 1、生产证书申请和真实面板数据库不在本次运行验收范围内。安装测试使用独立模拟面板，不能等同于生产面板部署验收。

## 隔离安装验收

一台 Linux amd64 测试服务器即可承载模拟面板、Node 和客户端。测试要求 root、systemd、Python 3、bash、iproute2、unshare 和 curl；脚本不会安装系统依赖。

构建目录应包含以下文件，上传后先核对 `SHA256SUMS`：

```text
xboard-node-linux-amd64
xbctl-linux-amd64
install.sh
singbox_isolated_install.sh
singbox_install_probe.py
SHA256SUMS
```

确认目标服务器和需要保留的服务后，在该目录执行：

```bash
unshare --net --fork python3 ./singbox_install_probe.py \
  --artifacts "$PWD" \
  --version v1.13-yz.17 \
  --preserve-service xboard-node.service
```

开发测试包的版本带 `-test` 后缀，命令中的 `--version` 应与产物输出一致。

测试实例使用 `/var/tmp/yzboard-singbox-test-*` 下的新目录和 `/run/systemd/system/yzboard-singbox-test-*.service`。服务通过 `NetworkNamespacePath` 加入测试网络空间，监听只在该空间内可达；内存上限 768 MiB，CPU 配额为一个核心。临时 Token、UUID 和生成配置不会写入报告，测试服务不把运行日志持久化到 journal。

脚本在退出时停止和移除本次测试服务，删除本次实例、备份和凭据；保留上传的程序包供核对。原 `xboard-node.service` 的 PID 及常用安装文件校验值在测试前后必须一致。完全清理时再删除本次上传的程序包目录，不能扩大到既有服务和其他数据。

## 发布前检查

执行 `make test`、Linux amd64/arm64 构建，并用 `go version -m` 检查最终产物的模块替换、目标架构、来源提交和 `vcs.modified`。兼容核心、Node 版本、Tag、变更说明和产物校验值应一致。

本文件描述验证范围和执行方法。2026-09-05 的固定依赖核验、开发构建、全量 Linux race 与隔离安装结果已记录在 `YZ_COMPATIBILITY.md`。Windows 本机 race 受运行时内存映射错误限制，因此使用相同源码交叉编译测试包后在 Linux amd64 执行。发布构建必须使用正式 `go.mod`，产物中的 sing-box replacement 应为远程固定版本，且 `vcs.modified=false`；实际来源提交和校验值记录在随产物生成的 `build-metadata.json` 与 `SHA256SUMS` 中。
