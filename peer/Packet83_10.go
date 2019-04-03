package peer

// ID_TAG
type Packet83_10 struct {
	// 12 or 13
	TagId uint32
}

func (thisStream *extendedReader) DecodePacket83_10(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_10{}

	inner.TagId, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_10) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.TagId)
}

func (Packet83_10) Type() uint8 {
	return 0x10
}
func (Packet83_10) TypeString() string {
	return "ID_REPLIC_TAG"
}

func (layer *Packet83_10) String() string {
	switch layer.TagId {
	case 12:
		return "ID_REPLIC_TAG: ReplicatedFirst finished"
	case 13:
		return "ID_REPLIC_TAG: JoinData replication finished"
	default:
		return "ID_REPLIC_TAG: Unknown tag"
	}
}
