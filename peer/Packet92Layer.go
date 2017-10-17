package peer

type Packet92Layer struct {
	UnknownValue uint32
}

func NewPacket92Layer() *Packet92Layer {
	return &Packet92Layer{}
}

func DecodePacket92Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket92Layer()
	thisBitstream := packet.Stream

	var err error
	layer.UnknownValue, err = thisBitstream.ReadUint32BE()
	return layer, err
}

func (layer *Packet92Layer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
	return stream.WriteUint32BE(layer.UnknownValue)
}
