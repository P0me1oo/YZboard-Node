# Custom Outbounds

## Quick Example

```json
{
  "custom_outbounds": [
    {
      "tag": "warp",
      "protocol": "wireguard",
      "settings": {
        "peers": [{"address": "162.159.195.1"}],
        "private_key": "aMzA..."
      }
    },
    {
      "tag": "us-proxy",
      "protocol": "socks",
      "settings": {"servers": [{"address": "1.2.3.4", "port": 1080}]},
      "proxy_tag": "warp"
    }
  ]
}
```

## Supported Protocols

| Protocol | Xray | Sing-box |
|----------|------|----------|
| direct / freedom | ✅ | ✅ |
| block / blackhole | ✅ | ✅ |
| vmess | ✅ | ✅ |
| vless | ✅ | ✅ |
| trojan | ✅ | ✅ |
| shadowsocks | ✅ | ✅ |
| socks | ✅ | ✅ |
| http | ✅ | ✅ |
| wireguard | ✅ | ✅ |
| tuic | ❌ | ✅ |
| hysteria2 | ❌ | ✅ |
| anytls | ❌ | ✅ |
| naive | ❌ | ✅ |
| mieru | ❌ | ✅ |

## Core Fields

| Field | Required | Description |
|-------|----------|-------------|
| `tag` | ✅ | Outbound tag, referenced in routes |
| `protocol` | ✅ | Protocol type |
| `settings` | ✅ | Protocol-specific configuration（`direct`/`block` 可省略） |
| `proxy_tag` | ❌ | Chain to another outbound (references its tag) |
| `settings.send_through` | ❌ | 绑定出站源地址；xray 下会被提升为 outbound 级的 `sendThrough` |

## 直连与拦截

两个内核的原生名不同，面板两种写法都接受，由 Node 翻译：

| 面板填写 | Xray | Sing-box |
| --- | --- | --- |
| `direct` 或 `freedom` | `freedom` | `direct` |
| `block` 或 `blackhole` | `blackhole` | `block` |

`settings` 里的其余字段**原样透传**，不做跨内核翻译，因此要按目标内核的原生字段名填写。
两边的语义并不等价，自动翻译反而会掩盖差异。

绑定源地址、解析策略在两个内核里的写法：

```jsonc
// Xray：sendThrough 是 outbound 级字段，这里写在 settings 里由 Node 提升
{
  "tag": "direct",
  "protocol": "freedom",
  "settings": {
    "domainStrategy": "UseIPv6v4",
    "send_through": "2001:db8::1"
  }
}

// Sing-box：绑定地址和解析策略都在出站对象上
{
  "tag": "direct",
  "protocol": "direct",
  "settings": {
    "inet4_bind_address": "203.0.113.9",
    "inet6_bind_address": "2001:db8::1",
    "domain_strategy": "prefer_ipv6",
    "fallback_delay": "300ms"
  }
}
```

注意行为差异：sing-box 的 `prefer_ipv6` + `fallback_delay` 是连接级的 happy-eyeballs 回落，
可以同时绑定 v4 和 v6 源地址；xray 的 `domainStrategy: UseIPv6v4` 只影响**解析顺序**，
`sendThrough` 也只能填一个地址，绑定 v6 源之后无法连接纯 v4 目标。
需要严格双栈回落时，xray 下要拆成两个出站并配合路由分流。

自定义出站的 tag 与内置的 `direct`、`block` 同名时，**面板配置优先**，内置的那个不再生成。
这正是覆盖默认直连出站的方式。

## Best Practices

- **Prefer** `custom_outbounds`: panel-managed, cross-kernel compatible
- **Use** `kernel.custom_outbound` only: when native fields are needed
