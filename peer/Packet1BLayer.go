package peer

// Packet1BLayer represents ID_TIMESTAMP - client <-> server
type Packet1BLayer struct {
	// Timestamp of when this packet was sent
	Timestamp  uint64
	Timestamp2 uint64
}

func (thisStream *extendedReader) DecodePacket1BLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet1BLayer{}

	var err error
	layer.Timestamp, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.Timestamp2, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}

	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
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

// TypeString implements RakNetPacket.TypeString()
func (Packet1BLayer) TypeString() string {
	return "ID_TIMESTAMP"
}

// Type implements RakNetPacket.Type()
func (Packet1BLayer) Type() byte {
	return 0x1B
}
