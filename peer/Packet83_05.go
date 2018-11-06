package peer

// ID_PING
type Packet83_05 struct {
	// Always false
	IsPingBack bool
	Timestamp  uint64
	SendStats  uint32
	// Hack flags
	ExtraStats uint32
}

func DecodePacket83_05(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_05{}
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
	if inner.Timestamp&0x20 != 0 {
		inner.ExtraStats ^= 0xFFFFFFFF
	}

	return inner, err
}

func (layer *Packet83_05) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.writeBoolByte(layer.IsPingBack)
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
	if layer.Timestamp&0x20 != 0 {
		layer.ExtraStats ^= 0xFFFFFFFF
	}

	err = stream.writeUint32BE(layer.ExtraStats)
	return err
}

func (Packet83_05) Type() uint8 {
	return 5
}
func (Packet83_05) TypeString() string {
	return "ID_REPLIC_PING"
}
