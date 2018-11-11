package packets

// ID_TAG
type ReplicationTag struct {
	// 12 or 13
	TagId uint32
}

func (thisBitstream *PacketReaderBitstream) DecodeReplicationTag(reader PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	inner := &ReplicationTag{}

	inner.TagId, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *ReplicationTag) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	return stream.writeUint32BE(layer.TagId)
}

func (ReplicationTag) Type() uint8 {
	return 0x10
}
func (ReplicationTag) TypeString() string {
	return "ID_REPLIC_TAG"
}
