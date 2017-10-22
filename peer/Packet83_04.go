package peer

type Packet83_04 struct {
	MarkerId uint32
}

func DecodePacket83_04(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_04{}
	thisBitstream := packet.Stream
	inner.MarkerId, err = thisBitstream.ReadUint32LE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_04) Serialize(isClient bool, context *CommunicationContext, stream *ExtendedWriter) error {
    return stream.WriteUint32LE(layer.MarkerId)
}
