package peer

type Packet90Layer struct {
	SchemaVersion uint32
	RequestedFlags []string
}

func NewPacket90Layer() *Packet90Layer {
	return &Packet90Layer{}
}

func DecodePacket90Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket90Layer()
	thisBitstream := packet.Stream

	var err error
	layer.SchemaVersion, err = thisBitstream.ReadUint32BE()

	flagsLen, err := thisBitstream.ReadUint16BE()
	if err != nil {
		return layer, err
	}

	layer.RequestedFlags = make([]string, flagsLen)
	for i := 0; i < int(flagsLen); i++ {
		flagLen, err := thisBitstream.ReadUint8()
		if err != nil {
			return layer, err
		}
		layer.RequestedFlags[i], err = thisBitstream.ReadASCII(int(flagLen))
		if err != nil {
			return layer, err
		}
	}
	return layer, nil
}

func (layer *Packet90Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
	err := stream.WriteByte(0x90)
	if err != nil {
		return err
	}
	err = stream.WriteUint32BE(layer.SchemaVersion)
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(uint16(len(layer.RequestedFlags)))
	if err != nil {
		return err
	}
	for i := 0; i < len(layer.RequestedFlags); i++ {
		err = stream.WriteByte(uint8(len(layer.RequestedFlags[i])))
		if err != nil {
			return err
		}
		err = stream.WriteASCII(layer.RequestedFlags[i])
		if err != nil {
			return err
		}
	}
	return nil
}
