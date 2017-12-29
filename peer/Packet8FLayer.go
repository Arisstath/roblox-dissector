package peer

type Packet8FLayer struct {
	SpawnName string
}

func NewPacket8FLayer() *Packet8FLayer {
	return &Packet8FLayer{}
}

func DecodePacket8FLayer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket8FLayer()
	thisBitstream := packet.Stream

	var err error
	spawnName, err := thisBitstream.ReadHuffman()
	layer.SpawnName = string(spawnName)
	return layer, err
}

func (layer *Packet8FLayer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
	err := stream.WriteByte(0x8F)
	if err != nil {
		return err
	}
	return stream.WriteHuffman([]byte(layer.SpawnName))
}
