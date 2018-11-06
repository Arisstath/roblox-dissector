package peer

// ID_PLACEID_VERIFICATION - client -> server
type Packet92Layer struct {
	PlaceId int64
}

func NewPacket92Layer() *Packet92Layer {
	return &Packet92Layer{}
}

func DecodePacket92Layer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	layer := NewPacket92Layer()
	thisBitstream := packet.stream

	var err error
	layer.PlaceId, err = thisBitstream.readVarsint64()
	return layer, err
}

func (layer *Packet92Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x92)
	if err != nil {
		return err
	}
	return stream.writeVarsint64(layer.PlaceId)
}
