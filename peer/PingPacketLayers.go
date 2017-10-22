package peer

type Packet00Layer struct {
	SendPingTime uint64
}

func NewPacket00Layer() *Packet00Layer {
	return &Packet00Layer{}
}

func DecodePacket00Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket00Layer()
	thisBitstream := packet.Stream

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}
func (layer *Packet00Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
	var err error
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPingTime)
	return err
}

type Packet03Layer struct {
	SendPingTime uint64
	SendPongTime uint64
}

func NewPacket03Layer() *Packet03Layer {
	return &Packet03Layer{}
}

func DecodePacket03Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket03Layer()
	thisBitstream := packet.Stream

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}

func (layer *Packet03Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
	var err error
	err = stream.WriteByte(3)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPongTime)
	return err
}
