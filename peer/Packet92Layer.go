package peer

// ID_PLACEID_VERIFICATION - client -> server
type Packet92Layer struct {
	PlaceId uint32
}

func NewPacket92Layer() *Packet92Layer {
	return &Packet92Layer{}
}

func decodePacket92Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket92Layer()
	thisBitstream := packet.stream

	var err error
	layer.PlaceId, err = thisBitstream.readUint32BE()
	return layer, err
}

func (layer *Packet92Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	err := stream.WriteByte(0x92)
	if err != nil {
		return err
	}
	return stream.writeUint32BE(layer.PlaceId)
}
