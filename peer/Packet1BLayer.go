package peer

// ID_TIMESTAMP - client <-> server
type Packet1BLayer struct {
	// Timestamp of when this packet was sent
	Timestamp  uint64
	Timestamp2 uint64
	stream     *extendedReader
}

func NewPacket1BLayer() *Packet1BLayer {
	return &Packet1BLayer{}
}

func (thisBitstream *extendedReader) DecodePacket1BLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket1BLayer()

	var err error
	layer.Timestamp, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.Timestamp2, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.stream = thisBitstream

	return layer, err
}

func (layer *Packet1BLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x1B)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.Timestamp)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.Timestamp2)
	return err
}

func (layer *Packet1BLayer) String() string {
	return "ID_TIMESTAMP"
}

func (Packet1BLayer) TypeString() string {
	return "ID_TIMESTAMP"
}
