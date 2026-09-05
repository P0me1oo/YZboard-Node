package singbox

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/sagernet/sing/common/buf"
	singBufio "github.com/sagernet/sing/common/bufio"
	singM "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type counterTestPacketConn struct {
	N.PacketConn
	reads [][]byte
}

func (c *counterTestPacketConn) ReadPacket(buffer *buf.Buffer) (singM.Socksaddr, error) {
	if len(c.reads) == 0 {
		return singM.Socksaddr{}, io.EOF
	}
	_, err := buffer.Write(c.reads[0])
	c.reads = c.reads[1:]
	return singM.ParseSocksaddr("127.0.0.1:80"), err
}

func (c *counterTestPacketConn) WritePacket(buffer *buf.Buffer, _ singM.Socksaddr) error {
	buffer.Release()
	return nil
}

func (c *counterTestPacketConn) Close() error { return nil }

func TestConnTrackerPacketCopyPreservesCountersAndDirection(t *testing.T) {
	tracker := NewConnTracker(0)
	tracker.SetUserMap(map[string]int{"packet-test-user": 1})
	base := &counterTestPacketConn{reads: [][]byte{[]byte("request")}}
	wrapped := tracker.RoutedPacketConnection(context.Background(), base,
		testInboundContext("packet-test-user", "127.0.0.1"), nil, nil)
	if _, err := singBufio.CopyPacket(&counterTestPacketConn{}, wrapped); !errors.Is(err, io.EOF) {
		t.Fatalf("上传复制失败: %v", err)
	}
	if _, err := singBufio.CopyPacket(wrapped, &counterTestPacketConn{reads: [][]byte{[]byte("ok")}}); !errors.Is(err, io.EOF) {
		t.Fatalf("下载复制失败: %v", err)
	}
	traffic, _, _ := tracker.GetUserTraffic()
	if got := traffic[1]; got != [2]int64{7, 2} {
		t.Fatalf("UDP 上传和下载应分别统计一次: got=%v, want=[7 2]", got)
	}
	_ = wrapped.Close()
	_ = wrapped.Close()
	_, _, count := tracker.GetUserTraffic()
	if count != 0 {
		t.Fatalf("重复关闭不能残留连接或重复扣减: %d", count)
	}
}
