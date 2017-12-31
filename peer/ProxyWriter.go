package peer
import "net"

// ConnectedPeer describes a connection to a peer
type ConnectedPeer struct {
	// Reader is a PacketReader reading packets sent by the peer.
	Reader *PacketReader
	// Writer is a PacketWriter writing packets to the peer.
	Writer *PacketWriter
	// All errors are dumped to ErrorHandler.
	ErrorHandler func(error)
	// OutputHandler sends the data for packets to be written to the peer.
	OutputHandler func([]byte, *net.UDPAddr)
	// Callback for simple pre-connection packets.
	SimpleHandler ReceiveHandler
	// Callback for ReliabilityLayer subpackets. This callback is invoked for every
	// split of every packets, possible duplicates, etc.
	ReliableHandler ReceiveHandler
	// Callback for generic packets (anything that is sent when a connection has been
	// established. You definitely want to bind to this.
	FullReliableHandler ReceiveHandler
	// Callback for ACKs and NAKs. Not very useful if you're just parsing packets.
	// However, you want to bind to this if you are writing a peer.
	ACKHandler func(*UDPPacket, *RakNetLayer)
	// Callback for ReliabilityLayer full packets. This callback is invoked for every
	// real ReliabilityLayer.
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

// Receive sends packets to Reader.ReadPacket()
func (w *ConnectedPeer) Receive(payload []byte, packet *UDPPacket) {
	w.Reader.ReadPacket(payload, packet)
}

// ProxyHalf describes a proxy connection to a connected peer.
type ProxyHalf struct {
	*ConnectedPeer
	fakePackets []uint32
}

func (w *ProxyHalf) rotateDN(old uint32) uint32 {
	for i := len(w.fakePackets) - 1; i >= 0; i-- {
		fakepacket := w.fakePackets[i]
		if old >= fakepacket {
			old++
		}
	}
	return old
}
func (w *ProxyHalf) rotateACK(ack ACKRange) (bool, ACKRange) {
	fakepackets := w.fakePackets
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
func (w *ProxyHalf) rotateACKs(acks []ACKRange) (bool, []ACKRange) {
	newacks := make([]ACKRange, 0, len(acks))
	for i := 0; i < len(acks); i++ {
		dropthis, newack := w.rotateACK(acks[i])
		if !dropthis {
			newacks = append(newacks, newack)
		}
	}
	return len(newacks) == 0, newacks
}

// ProxyWriter describes a proxy that connects two peers.
// ProxyWriters have injection capabilities.
type ProxyWriter struct {
	ClientHalf *ProxyHalf
	ServerHalf *ProxyHalf
	ClientAddr *net.UDPAddr
	ServerAddr *net.UDPAddr
	// When data should be sent to a peer, OutputHandler is called.
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
		serverHalf.Writer.writeReliableWithDN(
			reliabilityLayer,
			writer.ServerAddr,
			serverHalf.rotateDN(rakNetLayer.DatagramNumber),
		)
	}
	serverHalf.ReliabilityLayerHandler = func(packet *UDPPacket, reliabilityLayer *ReliabilityLayer, rakNetLayer *RakNetLayer) {
		clientHalf.Writer.writeReliableWithDN(
			reliabilityLayer,
			writer.ClientAddr,
			clientHalf.rotateDN(rakNetLayer.DatagramNumber),
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
		drop, newacks := serverHalf.rotateACKs(layer.ACKs)
		if !drop {
			layer.ACKs = newacks
			serverHalf.Writer.WriteRakNet(layer, writer.ServerAddr)
		}
	}
	serverHalf.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		drop, newacks := clientHalf.rotateACKs(layer.ACKs)
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

// ProxyClient should be called when the client sends a packet.
func (writer *ProxyWriter) ProxyClient(payload []byte, packet *UDPPacket) {
	writer.ClientHalf.Reader.ReadPacket(payload, packet)
}
// ProxyServer should be called when the server sends a packet.
func (writer *ProxyWriter) ProxyServer(payload []byte, packet *UDPPacket) {
	writer.ServerHalf.Reader.ReadPacket(payload, packet)
}
// (WIP) InjectServer should be called when an injected packet should be sent to
// the server.
func (writer *ProxyWriter) InjectServer(packet RakNetPacket) {
	olddn := writer.ServerHalf.Writer.datagramNumber
	writer.ServerHalf.Writer.WriteGeneric(
		writer.ServerHalf.Reader.Context,
		0x83,
		packet,
		0,
		writer.ServerAddr,
	) // Unreliable packets, might improve this sometime
	for i := olddn; i < writer.ServerHalf.Writer.datagramNumber; i++ {
		println("adding fakepacket", i)
		writer.ServerHalf.fakePackets = append(writer.ServerHalf.fakePackets, i)
	}
}
