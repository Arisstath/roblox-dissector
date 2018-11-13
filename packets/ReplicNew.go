package packets

import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/util"

// ID_CREATE_INSTANCE
type NewInstance struct {
	// The instance that was created
	Child *rbxfile.Instance
}

func (thisBitstream *PacketReaderBitstream) DecodeNewInstance(reader util.PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	result, err := decodeReplicationInstance(reader, thisBitstream, layers)
	return &NewInstance{result}, err
}

func (layer *NewInstance) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	return serializeReplicationInstance(layer.Child, writer, stream)
}

func (NewInstance) Type() uint8 {
	return 2
}
func (NewInstance) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}
