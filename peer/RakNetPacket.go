package peer
import "github.com/google/gopacket"
import "github.com/google/gopacket/layers"
import "bytes"
import "net"
import "github.com/gskartwii/go-bitstream"

type UDPPacket struct {
	Stream *ExtendedReader
	Source net.UDPAddr
	Destination net.UDPAddr
}

func BufferToStream(buffer []byte) *ExtendedReader {
	return &ExtendedReader{bitstream.NewReader(bytes.NewReader(buffer))}
}

func UDPPacketFromGoPacket(packet gopacket.Packet) *UDPPacket {
	if packet.Layer(layers.LayerTypeIPv4) == nil {
		return nil
	}

	ret := &UDPPacket{}
	ret.Stream = BufferToStream(packet.ApplicationLayer().Payload())
	ret.Source = net.UDPAddr{
		packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).SrcIP,
		int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort),
		"udp",
	}
	ret.Destination = net.UDPAddr{
		packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).DstIP,
		int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).DstPort),
		"udp",
	}

	return ret
}
