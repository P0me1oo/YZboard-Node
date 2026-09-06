package session

import (
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type discardConn struct{ net.Conn }

func (discardConn) Write(p []byte) (int, error) {
	runtime.Gosched()
	return len(p), nil
}

func (discardConn) SetWriteDeadline(time.Time) error { return nil }

func newTestStream() *Stream {
	session := NewServerSession(discardConn{}, nil, nil, nil)
	return newStream(1, session)
}

// Read/Write 必须能与远端关闭并发，且关闭后返回稳定的错误。
func TestStreamConcurrentCloseAndIO(t *testing.T) {
	for range 128 {
		stream := newTestStream()
		start := make(chan struct{})
		var workers sync.WaitGroup
		workers.Add(3)
		go func() {
			defer workers.Done()
			<-start
			for range 32 {
				if _, err := stream.Write([]byte{1}); err != nil {
					return
				}
			}
		}()
		go func() {
			defer workers.Done()
			<-start
			if _, err := stream.Read(make([]byte, 1)); !errors.Is(err, net.ErrClosed) {
				t.Errorf("关闭后的读取错误不正确: %v", err)
			}
		}()
		go func() {
			defer workers.Done()
			<-start
			stream.closeLocally()
		}()
		close(start)
		workers.Wait()
		if _, err := stream.Write([]byte{1}); !errors.Is(err, net.ErrClosed) {
			t.Fatalf("关闭后的写入错误不正确: %v", err)
		}
	}
}

func TestStreamCloseHookRegisteredAfterClose(t *testing.T) {
	for _, test := range []struct {
		name  string
		close func(*Stream)
	}{
		{"remote", (*Stream).closeLocally},
		{"local", func(stream *Stream) { stream.Close() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			stream := newTestStream()
			test.close(stream)
			var calls atomic.Int32
			stream.setDieHook(func() { calls.Add(1) })
			test.close(stream)
			if calls.Load() != 1 {
				t.Fatalf("延迟注册的关闭回调执行次数 = %d，预期 1", calls.Load())
			}
		})
	}
}

func TestStreamCloseHookConcurrentRegistration(t *testing.T) {
	for range 128 {
		stream := newTestStream()
		var calls atomic.Int32
		var workers sync.WaitGroup
		start := make(chan struct{})
		for _, action := range []func(){
			func() { stream.setDieHook(func() { calls.Add(1); stream.Close() }) },
			func() { stream.closeWithError(io.ErrClosedPipe) },
			stream.closeLocally,
		} {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				action()
			}()
		}
		close(start)
		workers.Wait()
		if calls.Load() != 1 {
			t.Fatalf("并发关闭时回调执行次数 = %d，预期 1", calls.Load())
		}
	}
}
