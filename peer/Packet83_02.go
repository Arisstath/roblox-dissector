package peer

import "github.com/gskartwii/rbxfile"

// ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	Child *rbxfile.Instance
}

func DecodePacket83_02(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	result, err := decodeReplicationInstance(reader, packet, packet.stream)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return serializeReplicationInstance(layer.Child, writer, stream)
}
