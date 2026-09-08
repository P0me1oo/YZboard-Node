// Package portset 解析、合并端口列表和连续范围，不展开为逐端口规则。
package portset

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Range struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func (r Range) String() string {
	if r.From == r.To {
		return strconv.Itoa(r.From)
	}
	return fmt.Sprintf("%d-%d", r.From, r.To)
}

func (r Range) Contains(port int) bool { return r.From <= port && port <= r.To }

func (r Range) Overlaps(other Range) bool { return r.From <= other.To && other.From <= r.To }

func (r Range) Valid() bool { return r.From > 0 && r.From <= r.To && r.To <= 65535 }

// Parse 接受 443、20000-20100、443,20000-20100 等形式，拒绝空项和倒序范围。
func Parse(raw string) ([]Range, error) {
	var ranges []Range
	for _, item := range strings.Split(raw, ",") {
		parts := strings.Split(strings.TrimSpace(item), "-")
		if len(parts) > 2 {
			return nil, fmt.Errorf("端口范围格式无效")
		}
		first, err := parsePort(parts[0])
		if err != nil {
			return nil, err
		}
		last := first
		if len(parts) == 2 {
			last, err = parsePort(parts[1])
			if err != nil {
				return nil, err
			}
		}
		if first > last {
			return nil, fmt.Errorf("端口范围起点不能大于终点")
		}
		ranges = append(ranges, Range{From: first, To: last})
	}
	return Merge(ranges), nil
}

func parsePort(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("端口列表不能包含空项")
	}
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("端口必须是 1 到 65535 之间的整数")
		}
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口必须是 1 到 65535 之间的整数")
	}
	return port, nil
}

func Merge(ranges []Range) []Range {
	if len(ranges) == 0 {
		return nil
	}
	ordered := append([]Range(nil), ranges...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].From < ordered[j].From })
	merged := []Range{ordered[0]}
	for _, current := range ordered[1:] {
		previous := &merged[len(merged)-1]
		if current.From <= previous.To+1 {
			previous.To = max(previous.To, current.To)
		} else {
			merged = append(merged, current)
		}
	}
	return merged
}

// Without 排除真实监听端口，避免生成指向自身的重定向规则。
func Without(ranges []Range, port int) []Range {
	var result []Range
	for _, r := range ranges {
		if !r.Contains(port) {
			result = append(result, r)
			continue
		}
		if r.From < port {
			result = append(result, Range{From: r.From, To: port - 1})
		}
		if port < r.To {
			result = append(result, Range{From: port + 1, To: r.To})
		}
	}
	return result
}
