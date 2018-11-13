package packets

import "github.com/gskartwii/roblox-dissector/util"

// ID_MARKER
type ReplicationMarker struct {
	MarkerId uint32
}

func (thisBitstream *PacketReaderBitstream) DecodeReplicationMarker(reader util.PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	inner := &ReplicationMarker{}

	inner.MarkerId, err = thisBitstream.ReadUint32LE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *ReplicationMarker) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	return stream.WriteUint32LE(layer.MarkerId)
}

func (ReplicationMarker) Type() uint8 {
	return 4
}
func (ReplicationMarker) TypeString() string {
	return "ID_REPLIC_MARKER"
}
