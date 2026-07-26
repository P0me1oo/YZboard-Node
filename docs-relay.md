# 中转节点（单入口多落地）

面板把一个 VLESS 入口节点和若干落地节点组成一组中转拓扑。客户端只连接入口，
订阅里的每个逻辑节点都使用同一个入口地址、端口和 Reality 参数，区别只在写入
UUID 的路由编号。Node 不需要额外的本地配置，全部参数由面板通过 `relay` 段下发。

## 路由编号

编号是一个 1-65535 的整数，写在客户端 UUID 的第 7、8 字节（0 基下标 6、7），
也就是标准写法的第三段。Xray 在校验 VLESS 用户前会把这两个字节清零，因此写入
编号不会影响用户身份匹配；认证通过后内核再按原始字节还原编号，交给路由规则的
`vlessRoute` 匹配项使用。

编号 0 不可用：Xray 的端口列表解析会丢弃数字 0。

该机制由 Xray 提供，因此中转拓扑的两端都必须使用 xray 内核，sing-box 会在配置
校验阶段直接报错。

安装器默认写入 xray；已经装好的节点可以用下面的命令切换，改完重启服务生效：

```bash
xbctl config kernel xray            # 全部实例
xbctl config kernel xray --instance <实例 ID>
xbctl config kernel singbox         # 换回去
systemctl restart xboard-node
```

xray 能承载的入站协议少于 sing-box（不支持 tuic、naive、anytls、mieru、socks、http），
因此切到 xray 时命令会拒绝这些协议的实例，确认要切再加 `--force`。

## 入口节点

面板下发：

```json
{
  "protocol": "vless",
  "server_port": 24443,
  "relay": {
    "mode": "entry",
    "route_id": 11,
    "children": [
      {
        "node_id": 7,
        "tag": "relay-7",
        "route_id": 12,
        "protocol": "shadowsocks",
        "address": "203.0.113.7",
        "port": 28388,
        "cipher": "2022-blake3-aes-128-gcm",
        "password": "<内部凭据>"
      }
    ]
  }
}
```

Node 据此生成：

- 一个真实的 VLESS + Reality 入站，端口和参数与普通节点完全一致；
- 每个逻辑节点一个独立的 Shadowsocks 出站，标签为 `relay-<逻辑节点 ID>`；
- 每个编号一条路由规则，入口自身的编号指向 `direct`。

逻辑节点数量不影响入站数量，入口上永远只有一个客户端入站。

### 路由规则优先级

生成的规则按以下顺序排列：

1. 结构化 `custom_route_rules`
2. 原始 `custom_routes`
3. 内置私有地址与回环地址拦截
4. 中转路由规则
5. 面板路由组

中转规则排在面板路由组之前，保证选中某个逻辑节点时出口稳定；管理员显式配置的
自定义规则和内网拦截仍然优先。

## 落地节点

面板下发：

```json
{
  "protocol": "shadowsocks",
  "server_port": 28388,
  "relay": {
    "mode": "landing",
    "protocol": "shadowsocks",
    "listen_port": 28388,
    "cipher": "2022-blake3-aes-128-gcm",
    "password": "<内部凭据>",
    "entry_node_id": 3
  }
}
```

Node 生成一个标签为 `relay-in` 的 Shadowsocks 入站，只包含这一份内部凭据，
不包含任何面板用户，同时开启 `tcp,udp`。落地服务器上的 Xray 直接向目标建立
连接，不需要开启系统转发或出口地址转换。

因为落地节点没有面板用户，Node 对这类节点放宽了“用户列表为空就停止内核”的
规则：空用户集是落地节点的正常状态，内核会保持监听。

## 支持的内部加密算法

`2022-blake3-aes-128-gcm`、`2022-blake3-aes-256-gcm`、`2022-blake3-chacha20-poly1305`、
`aes-128-gcm`、`aes-192-gcm`、`aes-256-gcm`、`chacha20-ietf-poly1305`、
`xchacha20-ietf-poly1305`。

内部凭据由面板派生后分别下发给两端，不写入 Node 的本地配置文件，也不出现在
客户端订阅中。

## 流量统计

- 用户流量仍然只在入口按真实 VLESS 用户身份统计一次，通过 `traffic` 上报。
- 入口额外读取各 `relay-<ID>` 出站的计数器，通过 `relay_traffic` 上报，键为
  逻辑节点 ID。面板把它记入对应落地节点的节点流量，不参与用户套餐扣费。
- 落地节点的内部入站没有面板用户，因此不会重复上报用户流量。

Shadowsocks 的协议封装会让出站统计与用户有效流量存在小幅差异，这属于不同统计
口径，只用于节点运营信息。

## 幂等与生命周期

配置哈希覆盖整个 `relay` 段，重复同步同一份配置不会重建实例，也不会产生重复的
出站或路由规则。逻辑节点被禁用或删除后，面板不再下发对应的 child，入口重新生成
配置时该出站和路由规则一并消失，真实入站和其他逻辑节点不受影响。
