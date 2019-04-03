package peer

import "sort"
import "net"

type ErrorHandler func(error)

// ConnectedPeer describes a connection to a peer
// FIXME: ACKs and NAKs are not properly reacted to.
// create a resend queue before you forget!
type ConnectedPeer struct {
	// Reader is a PacketReader reading packets sent by the peer.
	*DefaultPacketReader
	// Writer is a PacketWriter writing packets to the peer.
	*DefaultPacketWriter
	DestinationAddress *net.UDPAddr

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
		Flags: RakNetFlags{
			IsValid: true,
			IsACK:   true,
		},
		ACKs: ackStructure,
	}

	peer.WriteRakNet(result)
}

func NewConnectedPeer(context *CommunicationContext, withClient bool) *ConnectedPeer {
	myPeer := &ConnectedPeer{}

	reader := NewPacketReader()
	writer := NewPacketWriter()

	reader.SetContext(context)
	writer.SetContext(context)
	reader.SetCaches(new(Caches))
	writer.SetCaches(new(Caches))
	reader.SetIsClient(withClient)
	writer.SetToClient(withClient)

	myPeer.DefaultPacketReader = reader
	myPeer.DefaultPacketWriter = writer
	return myPeer
}
