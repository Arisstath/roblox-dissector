package peer
import "net"

type ProxyWriter struct {
	Reader *PacketReader
	Writer *PacketWriter
	ErrorHandler func(error)
	OutputHandler func([]byte, *net.UDPAddr)
	ReliableHandler ReceiveHandler
	FullReliableHandler ReceiveHandler
	SimpleHandler ReceiveHandler
}

func NewProxyWriter() *ProxyWriter {
	proxy := &ProxyWriter{}

	writer := NewPacketWriter()
	writer.ErrorHandler = func(err error) {
		proxy.ErrorHandler(err)
	}
	writer.OutputHandler = func(payload []byte, addr *net.UDPAddr) {
		proxy.OutputHandler(payload, addr)
	}

	reader := NewPacketReader()

	reader.ErrorHandler = func(err error) {
		proxy.ErrorHandler(err)
	}
	reader.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		writer.WriteSimple(packetType, layers.Main.(RakNetPacket), &packet.Destination)
		proxy.SimpleHandler(packetType, packet, layers)
	}
	reader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		proxy.ReliableHandler(packetType, packet, layers)
	}
	reader.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		proxy.FullReliableHandler(packetType, packet, layers)
	}

	proxy.Reader = reader
	proxy.Writer = writer
	return proxy
}

func (w *ProxyWriter) Proxy(payload []byte, packet *UDPPacket) {
	w.Reader.ReadPacket(payload, packet)
}
