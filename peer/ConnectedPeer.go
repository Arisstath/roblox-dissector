package peer

import "sort"
import "net"

type ErrorHandler func(error)

// ConnectedPeer describes a connection to a peer
type ConnectedPeer struct {
	// Reader is a PacketReader reading packets sent by the peer.
	Reader *DefaultPacketReader
	// Writer is a PacketWriter writing packets to the peer.
	Writer *DefaultPacketWriter
	// All errors are dumped to ErrorHandler.
	ErrorHandler ErrorHandler
	// OutputHandler sends the data for packets to be written to the peer.
	// TODO: include all layer data in this packet as well?
	OutputHandler func([]byte)
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
	DestinationAddress      *net.UDPAddr

	mustACK []int
}

func (peer *ConnectedPeer) sendACKs() {
	if len(peer.mustACK) == 0 {
		return
	}
	acks := peer.mustACK
	peer.mustACK = []int{}
	var ackStructure []ACKRange
	sort.Ints(acks)

	for _, ack := range acks {
		if len(ackStructure) == 0 {
			ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
			continue
		}

		inserted := false
		for i, ackRange := range ackStructure {
			if int(ackRange.Max) == ack {
				inserted = true
				break
			}
			if int(ackRange.Max+1) == ack {
				ackStructure[i].Max++
				inserted = true
				break
			}
		}
		if inserted {
			continue
		}

		ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
	}

	result := &RakNetLayer{
		IsACK: true,
		ACKs:  ackStructure,
	}

	peer.Writer.WriteRakNet(result)
}

func NewConnectedPeer(context *CommunicationContext) *ConnectedPeer {
	proxy := &ConnectedPeer{}

	writer := NewPacketWriter()
	writer.ErrorHandler = func(err error) {
		proxy.ErrorHandler(err, nil)
	}
	writer.OutputHandler = func(payload []byte) {
		proxy.OutputHandler(payload)
	}

	reader := NewPacketReader()

	reader.ErrorHandler = func(err error, packet *UDPPacket) {
		proxy.ErrorHandler(err, packet)
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
	reader.ValContext = context
	writer.ValContext = context
	reader.ValCaches = new(Caches)
	writer.ValCaches = new(Caches)

	proxy.Reader = reader
	proxy.Writer = writer
	return proxy
}

// Receive sends packets to Reader.ReadPacket()
func (peer *ConnectedPeer) ReadPacket(payload []byte, packet *UDPPacket) {
	peer.Reader.ReadPacket(payload, packet)
}

// TODO: Perhaps different error handling for writing?
func (peer *ConnectedPeer) WriteSimple(packet RakNetPacket) {
	peer.Writer.WriteSimple(packet)
}
func (peer *ConnectedPeer) WritePacket(packet RakNetPacket) []byte {
	return peer.Writer.WriteGeneric(packet, RELIABLE_ORD)
}
func (peer *ConnectedPeer) WriteTimestamped(timestamp *Packet1BLayer, packet RakNetPacket) {
	peer.Writer.WriteTimestamped(timestamp, packet, UNRELIABLE)
}
