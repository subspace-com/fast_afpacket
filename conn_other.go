//go:build !linux

package fastafpacket

import (
	"fmt"
	"net"
	"runtime"

	"golang.org/x/net/bpf"
)

var errNotSupported = fmt.Errorf("packet: not supported on %s", runtime.GOOS)

func (c *Conn) recvTimestamps(_ []byte, _ int) (int, net.Addr, SocketTimestamps, error) {
	return 0, nil, SocketTimestamps{}, errNotSupported
}
func (c *Conn) readFrom(_ []byte) (int, net.Addr, error)  { return 0, nil, errNotSupported }
func (c *Conn) writeTo(_ []byte, _ net.Addr) (int, error) { return 0, errNotSupported }
func (c *Conn) setBPF(_ []bpf.RawInstruction) error       { return errNotSupported }

func Listen(iface *net.Interface, socketType int, socketProtocol int, config *Config) (*Conn, error) {
	return nil, errNotSupported
}
