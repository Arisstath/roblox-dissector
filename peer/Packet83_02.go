package peer

import "github.com/gskartwii/rbxfile"

// ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	Child *rbxfile.Instance
}

func (thisBitstream *extendedReader) DecodePacket83_02(reader PacketReader) (Packet83Subpacket, error) {
	result, err := decodeReplicationInstance(reader, packet, packet.stream)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return serializeReplicationInstance(layer.Child, writer, stream)
}

func (Packet83_02) Type() uint8 {
	return 2
}
func (Packet83_02) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}
