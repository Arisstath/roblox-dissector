package packets

import "github.com/gskartwii/roblox-dissector/util"
// ID_PREFERRED_SPAWN_NAME - client -> server
type SpawnNamePacket struct {
	SpawnName string
}

func NewSpawnNamePacket() *SpawnNamePacket {
	return &SpawnNamePacket{}
}

func (thisBitstream *PacketReaderBitstream) DecodeSpawnNamePacket(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewSpawnNamePacket()

	var err error
	spawnName, err := thisBitstream.ReadVarLengthString()
	layer.SpawnName = string(spawnName)
	return layer, err
}

func (layer *SpawnNamePacket) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteByte(0x8F)
	if err != nil {
		return err
	}
	return stream.WriteVarLengthString(layer.SpawnName)
}
