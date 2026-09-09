package session

import (
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/anytls/sing-anytls/pipe"
)

// Stream implements net.Conn
type Stream struct {
	id uint32

	sess *Session

	pipeR         *pipe.PipeReader
	pipeW         *pipe.PipeWriter
	writeDeadline pipe.PipeDeadline

	dieOnce sync.Once
	// YZ 兼容修复（2026-09-07）：关闭状态和回调可与 Read/Write、回调注册并发访问。
	stateMu sync.RWMutex
	dieHook func()
	dieErr  error

	reportOnce sync.Once
}

// newStream initiates a Stream struct
func newStream(id uint32, sess *Session) *Stream {
	s := new(Stream)
	s.id = id
	s.sess = sess
	s.pipeR, s.pipeW = pipe.Pipe()
	s.writeDeadline = pipe.MakePipeDeadline()
	return s
}

// Read implements net.Conn
func (s *Stream) Read(b []byte) (n int, err error) {
	n, err = s.pipeR.Read(b)
	if n == 0 {
		if closeErr := s.closeError(); closeErr != nil {
			err = closeErr
		}
	}
	return
}

// Write implements net.Conn
func (s *Stream) Write(b []byte) (n int, err error) {
	select {
	case <-s.writeDeadline.Wait():
		return 0, os.ErrDeadlineExceeded
	default:
	}
	if err := s.closeError(); err != nil {
		return 0, err
	}
	n, err = s.sess.writeDataFrame(s.id, b)
	return
}

// Close implements net.Conn
func (s *Stream) Close() error {
	return s.closeWithError(io.ErrClosedPipe)
}

// closeLocally only closes Stream and don't notify remote peer
func (s *Stream) closeLocally() {
	s.closeState(net.ErrClosed)
}

func (s *Stream) closeWithError(err error) error {
	if s.closeState(err) {
		return s.sess.streamClosed(s.id)
	}
	return s.closeError()
}

func (s *Stream) closeError() error {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.dieErr
}

// 注册可能晚于远端关闭；此时仍执行一次回调，避免会话清理丢失。
func (s *Stream) setDieHook(hook func()) {
	s.stateMu.Lock()
	closed := s.dieErr != nil
	if !closed {
		s.dieHook = hook
	}
	s.stateMu.Unlock()
	if closed && hook != nil {
		hook()
	}
}

func (s *Stream) closeState(err error) bool {
	var once bool
	var hook func()
	s.dieOnce.Do(func() {
		s.stateMu.Lock()
		s.dieErr = err
		hook = s.dieHook
		s.dieHook = nil
		s.pipeR.Close()
		s.stateMu.Unlock()
		once = true
	})
	// 回调可能继续关闭会话，必须在状态锁和 Once 之外调用。
	if hook != nil {
		hook()
	}
	return once
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	return s.pipeR.SetReadDeadline(t)
}

func (s *Stream) SetWriteDeadline(t time.Time) error {
	s.writeDeadline.Set(t)
	return nil
}

func (s *Stream) SetDeadline(t time.Time) error {
	s.SetWriteDeadline(t)
	return s.SetReadDeadline(t)
}

// LocalAddr satisfies net.Conn interface
func (s *Stream) LocalAddr() net.Addr {
	if ts, ok := s.sess.conn.(interface {
		LocalAddr() net.Addr
	}); ok {
		return ts.LocalAddr()
	}
	return nil
}

// RemoteAddr satisfies net.Conn interface
func (s *Stream) RemoteAddr() net.Addr {
	if ts, ok := s.sess.conn.(interface {
		RemoteAddr() net.Addr
	}); ok {
		return ts.RemoteAddr()
	}
	return nil
}

// HandshakeFailure should be called when Server fail to create outbound proxy
func (s *Stream) HandshakeFailure(err error) error {
	var once bool
	s.reportOnce.Do(func() {
		once = true
	})
	if once && err != nil && s.sess.peerVersion >= 2 {
		f := newFrame(cmdSYNACK, s.id)
		f.data = []byte(err.Error())
		if _, err := s.sess.writeControlFrame(f); err != nil {
			return err
		}
	}
	return nil
}

// HandshakeSuccess should be called when Server success to create outbound proxy
func (s *Stream) HandshakeSuccess() error {
	var once bool
	s.reportOnce.Do(func() {
		once = true
	})
	if once && s.sess.peerVersion >= 2 {
		if _, err := s.sess.writeControlFrame(newFrame(cmdSYNACK, s.id)); err != nil {
			return err
		}
	}
	return nil
}
