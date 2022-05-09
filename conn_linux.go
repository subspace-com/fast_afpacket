//go:build linux

package fastafpacket

import (
	"errors"
	"fmt"
	"net"
	"os"
	"unsafe"

	"github.com/mdlayher/socket"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

func (c *Conn) recvTimestamps(b []byte, flag int) (int, net.Addr, SocketTimestamps, error) {
	oob := make([]byte, 1024)

	conn := c.r
	if flag == MsgErrQueue {
		conn = c.s
		fmt.Println("sender conn")
	}

	n, oobn, _, sa, err := conn.Recvmsg(b, oob, flag)
	if err != nil {
		return 0, nil, SocketTimestamps{}, err
	}

	ts, err := parseTimestamps(oob[:oobn])
	if err != nil {
		return 0, nil, SocketTimestamps{}, err
	}

	return n, sockaddrToAddr(sa), ts, nil
}

func (c *Conn) readFrom(b []byte) (int, net.Addr, error) {
	n, sa, err := c.r.Recvfrom(b, 0)
	if err != nil {
		return 0, nil, err
	}

	return n, sockaddrToAddr(sa), nil
}

func (c *Conn) writeTo(b []byte, addr net.Addr) (int, error) {
	sa, err := addrToSockaddr(addr, c.iface.Index, c.protocol)
	if err != nil {
		return 0, err
	}

	err = c.s.Sendto(b, sa, 0)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *Conn) setBPF(filter []bpf.RawInstruction) error {
	return c.r.SetBPF(filter)
}

func parseTimestamps(cmsg []byte) (SocketTimestamps, error) {
	msgs, err := unix.ParseSocketControlMessage(cmsg)
	if err != nil {
		return SocketTimestamps{}, err
	}

	ts, err := ParseSocketTimestamps(msgs[0])
	if err != nil {
		return SocketTimestamps{}, err
	}

	return ts, nil
}

func sockaddrToAddr(sa unix.Sockaddr) *Addr {
	if sa == nil {
		return nil
	}

	if a, ok := sa.(*unix.SockaddrLinklayer); ok {
		return &Addr{
			HardwareAddr: net.HardwareAddr(a.Addr[:a.Halen]),
		}
	}

	return nil
}

func addrToSockaddr(addr net.Addr, ifIndex int, protocol int) (unix.Sockaddr, error) {
	a, ok := addr.(*Addr)
	if !ok || a.HardwareAddr == nil {
		return nil, errors.New("invalid net.Addr")
	}

	sa := unix.SockaddrLinklayer{
		Ifindex:  ifIndex,
		Protocol: uint16(protocol),
	}

	if len(a.HardwareAddr) > len(sa.Addr) {
		return nil, errors.New("invalid net.Addr")
	}

	sa.Halen = uint8(len(a.HardwareAddr))
	copy(sa.Addr[:], a.HardwareAddr)

	return &sa, nil
}

func Listen(iface *net.Interface, socketType int, socketProtocol int, config *Config) (*Conn, error) {
	if err := enableHardwareTimestamping(iface); err != nil {
		return nil, fmt.Errorf("failed to enable hardware timestamping: %v", err)
	}

	sender, err := socket.Socket(unix.AF_PACKET, socketType, 0, "fastafpacket_sender", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create conn: %v", err)
	}

	if err := enableSocketOptions(sender, iface); err != nil {
		return nil, fmt.Errorf("failed to enable socket timestamping: %v", err)
	}

	receiver, err := socket.Socket(unix.AF_PACKET, socketType, 0, "fastafpacket_receiver", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create conn: %v", err)
	}

	if err := enableSocketOptions(receiver, iface); err != nil {
		return nil, fmt.Errorf("failed to enable socket timestamping: %v", err)
	}

	if config.Filter != nil {
		err := receiver.SetBPF(config.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to apply BPF filter: %v", err)
		}
	}

	err = receiver.Bind(&unix.SockaddrLinklayer{Protocol: htons(uint16(socketProtocol))})
	if err != nil {
		return nil, fmt.Errorf("failed to bind: %v", err)
	}

	return &Conn{
		iface:    iface,
		addr:     &Addr{HardwareAddr: iface.HardwareAddr},
		protocol: socketProtocol,
		s:        sender,
		r:        receiver,
	}, nil
}

func enableSocketOptions(conn *socket.Conn, iface *net.Interface) error {
	tsInfo, err := ethtoolTimstampingInfo(iface.Name)
	if err != nil {
		return err
	}

	if err := conn.SetsockoptInt(unix.SOL_SOCKET, unix.PACKET_VERSION, unix.TPACKET_V3); err != nil {
		return fmt.Errorf("failed to enable packet statistics: %v", err)
	}

	if err := conn.SetsockoptInt(unix.SOL_SOCKET, unix.SO_TIMESTAMPING, int(tsInfo.soTimestamping)); err != nil {
		return fmt.Errorf("could not set SO_TIMESTAMPING: %v", err)
	}

	if err := conn.SetsockoptInt(unix.SOL_PACKET, unix.PACKET_TIMESTAMP, int(tsInfo.soTimestamping)); err != nil {
		return fmt.Errorf("could not set PACKET_TIMESTAMP: %v", err)
	}

	return nil
}

func enableHardwareTimestamping(iface *net.Interface) error {
	tsInfo, err := ethtoolTimstampingInfo(iface.Name)
	if err != nil {
		return err
	}

	const (
		HWTSTAMP_FILTER_ALL = 1
		HWTSTAMP_TX_ON      = 1
	)

	type hwtstampConfig struct {
		flags    uint32
		txType   uint32
		rxFilter uint32
	}

	config := hwtstampConfig{}

	// ideally we want HWTSTAMP_TX_ON and HWTSTAMP_FILTER_ALL
	if tsInfo.txTypes&(1<<HWTSTAMP_TX_ON) != 0 {
		config.txType = HWTSTAMP_TX_ON
	}

	if tsInfo.rxFilters&(1<<HWTSTAMP_FILTER_ALL) != 0 {
		config.rxFilter = HWTSTAMP_FILTER_ALL
	}

	// if either option was updated, set them on the device
	if config.txType == HWTSTAMP_TX_ON || config.rxFilter == HWTSTAMP_FILTER_ALL {
		if err := ioctlInterfaceRequest(iface.Name, unix.SIOCSHWTSTAMP, uintptr(unsafe.Pointer(&config))); err != nil {
			return os.NewSyscallError(fmt.Sprintf("ioctl_siocshwtstamp_%v", iface.Name), err)
		}
	}

	return nil
}

type timstampingInfo struct {
	cmd            uint32
	soTimestamping uint32
	phcIndex       int32
	txTypes        uint32
	txReserved     [3]uint32
	rxFilters      uint32
	rxReserved     [3]uint32
}

func ethtoolTimstampingInfo(ifaceName string) (*timstampingInfo, error) {
	tsinfo := timstampingInfo{
		cmd: unix.ETHTOOL_GET_TS_INFO,
	}

	if err := ioctlInterfaceRequest(ifaceName, unix.SIOCETHTOOL, uintptr(unsafe.Pointer(&tsinfo))); err != nil {
		return nil, os.NewSyscallError(fmt.Sprintf("ioctl_ethtool_get_ts_info_%v", ifaceName), err)
	}

	return &tsinfo, nil
}

func ioctlInterfaceRequest(ifaceName string, key int, dataPtr uintptr) error {
	var name [unix.IFNAMSIZ]byte
	copy(name[:], []byte(ifaceName))

	type ifreq struct {
		ifr_name [unix.IFNAMSIZ]byte
		ifr_data uintptr
	}

	ifr := ifreq{
		ifr_name: name,
		ifr_data: dataPtr,
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		return err
	}

	defer unix.Close(fd)

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(key), uintptr(unsafe.Pointer(&ifr))); errno != 0 {
		return fmt.Errorf("errno %v", errno)
	}

	return nil
}

func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
