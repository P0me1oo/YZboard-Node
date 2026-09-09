package xray

import (
	"sync"
	"testing"
)

func TestLimitDispatcherSnapshotDuringConnectionChanges(t *testing.T) {
	dispatcher := newTestDispatcher()
	email := userEmail(1)
	dispatcher.UpdateLimits(map[string]int{email: 1}, map[string]int{email: 2}, nil)
	var work sync.WaitGroup
	work.Add(2)
	go func() {
		defer work.Done()
		for range 10000 {
			dispatcher.checkDeviceLimit(email, "127.0.0.1", true)
			dispatcher.delConn(email, "127.0.0.1")
		}
	}()
	go func() {
		defer work.Done()
		for range 10000 {
			ips, _ := dispatcher.GetConnectionState()
			if len(ips[1]) > 1 {
				t.Error("在线快照出现了不存在的设备")
				return
			}
		}
	}()
	work.Wait()
	ips, _ := dispatcher.GetConnectionState()
	if len(ips) != 0 {
		t.Fatal("连接全部关闭后仍报告在线设备")
	}
}
