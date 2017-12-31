package peer

// ID_TAG
type Packet83_10 struct {
	// 12 or 13
	TagId uint32
}

func decodePacket83_10(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_10{}
	thisBitstream := packet.stream
	inner.TagId, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_10) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    return stream.writeUint32BE(layer.TagId)
}
