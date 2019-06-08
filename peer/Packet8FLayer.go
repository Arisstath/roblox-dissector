package peer

import "fmt"

// Packet8FLayer represents ID_PREFERRED_SPAWN_NAME - client -> server
type Packet8FLayer struct {
	SpawnName string
}

func (thisStream *extendedReader) DecodePacket8FLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet8FLayer{}

	var err error
	spawnName, err := thisStream.readVarLengthString()
	layer.SpawnName = string(spawnName)
	return layer, err
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet8FLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeVarLengthString(layer.SpawnName)
}

func (layer *Packet8FLayer) String() string {
	return fmt.Sprintf("ID_PREFERRED_SPAWN_NAME: %s", layer.SpawnName)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet8FLayer) TypeString() string {
	return "ID_PREFERRED_SPAWN_NAME"
}

// Type implements RakNetPacket.Type()
func (Packet8FLayer) Type() byte {
	return 0x8F
}
