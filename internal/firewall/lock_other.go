//go:build !linux

package firewall

import "os"

// 非 Linux 平台不执行系统防火墙操作，此实现仅供状态存储单元测试使用。
func lockState(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
}

func syncStateDirectory(string) error { return nil }
