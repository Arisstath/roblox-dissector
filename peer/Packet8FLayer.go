package peer

import "fmt"

// ID_PREFERRED_SPAWN_NAME - client -> server
type Packet8FLayer struct {
	SpawnName string
}

func NewPacket8FLayer() *Packet8FLayer {
	return &Packet8FLayer{}
}

func (thisStream *extendedReader) DecodePacket8FLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket8FLayer()

	var err error
	spawnName, err := thisStream.readVarLengthString()
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

func (layer *Packet8FLayer) String() string {
	return fmt.Sprintf("ID_PREFERRED_SPAWN_NAME: %s", layer.SpawnName)
}

func (Packet8FLayer) TypeString() string {
	return "ID_PREFERRED_SPAWN_NAME"
}

func (Packet8FLayer) Type() byte {
	return 0x8F
}
