package peer

import "net"

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

func (peer *ConnectedPeer) sendACKs() error {
	if len(peer.mustACK) == 0 {
		return nil
	}
	err := peer.WriteACKs(peer.mustACK, false)
	if err != nil {
		return err
	}

	peer.mustACK = []int{}
	return nil
}

// NewConnectedPeer returns a new ConnectedPeer instance
// withClient specifies whether the target of the connection
// is a client, i.e. if the caller is acting as a server
func NewConnectedPeer(context *CommunicationContext, withClient bool) *ConnectedPeer {
	myPeer := &ConnectedPeer{}

	reader := NewPacketReader()
	writer := NewPacketWriter()

	reader.SetContext(context)
	writer.SetContext(context)
	reader.SetIsClient(withClient)
	writer.SetToClient(withClient)

	myPeer.DefaultPacketReader = reader
	myPeer.DefaultPacketWriter = writer
	return myPeer
}
