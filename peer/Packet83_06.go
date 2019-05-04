package peer

// Packet83_06 represents ID_PING_BACK
type Packet83_06 struct {
	// Always true
	IsPingBack bool
	Timestamp  uint64
	SendStats  uint32
	// Hack flags
	ExtraStats uint32
}

func (thisStream *extendedReader) DecodePacket83_06(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_06{}

	inner.IsPingBack, err = thisStream.readBoolByte()
	if err != nil {
		return inner, err
	}

	inner.Timestamp, err = thisStream.readUint64BE()
	if err != nil {
		return inner, err
	}
	inner.SendStats, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	inner.ExtraStats, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	if inner.Timestamp&0x20 != 0 {
		inner.ExtraStats ^= 0xFFFFFFFF
	}

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_06) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.writeBoolByte(layer.IsPingBack)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.Timestamp)
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

// Type implements Packet83Subpacket.Type()
func (Packet83_06) Type() uint8 {
	return 6
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_06) TypeString() string {
	return "ID_REPLIC_PING_BACK"
}

func (layer *Packet83_06) String() string {
	// yes, these packets are boring
	return "ID_REPLIC_PING_BACK"
}
