package peer
import "net"

type ConnectedPeer struct {
	Reader *PacketReader
	Writer *PacketWriter
	ErrorHandler func(error)
	OutputHandler func([]byte, *net.UDPAddr)
	ReliableHandler ReceiveHandler
	FullReliableHandler ReceiveHandler
	SimpleHandler ReceiveHandler
	ACKHandler func(*UDPPacket, *RakNetLayer)
}

func NewConnectedPeer(context *CommunicationContext) *ConnectedPeer {
	proxy := &ConnectedPeer{}

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
		proxy.SimpleHandler(packetType, packet, layers)
	}
	reader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		proxy.ReliableHandler(packetType, packet, layers)
	}
	reader.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		proxy.FullReliableHandler(packetType, packet, layers)
	}
	reader.ACKHandler = func(p *UDPPacket, r *RakNetLayer) {
		proxy.ACKHandler(p, r)
	}
	reader.Context = context

	proxy.Reader = reader
	proxy.Writer = writer
	return proxy
}

func (w *ConnectedPeer) Receive(payload []byte, packet *UDPPacket) {
	w.Reader.ReadPacket(payload, packet)
}

type ProxyHalf struct {
	*ConnectedPeer
	FakePackets []uint32
}

func (w *ProxyHalf) RotateDN(old uint32) uint32 {
	for i := len(w.FakePackets) - 1; i >= 0; i-- {
		fakepacket := w.FakePackets[i]
		if old >= fakepacket {
			old++
		}
	}
	return old
}
func (w *ProxyHalf) RotateACK(ack ACKRange) (bool, ACKRange) {
	fakepackets := w.FakePackets
	for i := len(fakepackets) - 1; i >= 0; i-- {
		fakepacket := fakepackets[i]
		if ack.Max >= fakepacket {
			ack.Max--
		}
		if ack.Min > fakepacket {
			ack.Min--
		}
	}
	return ack.Min > ack.Max, ack
}
func (w *ProxyHalf) RotateACKs(acks []ACKRange) (bool, []ACKRange) {
	newacks := make([]ACKRange, 0, len(acks))
	for i := 0; i < len(acks); i++ {
		dropthis, newack := w.RotateACK(acks[i])
		if !dropthis {
			newacks = append(newacks, newack)
		}
	}
	return len(newacks) == 0, newacks
}

type ProxyWriter struct {
	ClientHalf *ProxyHalf
	ServerHalf *ProxyHalf
	ClientAddr *net.UDPAddr
	ServerAddr *net.UDPAddr
	OutputHandler func([]byte, *net.UDPAddr)
}

func NewProxyWriter(context *CommunicationContext) *ProxyWriter {
	writer := &ProxyWriter{}
	clientHalf := &ProxyHalf{NewConnectedPeer(context), nil}
	serverHalf := &ProxyHalf{NewConnectedPeer(context), nil}

	clientHalf.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		println("client simple", packetType)
		if packetType == 5 {
			println("recv 5, protocol type", layers.Main.(*Packet05Layer).ProtocolVersion)
		}
		serverHalf.Writer.WriteSimple(packetType, layers.Main.(RakNetPacket), writer.ServerAddr)
	}
	serverHalf.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		println("server simple", packetType)
		clientHalf.Writer.WriteSimple(packetType, layers.Main.(RakNetPacket), writer.ClientAddr)
	}

	clientHalf.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		serverHalf.Writer.WriteReliableWithDN(
			&ReliabilityLayer{
				[]*ReliablePacket{layers.Reliability},
			},
			writer.ServerAddr,
			clientHalf.RotateDN(layers.RakNet.DatagramNumber),
		)
	}
	serverHalf.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		clientHalf.Writer.WriteReliableWithDN(
			&ReliabilityLayer{
				[]*ReliablePacket{layers.Reliability},
			},
			writer.ClientAddr,
			serverHalf.RotateDN(layers.RakNet.DatagramNumber),
		)
	}

	clientHalf.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		println("client recv", packetType)
	}
	serverHalf.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		println("server recv", packetType)
	}

	clientHalf.ErrorHandler = func(err error) {
		println("clienthalf err:", err.Error())
	}
	serverHalf.ErrorHandler = func(err error) {
		println("serverhalf err:", err.Error())
	}

	clientHalf.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		drop, newacks := serverHalf.RotateACKs(layer.ACKs)
		if !drop {
			println("client rotated acks:", newacks[0].Min, newacks[0].Max)
			layer.ACKs = newacks
			serverHalf.Writer.WriteRakNet(layer, writer.ServerAddr)
		} else {
			println("dropping ack")
		}
	}
	serverHalf.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		drop, newacks := clientHalf.RotateACKs(layer.ACKs)
		if !drop {
			println("server rotated acks:", newacks[0].Min, newacks[0].Max)
			layer.ACKs = newacks
			clientHalf.Writer.WriteRakNet(layer, writer.ClientAddr)
		} else {
			println("dropping ack")
		}
	}

	writer.ClientHalf = clientHalf
	writer.ServerHalf = serverHalf
	return writer
}

func (writer *ProxyWriter) ProxyClient(payload []byte, packet *UDPPacket) {
	writer.ClientHalf.Reader.ReadPacket(payload, packet)
}
func (writer *ProxyWriter) ProxyServer(payload []byte, packet *UDPPacket) {
	writer.ServerHalf.Reader.ReadPacket(payload, packet)
}
