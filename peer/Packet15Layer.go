package peer

import (
	"fmt"
)

// Packet15Layer represents ID_DISCONNECTION_NOTIFICATION - client <-> server
type Packet15Layer struct {
	Reason int32
}

func (thisStream *extendedReader) DecodePacket15Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet15Layer{}

	var err error
	reason, err := thisStream.readUint32BE()
	layer.Reason = int32(reason)
	// DisconnectReceivePacketError: most parser functions -- NOT NetworkStream reading
	// DisconnectReceivePacketStreamError: when reading from network streams
	// DisconnectSendPacketError: from ClusterJobStep, from DataOutStep, when creating player
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet15Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(uint32(layer.Reason))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet15Layer) TypeString() string {
	return "ID_DISCONNECTION_NOTIFICATION"
}
func (layer *Packet15Layer) String() string {
	return fmt.Sprintf("ID_DISCONNECTION_NOTIFICATION: Reason %d", layer.Reason)
}

// Type implements RakNetPacket.Type()
func (Packet15Layer) Type() byte {
	return 0x15
}
