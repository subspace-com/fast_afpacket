package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	fastafpacket "github.com/subspace-com/fast_afpacket"
	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
	"inet.af/netaddr"
)

var (
	ProbeBypassPrefix = []byte{0x93, 0x72, 0xe7, 0x59, 0xc7, 0xb5, 0xaa, 0x73}

	ProbeID uint64

	Drops            uint32
	Packets          uint32
	FreezeQueueCount uint32
)

func main() {
	logrus.SetFormatter(logrus.StandardLogger().Formatter)

	ifname := flag.String("iface-name", "", "interface name to bind to")
	srcmac := flag.String("src-mac", "", "source MAC to use for packets")
	srcaddr := flag.String("src-addr", "", "source IPv4 address to use for packets")
	srcport := flag.Int("src-port", 60000, "source port to use for packets")
	dstmac := flag.String("dst-mac", "", "destination MAC to use for packets")
	dstaddr := flag.String("dst-addr", "", "destination IPv4 address to use for packets")
	dstport := flag.Int("dst-port", 61000, "destination port to use for packets")
	flag.Parse()

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		logrus.Fatal(err)
	}

	smac, _ := net.ParseMAC(*srcmac)
	saddr := netaddr.MustParseIP(*srcaddr)
	dmac, _ := net.ParseMAC(*dstmac)
	daddr := netaddr.MustParseIP(*dstaddr)

	filter := fmt.Sprintf("dst port %v and src portrange %v-%v and %v and dst %v", *dstport, *srcport, *srcport+5, "udp", saddr)

	pcapBpf, err := pcap.CompileBPFFilter(layers.LinkTypeEthernet, 1024, filter)
	if err != nil {
		logrus.Fatal(err)
	}

	allInstructions := []bpf.RawInstruction{}
	for _, ins := range pcapBpf {
		oneInstruction := bpf.RawInstruction{
			Op: ins.Code,
			Jt: ins.Jt,
			Jf: ins.Jf,
			K:  ins.K,
		}
		allInstructions = append(allInstructions, oneInstruction)
	}

	config := fastafpacket.Config{
		DualConn: true,

		Filter: allInstructions,
	}

	conn, err := fastafpacket.Listen(iface, unix.SOCK_RAW, unix.ETH_P_ALL, &config)
	if err != nil {
		logrus.Fatal(err)
	}
	{
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				packet, err := encodePacket(smac, saddr.IPAddr().IP, *srcport, dmac, daddr.IPAddr().IP, *dstport)
				if err != nil {
					logrus.Fatalf("tx new packet: %v", err)
				}

				_, err = conn.WriteTo(packet, &fastafpacket.Addr{HardwareAddr: dmac})
				if err != nil {
					logrus.Fatalf("tx send to: %v", err)
				}
			}
		}()

		go func() {
			for {
				packet := make([]byte, 1024)

				_, _, ts, err := conn.RecvTxTimestamps(packet)
				if err != nil {
					logrus.Fatalf("tx receive msg: %v", err)
				}

				logrus.WithFields(logrus.Fields{
					"probe":       ProbeID,
					"hardware_ns": ts.Hardware.UnixNano(),
					"software_ns": ts.Software.UnixNano(),
					"hardware":    ts.Hardware.UTC().Format(time.RFC3339Nano),
					"software":    ts.Software.UTC().Format(time.RFC3339Nano),
				}).Println("TX Recvmsg")
			}
		}()
	}

	{
		go func() {
			for {
				packet := make([]byte, 1024)

				n, _, ts, err := conn.RecvRxTimestamps(packet)
				if err != nil {
					logrus.Fatalf("rx receive msg: %v", err)
				}

				recvTime := time.Now()

				payload := decodePacket(packet[:n])

				probeID := binary.LittleEndian.Uint64(payload[8:])

				logrus.WithFields(logrus.Fields{
					"probe":        probeID,
					"userspace_ns": recvTime.UnixNano(),
					"hardware_ns":  ts.Hardware.UnixNano(),
					"software_ns":  ts.Software.UnixNano(),
					"userspace":    recvTime.UTC().Format(time.RFC3339Nano),
					"hardware":     ts.Hardware.UTC().Format(time.RFC3339Nano),
					"software":     ts.Software.UTC().Format(time.RFC3339Nano),
					"delay":        ts.Software.Sub(recvTime),
				}).Println("RX Recvmsg")
			}
		}()
	}

	{
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				stats, err := conn.Stats()
				if err != nil {
					logrus.Warnln(err)
				}

				Drops += stats.Drops
				Packets += stats.Packets
				FreezeQueueCount += stats.FreezeQueueCount

				logrus.WithFields(logrus.Fields{
					"drops":         Drops,
					"packets":       Packets,
					"queue_freezes": FreezeQueueCount,
				}).Println("RX Stats")
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func encodePacket(smac net.HardwareAddr, saddr net.IP, sport int, dmac net.HardwareAddr, daddr net.IP, dport int) ([]byte, error) {
	ethlayer := &layers.Ethernet{
		SrcMAC:       smac,
		DstMAC:       dmac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	var iplayer gopacket.Layer = &layers.IPv4{
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    saddr,
		DstIP:    daddr,
	}

	udplayer := &layers.UDP{
		SrcPort: layers.UDPPort(sport),
		DstPort: layers.UDPPort(dport),
	}

	err := udplayer.SetNetworkLayerForChecksum(iplayer.(gopacket.NetworkLayer))
	if err != nil {
		return nil, err
	}

	pkt := gopacket.NewSerializeBuffer()

	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	payload := make([]byte, 16)
	copy(payload, ProbeBypassPrefix)

	ProbeID++
	binary.LittleEndian.PutUint64(payload[8:], ProbeID)

	err = gopacket.SerializeLayers(pkt, options, ethlayer, iplayer.(gopacket.SerializableLayer), udplayer, gopacket.Payload(payload))
	if err != nil {
		return nil, err
	}

	return pkt.Bytes(), nil
}

func decodePacket(packet []byte) []byte {
	pkt := gopacket.NewPacket(packet, layers.LayerTypeEthernet, gopacket.DecodeOptions{
		Lazy:   true,
		NoCopy: true,
	})

	app := pkt.ApplicationLayer()

	return app.Payload()
}
