package peer

import "github.com/gskartwii/rbxfile"

// ID_CREATE_INSTANCE
type NewInstance struct {
	// The instance that was created
	Child *rbxfile.Instance
}

func (thisBitstream *extendedReader) DecodeNewInstance(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	result, err := decodeReplicationInstance(reader, thisBitstream, layers)
	return &NewInstance{result}, err
}

func (layer *NewInstance) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return serializeReplicationInstance(layer.Child, writer, stream)
}

func (NewInstance) Type() uint8 {
	return 2
}
func (NewInstance) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}
