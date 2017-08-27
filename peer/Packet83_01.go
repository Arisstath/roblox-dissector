package peer

type Packet83_01 struct {
	Object1 Object
}

func DecodePacket83_01(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_01{}
	thisBitstream := packet.Stream
	inner.Object1, err = thisBitstream.ReadObject(false, context)
	return inner, err
}
