package peer

// ID_PREFERRED_SPAWN_NAME - client -> server
type SpawnNamePacket struct {
	SpawnName string
}

func NewSpawnNamePacket() *SpawnNamePacket {
	return &SpawnNamePacket{}
}

func (thisBitstream *extendedReader) DecodeSpawnNamePacket(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewSpawnNamePacket()

	var err error
	spawnName, err := thisBitstream.readVarLengthString()
	layer.SpawnName = string(spawnName)
	return layer, err
}

func (layer *SpawnNamePacket) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x8F)
	if err != nil {
		return err
	}
	return stream.writeVarLengthString(layer.SpawnName)
}
