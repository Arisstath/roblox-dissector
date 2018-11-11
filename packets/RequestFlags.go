package packets

// ID_PROTOCOL_SYNC - client -> server
type FlagRequest struct {
	SchemaVersion  uint32
	RequestedFlags []string
}

func NewFlagRequest() *FlagRequest {
	return &FlagRequest{}
}

func (thisBitstream *PacketReaderBitstream) DecodeFlagRequest(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewFlagRequest()

	var err error
	layer.SchemaVersion, err = thisBitstream.readUint32BE()

	flagsLen, err := thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}

	layer.RequestedFlags = make([]string, flagsLen)
	for i := 0; i < int(flagsLen); i++ {
		flagLen, err := thisBitstream.readUint8()
		if err != nil {
			return layer, err
		}
		layer.RequestedFlags[i], err = thisBitstream.readASCII(int(flagLen))
		if err != nil {
			return layer, err
		}
	}
	return layer, nil
}

func (layer *FlagRequest) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteByte(0x90)
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(layer.SchemaVersion)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(uint16(len(layer.RequestedFlags)))
	if err != nil {
		return err
	}
	for i := 0; i < len(layer.RequestedFlags); i++ {
		err = stream.WriteByte(uint8(len(layer.RequestedFlags[i])))
		if err != nil {
			return err
		}
		err = stream.writeASCII(layer.RequestedFlags[i])
		if err != nil {
			return err
		}
	}
	return nil
}
