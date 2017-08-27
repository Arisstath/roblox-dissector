package peer

type Packet83_05 struct {
	Bool1 bool
	Int1 uint64
	Int2 uint32
	Int3 uint32
}

func DecodePacket83_05(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_05{}
	thisBitstream := packet.Stream
	inner.Bool1, err = thisBitstream.ReadBool()
	if err != nil {
		return inner, err
	}

	inner.Int1, err = thisBitstream.Bits(64)
	if err != nil {
		return inner, err
	}
	inner.Int2, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	inner.Int3, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	//println(DebugInfo(context, packet), "Receive 0x05", inner.Bool1, ",", inner.Int1, ",", inner.Int2, ",", inner.Int3)

	return inner, err
}
