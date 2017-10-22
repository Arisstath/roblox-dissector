package peer

type Packet90Layer struct {
	SchemaVersion uint32
}

func NewPacket90Layer() *Packet90Layer {
	return &Packet90Layer{}
}

func DecodePacket90Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket90Layer()
	thisBitstream := packet.Stream

	var err error
	layer.SchemaVersion, err = thisBitstream.ReadUint32BE()
	return layer, err
}

func (layer *Packet90Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
	return stream.WriteUint32BE(layer.SchemaVersion)
}
