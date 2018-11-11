package peer

// ID_PLACEID_VERIFICATION - client -> server
type VerifyPlaceId struct {
	PlaceId int64
}

func NewVerifyPlaceId() *VerifyPlaceId {
	return &VerifyPlaceId{}
}

func (thisBitstream *PacketReaderBitstream) DecodeVerifyPlaceId(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewVerifyPlaceId()

	var err error
	layer.PlaceId, err = thisBitstream.readVarsint64()
	return layer, err
}

func (layer *VerifyPlaceId) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteByte(0x92)
	if err != nil {
		return err
	}
	return stream.writeVarsint64(layer.PlaceId)
}
