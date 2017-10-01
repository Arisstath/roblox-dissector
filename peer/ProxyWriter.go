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
	reader := NewPacketReader()
	reader.ErrorHandler = func(err error) {
		proxy.ErrorHandler(err)
	}
	reader.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		writer.WriteSimple(packetType, layers.Simple, packet.Destination)
		proxy.SimpleHandler(packetType, packet, layers)
	}
	reader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		writer.WriteReliable(layers.Reliability, packet.Destination)
		proxy.ReliableHandler(packetType, packet, layers)
	}
	reader.FullReliableHandler = func(packetType byte, packet *UDPAddr, layers *PacketReader) {
		proxy.FullReliableHandler(packetType, packet, layers)
	}

	writer := NewPacketWriter()
	writer.ErrorHandler = func(err error) {
		proxy.ErrorHandler(err)
	}
	writer.OutputHandler = func(payload []byte, addr *net.UDPAddr) {
		proxy.OutputHandler(payload, addr)
	}

	proxy.Reader = reader
	proxy.Writer = writer
	return proxy
}

func (w *ProxyWriter) Proxy(payload []byte, packet *UDPPacket) {
	w.Reader.ReadPacket(payload, packet)
}
