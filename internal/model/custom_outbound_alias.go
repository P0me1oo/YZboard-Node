package model

import "strings"

// 直连与拦截出站在两个内核里的原生名不同：xray 用 freedom / blackhole，
// sing-box 用 direct / block。面板里两种写法都允许，由各自的配置构建阶段翻译。

// IsDirectOutbound 判断是否为「直连」语义的出站。
func IsDirectOutbound(protocol string) bool {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "freedom", "direct":
		return true
	}
	return false
}

// IsBlockOutbound 判断是否为「拦截」语义的出站。
func IsBlockOutbound(protocol string) bool {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "blackhole", "block":
		return true
	}
	return false
}

// OutboundNeedsSettings 报告该协议是否必须提供 settings。
// 直连和拦截可以完全没有参数，其余协议至少要有连接信息。
func OutboundNeedsSettings(protocol string) bool {
	return !IsDirectOutbound(protocol) && !IsBlockOutbound(protocol)
}

// SendThroughKeys 是 settings 里会被提升到出站顶层的源地址字段。
// xray 的 sendThrough 是 outbound 级字段，无法放在 settings 内。
var SendThroughKeys = []string{"send_through", "sendThrough"}
