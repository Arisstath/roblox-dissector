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
	ReliabilityLayerHandler func(*UDPPacket, *ReliabilityLayer, *RakNetLayer)
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
	reader.ReliabilityLayerHandler = func(p *UDPPacket, re *ReliabilityLayer, ra *RakNetLayer) {
		proxy.ReliabilityLayerHandler(p, re, ra)
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

	clientHalf.ReliabilityLayerHandler = func(packet *UDPPacket, reliabilityLayer *ReliabilityLayer, rakNetLayer *RakNetLayer) {
		serverHalf.Writer.WriteReliableWithDN(
			reliabilityLayer,
			writer.ServerAddr,
			serverHalf.RotateDN(rakNetLayer.DatagramNumber),
		)
	}
	serverHalf.ReliabilityLayerHandler = func(packet *UDPPacket, reliabilityLayer *ReliabilityLayer, rakNetLayer *RakNetLayer) {
		clientHalf.Writer.WriteReliableWithDN(
			reliabilityLayer,
			writer.ClientAddr,
			clientHalf.RotateDN(rakNetLayer.DatagramNumber),
		)
	}

	clientHalf.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
	}
	serverHalf.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
	}
	clientHalf.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		// nop
	}
	serverHalf.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		// nop
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
			layer.ACKs = newacks
			serverHalf.Writer.WriteRakNet(layer, writer.ServerAddr)
		}
	}
	serverHalf.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		drop, newacks := clientHalf.RotateACKs(layer.ACKs)
		if !drop {
			layer.ACKs = newacks
			clientHalf.Writer.WriteRakNet(layer, writer.ClientAddr)
		} 
	}

	clientHalf.Writer.ToClient = false // doesn't write TO client!
	serverHalf.Writer.ToClient = true // writes TO client!

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
func (writer *ProxyWriter) InjectServer(packet RakNetPacket) {
	olddn := writer.ServerHalf.Writer.DatagramNumber
	writer.ServerHalf.Writer.WriteGeneric(
		writer.ServerHalf.Reader.Context,
		0x83,
		packet,
		0,
		writer.ServerAddr,
	) // Unreliable packets, might improve this sometime
	for i := olddn; i < writer.ServerHalf.Writer.DatagramNumber; i++ {
		println("adding fakepacket", i)
		writer.ServerHalf.FakePackets = append(writer.ServerHalf.FakePackets, i)
	}
}
