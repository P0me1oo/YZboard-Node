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
	mu         sync.RWMutex
	users      map[int]*userTraffic
	relays     map[int]*userTraffic
	relayUsers map[int]map[int]*userTraffic
}

func newTrafficTotals() *trafficTotals {
	return &trafficTotals{users: make(map[int]*userTraffic), relays: make(map[int]*userTraffic), relayUsers: make(map[int]map[int]*userTraffic)}
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

// 中转计数与套餐计数共享读写事件，但分开累计，重载和停止不清空。
type relayCounters struct {
	node *userTraffic
	user *userTraffic
}

func (c relayCounters) upload(n int64) {
	if c.node != nil {
		c.node.upload.Add(n)
		c.user.upload.Add(n)
	}
}

func (c relayCounters) download(n int64) {
	if c.node != nil {
		c.node.download.Add(n)
		c.user.download.Add(n)
	}
}

func (t *trafficTotals) relayCounters(userID, nodeID int) relayCounters {
	if userID <= 0 || nodeID <= 0 {
		return relayCounters{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.relays[nodeID] == nil {
		t.relays[nodeID] = &userTraffic{}
	}
	if t.relayUsers[userID] == nil {
		t.relayUsers[userID] = make(map[int]*userTraffic)
	}
	if t.relayUsers[userID][nodeID] == nil {
		t.relayUsers[userID][nodeID] = &userTraffic{}
	}
	return relayCounters{node: t.relays[nodeID], user: t.relayUsers[userID][nodeID]}
}

func snapshotTraffic(counters map[int]*userTraffic) map[int][2]int64 {
	out := make(map[int][2]int64, len(counters))
	for id, counter := range counters {
		value := [2]int64{counter.upload.Load(), counter.download.Load()}
		if value[0] > 0 || value[1] > 0 {
			out[id] = value
		}
	}
	return out
}

func (t *trafficTotals) relaySnapshot() map[int][2]int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return snapshotTraffic(t.relays)
}

func (t *trafficTotals) relayUserSnapshot() map[int]map[int][2]int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[int]map[int][2]int64, len(t.relayUsers))
	for userID, nodes := range t.relayUsers {
		if values := snapshotTraffic(nodes); len(values) > 0 {
			out[userID] = values
		}
	}
	return out
}
