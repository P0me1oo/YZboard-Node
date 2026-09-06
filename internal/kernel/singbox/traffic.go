package singbox

import (
	"sync"
	"sync/atomic"
)

type userTraffic struct {
	upload   atomic.Int64
	download atomic.Int64
}

// trafficTotals 的生命周期跟随 Node 内核对象，重建 Box 时只更换连接状态。
// 旧连接持有同一组原子计数器，延迟完成的读写也不会在重启时丢失。
type trafficTotals struct {
	mu    sync.RWMutex
	users map[int]*userTraffic
}

func newTrafficTotals() *trafficTotals {
	return &trafficTotals{users: make(map[int]*userTraffic)}
}

func (t *trafficTotals) user(id int) *userTraffic {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.users[id] == nil {
		t.users[id] = &userTraffic{}
	}
	return t.users[id]
}

func (t *trafficTotals) snapshot() map[int][2]int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[int][2]int64, len(t.users))
	for id, user := range t.users {
		value := [2]int64{user.upload.Load(), user.download.Load()}
		if value[0] > 0 || value[1] > 0 {
			out[id] = value
		}
	}
	return out
}
