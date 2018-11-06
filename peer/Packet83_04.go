package peer

// ID_MARKER
type Packet83_04 struct {
	MarkerId uint32
}

func DecodePacket83_04(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_04{}
	thisBitstream := packet.stream
	inner.MarkerId, err = thisBitstream.readUint32LE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_04) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32LE(layer.MarkerId)
}
