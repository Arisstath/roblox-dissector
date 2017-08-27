package peer

type Packet83_10 struct {
	TagId uint32
}

func DecodePacket83_10(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_10{}
	thisBitstream := packet.Stream
	inner.TagId, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}
