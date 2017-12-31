package peer

// ID_CONNECTED_PING - client <-> server
type Packet00Layer struct {
	// Timestamp (seconds)
	SendPingTime uint64
}

func NewPacket00Layer() *Packet00Layer {
	return &Packet00Layer{}
}

func decodePacket00Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket00Layer()
	thisBitstream := packet.stream

	var err error
	layer.SendPingTime, err = thisBitstream.readUint64BE()

	return layer, err
}
func (layer *Packet00Layer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	return err
}

// ID_CONNECTED_PONG - client <-> server
type Packet03Layer struct {
	// Timestamp from ID_CONNECTED_PING
	SendPingTime uint64
	// Timestamp of reply (seconds)
	SendPongTime uint64
}

func NewPacket03Layer() *Packet03Layer {
	return &Packet03Layer{}
}

func decodePacket03Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket03Layer()
	thisBitstream := packet.stream

	var err error
	layer.SendPingTime, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.readUint64BE()

	return layer, err
}

func (layer *Packet03Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(3)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPongTime)
	return err
}
