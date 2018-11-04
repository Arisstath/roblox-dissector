package peer

// ID_PREFERRED_SPAWN_NAME - client -> server
type Packet8FLayer struct {
	SpawnName string
}

func NewPacket8FLayer() *Packet8FLayer {
	return &Packet8FLayer{}
}

func DecodePacket8FLayer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	layer := NewPacket8FLayer()
	thisBitstream := packet.stream

	var err error
	spawnName, err := thisBitstream.readVarLengthString()
	layer.SpawnName = string(spawnName)
	return layer, err
}

func (layer *Packet8FLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x8F)
	if err != nil {
		return err
	}
	return stream.writeVarLengthString(layer.SpawnName)
}
