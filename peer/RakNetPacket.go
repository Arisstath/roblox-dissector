package peer
import "github.com/google/gopacket"
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
	ret := &RakNetPacket{}
	ret.Stream = BufferToStream(packet.ApplicationLayer().Payload())
	ret.Source = net.UDPAddr{
		packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).SrcIP,
		packet.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort,
	}
	ret.Destination = net.UDPAddr{
		packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).DstIP,
		packet.Layer(layers.LayerTypeUDP).(*layers.UDP).DstPort,
	}
}
