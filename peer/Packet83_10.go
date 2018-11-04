package peer

// ID_TAG
type Packet83_10 struct {
	// 12 or 13
	TagId uint32
}

func DecodePacket83_10(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_10{}
	thisBitstream := packet.stream
	inner.TagId, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_10) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.TagId)
}
