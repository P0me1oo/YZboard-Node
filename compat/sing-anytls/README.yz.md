# AnyTLS 关闭状态兼容补丁

源码来自 `anytls/sing-anytls v0.0.11`，固定提交 `130d2e61b8895727bfed4942c535e91b246a9603`。原许可证和作者信息保留，原始文件校验值见 `UPSTREAM.json`。

Node 的 AnyTLS 用户生命周期测试在认证拒绝时触发上游数据竞争：接收协程关闭 Stream，同时客户端写入目标地址或注册关闭回调。`dieErr` 与 `dieHook` 缺少并发保护。上游 v0.0.13 仍保留这些未保护访问，不能通过更新该 Tag 解决。

本补丁只修改 `session/stream.go` 与 `session/client.go`：对关闭错误和回调加锁；关闭回调在锁外执行；回调注册晚于关闭时仍执行一次。协议格式、认证、会话复用和默认配置保持原语义。新增测试覆盖并发读写与关闭、关闭后的回调注册以及并发注册和重复关闭。

补丁通过 Node 根模块的本地 replace 使用，随 Node 的固定提交一起发布，不冒充上游新版本。CI 明确执行 `go test -race github.com/anytls/sing-anytls/session`，并保留 Node 原有完整 AnyTLS 生命周期回归。

影响范围为 Node 内嵌 AnyTLS 入站、出站及测试客户端。后续上游提供等价并发修复且现有接口兼容时，可以删除此目录和根模块 replace，固定到通过同一组测试的上游版本。
