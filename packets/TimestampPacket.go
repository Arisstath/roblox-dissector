package peer

// ID_TIMESTAMP - client <-> server
type Timestamp struct {
	// Timestamp of when this packet was sent
	Timestamp  uint64
	Timestamp2 uint64
	stream     *extendedReader
}

func NewTimestamp() *Timestamp {
	return &Timestamp{}
}

func (thisBitstream *extendedReader) DecodeTimestamp(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewTimestamp()

	var err error
	layer.Timestamp, err = thisBitstream.bits(64)
	if err != nil {
		return layer, err
	}
	layer.Timestamp2, err = thisBitstream.bits(64)
	if err != nil {
		return layer, err
	}
	layer.stream = thisBitstream

	return layer, err
}

func (layer *Timestamp) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x1B)
	if err != nil {
		return err
	}
	err = stream.bits(64, layer.Timestamp)
	if err != nil {
		return err
	}
	err = stream.bits(64, layer.Timestamp2)
	return err
}
