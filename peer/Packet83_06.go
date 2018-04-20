package peer

// ID_PING_BACK
type Packet83_06 struct {
	// Always true
	IsPingBack bool
	Timestamp uint64
	SendStats uint32
	// Hack flags
	ExtraStats uint32
}

func decodePacket83_06(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_06{}
	thisBitstream := packet.stream
	inner.IsPingBack, err = thisBitstream.readBoolByte()
	if err != nil {
		return inner, err
	}

	inner.Timestamp, err = thisBitstream.bits(64)
	if err != nil {
		return inner, err
	}
	inner.SendStats, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}
	inner.ExtraStats, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}
	if inner.Timestamp & 0x20 != 0 {
		inner.ExtraStats ^= 0xFF
	}

	return inner, err
}

func (layer *Packet83_06) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    var err error
    err = stream.writeBool(layer.IsPingBack)
    if err != nil {
        return err
    }
    err = stream.bits(64, layer.Timestamp)
    if err != nil {
        return err
    }
    err = stream.writeUint32BE(layer.SendStats)
    if err != nil {
        return err
    }
	if layer.Timestamp & 0x20 != 0 {
		layer.ExtraStats ^= 0xFF
	}

    err = stream.writeUint32BE(layer.ExtraStats)
    return err
}
