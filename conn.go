package fastafpacket

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/socket"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

type Queue int

const (
	// SocketOptTimestamping is an alias for unix.SO_TIMESTAMPING
	SocketOptTimestamping = 0x25

	// MsgErrQueue is an alias for unix.MSG_ERRQUEUE
	MsgErrQueue = 0x2000
)

var (
	unixEpoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type Addr struct {
	HardwareAddr net.HardwareAddr
}

func (a *Addr) Network() string {
	return "packet"
}

func (a *Addr) String() string {
	return a.HardwareAddr.String()
}

type Config struct {
	Filter []bpf.RawInstruction
}

type Conn struct {
	iface    *net.Interface
	addr     net.Addr
	protocol int
	s        *socket.Conn
	r        *socket.Conn
}

func (c *Conn) Close() error {
	if err := c.s.Close(); err != nil {
		return err
	}
	if err := c.r.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return c.addr
}

func (c *Conn) RecvTxTimestamps(b []byte) (int, net.Addr, SocketTimestamps, error) {
	return c.recvTimestamps(b, MsgErrQueue)
}

func (c *Conn) RecvRxTimestamps(b []byte) (int, net.Addr, SocketTimestamps, error) {
	return c.recvTimestamps(b, 0)
}

func (c *Conn) ReadFrom(b []byte) (int, net.Addr, error) {
	return c.readFrom(b)
}

func (c *Conn) WriteTo(b []byte, addr net.Addr) (int, error) {
	return c.writeTo(b, addr)
}

func (c *Conn) SetDeadline(t time.Time) error {
	if err := c.s.SetDeadline(t); err != nil {
		return err
	}
	if err := c.r.SetDeadline(t); err != nil {
		return err
	}

	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	if err := c.r.SetReadDeadline(t); err != nil {
		return err
	}

	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	if err := c.s.SetWriteDeadline(t); err != nil {
		return err
	}

	return nil
}

func (c *Conn) SetBPF(filter []bpf.RawInstruction) error {
	return c.setBPF(filter)
}

type SocketTimestamps struct {
	Software time.Time
	Hardware time.Time
}

// ParseSocketTimestamps parses the timestamp information from the control message
// it will also convert Unix epoch times to Go's zero value time to be able
// to use .IsZero() on timestamps that have a default value or are not available.
// https://www.kernel.org/doc/html/v5.14/networking/timestamping.html#scm-timestamping-records
func ParseSocketTimestamps(msg unix.SocketControlMessage) (SocketTimestamps, error) {
	var ts SocketTimestamps

	if msg.Header.Level != unix.SOL_SOCKET && msg.Header.Type != SocketOptTimestamping {
		return ts, fmt.Errorf("no timestamp control messages")
	}

	ts.Software = time.Unix(int64(binary.LittleEndian.Uint64(msg.Data[0:])), int64(binary.LittleEndian.Uint64(msg.Data[8:])))
	ts.Hardware = time.Unix(int64(binary.LittleEndian.Uint64(msg.Data[32:])), int64(binary.LittleEndian.Uint64(msg.Data[40:])))

	// convert Unix epoch to Go's zero value of time

	if ts.Software.Equal(unixEpoch) {
		ts.Software = time.Time{}
	}

	if ts.Hardware.Equal(unixEpoch) {
		ts.Hardware = time.Time{}
	}

	return ts, nil
}
