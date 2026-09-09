package model

func OutboundSupportMatrix() map[string]KernelOutboundSupport {
	return map[string]KernelOutboundSupport{
		"xray": {
			Protocols: []string{
				"vmess",
				"vless",
				"trojan",
				"shadowsocks",
				"socks",
				"http",
				"wireguard",
				// 直连与拦截：两个内核的原生名不同，两种写法都收，由各自的
				// 配置构建阶段翻译成内核名（xray 为 freedom / blackhole）。
				"freedom",
				"direct",
				"blackhole",
				"block",
			},
			Features: []string{
				"tag",
				"protocol",
				"settings",
				"proxy_tag",
			},
		},
		"singbox": {
			Protocols: []string{
				"vmess",
				"vless",
				"trojan",
				"shadowsocks",
				"socks",
				"http",
				"wireguard",
				"tuic",
				"hysteria2",
				"anytls",
				"naive",
				"mieru",
				"freedom",
				"direct",
				"blackhole",
				"block",
			},
			Features: []string{
				"tag",
				"protocol",
				"settings",
				"proxy_tag",
			},
		},
	}
}

type KernelOutboundSupport struct {
	Protocols []string
	Features  []string
}
