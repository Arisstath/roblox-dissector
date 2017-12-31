package peer
import "github.com/google/gopacket"
import "github.com/google/gopacket/layers"
import "bytes"
import "net"
import "github.com/gskartwii/go-bitstream"

// An UDPPacket describes a packet with a source and a destination, along with
// containing its contents internally.
type UDPPacket struct {
	stream *extendedReader
	Source net.UDPAddr
	Destination net.UDPAddr
}

func bufferToStream(buffer []byte) *extendedReader {
	return &extendedReader{bitstream.NewReader(bytes.NewReader(buffer))}
}

func UDPPacketFromGoPacket(packet gopacket.Packet) *UDPPacket {
	if packet.Layer(layers.LayerTypeIPv4) == nil {
		return nil
	}

	ret := &UDPPacket{}
	ret.stream = bufferToStream(packet.ApplicationLayer().Payload())
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

// Constructs a UDPPacket from a buffer of bytes
func UDPPacketFromBytes(buf []byte) *UDPPacket {
	return &UDPPacket{stream: bufferToStream(buf)}
}
