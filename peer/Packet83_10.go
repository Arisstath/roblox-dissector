package peer

// Packet83_10 represents ID_TAG
type Packet83_10 struct {
	// 12 => ReplicatedFirst replication finished
	// 13 => Initial replication finished
	TagID uint32
}

func (thisStream *extendedReader) DecodePacket83_10(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_10{}

	inner.TagID, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_10) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.TagID)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_10) Type() uint8 {
	return 0x10
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_10) TypeString() string {
	return "ID_REPLIC_TAG"
}

func (layer *Packet83_10) String() string {
	switch layer.TagID {
	case 12:
		return "ID_REPLIC_TAG: ReplicatedFirst finished"
	case 13:
		return "ID_REPLIC_TAG: JoinData replication finished"
	default:
		return "ID_REPLIC_TAG: Unknown tag"
	}
}
