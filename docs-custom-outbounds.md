# 自定义出站

面板通过 `custom_outbounds` 下发出站。Node 沿用现有面板格式，并按所选内核生成配置。
`tag`、`protocol` 等外层字段通用，`settings` 中仅有明确对应关系的参数会自动转换，其余参数仍使用目标内核的原生格式。

## 直连配置示例

从 `v1.13-yz.24` 起，下面这份现有格式可以同时用于 Xray 和 sing-box，无须在切换内核时重填：

```json
{
  "custom_outbounds": [
    {
      "tag": "direct",
      "protocol": "direct",
      "settings": {
        "domainStrategy": "ForceIPv4",
        "send_through": "192.0.2.10"
      }
    }
  ]
}
```

`192.0.2.10` 是文档示例地址，使用时填写服务器已有的实际出口地址。
IPv6 出口使用 `ForceIPv6` 和对应的 IPv6 地址，例如文档地址 `2001:db8::10`。
出站标签与内置 `direct`、`block` 同名时，面板配置优先；其他标签可由路由引用。

## 字段与协议

| 字段 | 要求 | 含义 |
| --- | --- | --- |
| `tag` | 必填 | 路由引用的出站标签 |
| `protocol` | 必填 | 出站协议 |
| `settings` | 按协议填写 | 直连与拦截可以省略，其余协议需要连接参数 |
| `proxy_tag` | 可选 | 引用另一个出站标签作为代理链 |
| `settings.send_through` | 可选 | 单一出口源地址；直连出站在两种内核中均保留该地址约束 |

Node 的协议校验允许以下类型，具体原生参数仍由目标内核校验：

| 协议 | Xray | sing-box |
| --- | --- | --- |
| direct / freedom | 支持 | 支持 |
| block / blackhole | 支持 | 支持 |
| vmess / vless / trojan / shadowsocks | 支持 | 支持 |
| socks / http / wireguard | 支持 | 支持 |
| tuic / hysteria2 / anytls / naive / mieru | 不支持 | 支持 |

## 现有直连参数的自动转换

转换只作用于面板 `custom_outbounds` 中的 `direct`／`freedom` 出站；本地 `kernel.custom_outbound` 和 `kernel.custom_config` 继续使用原生配置。

| 面板现有写法 | Xray | sing-box |
| --- | --- | --- |
| `protocol: direct` 或 `freedom` | `freedom` | `direct` |
| `protocol: block` 或 `blackhole` | `blackhole` | `block` |
| `domainStrategy: ForceIPv4` | 保留原策略 | `domain_strategy: ipv4_only` |
| `domainStrategy: ForceIPv6` | 保留原策略 | `domain_strategy: ipv6_only` |
| `domainStrategy: AsIs` | 保留原策略 | 未绑定源地址时使用 `as_is` |
| `send_through` 或 `sendThrough` | 提升为出站顶层 `sendThrough` | 按地址族生成 `inet4_bind_address` 或 `inet6_bind_address` |

单地址绑定同时约束域名解析和实际连接：指定 IPv4 源地址后，不能通过 IPv6 默认出口发送；IPv6 源地址同理。
未填写解析策略或填写 `AsIs` 时，sing-box 按绑定源地址限制解析地址族。
该约束覆盖 TCP、UDP、路由提前解析的目标以及 UDP 会话后续数据包。
IPv4 映射地址（例如 `::ffff:192.0.2.10`）按 IPv4 处理，避免误选未绑定的 IPv6 出口。

以下配置会明确报错，交由现有失败处理停止节点并等待配置修正：

- 两个源地址别名填写了不同地址。
- 强制解析策略与绑定源地址的地址族冲突。
- 旧参数与 `domain_strategy`、`domain_resolver.strategy` 或原生绑定地址冲突。
- 单地址绑定同时混入另一地址族的原生绑定。
- 绑定值不是具体的 IPv4／IPv6 地址，例如域名、地址段、未指定地址或带区域标识的地址。
- 将尚未提供等价转换的 Xray 策略传给 sing-box，例如 `UseIPv6v4`。

转换生成新对象，不改写面板保存的 `settings`；重复同步、重载和切回 Xray 继续使用原有参数。
原生 sing-box 绑定配置不附加单地址限制，保留原来的双栈行为。

## 原生解析优先级与回落

`UseIPv6v4` 等策略涉及解析失败后的处理，不能直接等同于 sing-box 的连接级回落。
需要这些能力时，继续填写目标内核的原生参数。例如 sing-box 的双栈配置：

```json
{
  "tag": "direct",
  "protocol": "direct",
  "settings": {
    "inet4_bind_address": "192.0.2.10",
    "inet6_bind_address": "2001:db8::10",
    "domain_strategy": "prefer_ipv6",
    "fallback_delay": "300ms"
  }
}
```

此例只适用于 sing-box，不能作为可直接切换到 Xray 的配置。
Xray 的单一 `sendThrough` 绑定无法表示两个源地址，需要按其原生能力拆分出站并配置路由。
其他协议的 `settings` 也不会因本次直连兼容而自动获得完整的跨内核转换。
