package peer

import "sort"
import "net"

type ErrorHandler func(error)

// ConnectedPeer describes a connection to a peer
type ConnectedPeer struct {
	// Reader is a PacketReader reading packets sent by the peer.
	*DefaultPacketReader
	// Writer is a PacketWriter writing packets to the peer.
	*DefaultPacketWriter
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
	ACKHandler func(layers *PacketLayers)
	// Callback for ReliabilityLayer full packets. This callback is invoked for every
	// real ReliabilityLayer.
	ReliabilityLayerHandler func(layers *PacketLayers)
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

	peer.WriteRakNet(result)
}

func NewConnectedPeer(context *CommunicationContext) *ConnectedPeer {
	myPeer := &ConnectedPeer{}

    reader := NewPacketReader()
	writer := NewPacketWriter()

    reader.SetContext(context)
    writer.SetContext(context)
    reader.SetCaches(new(Caches))
    writer.SetCaches(new(Caches))

	myPeer.DefaultPacketReader = reader
	myPeer.DefaultPacketWriter = writer
	return myPeer
}
