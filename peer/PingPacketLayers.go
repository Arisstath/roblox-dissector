package peer

// ID_CONNECTED_PING - client <-> server
type Packet00Layer struct {
	// Timestamp (seconds)
	SendPingTime uint64
}

func NewPacket00Layer() *Packet00Layer {
	return &Packet00Layer{}
}

func (thisStream *extendedReader) DecodePacket00Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet00Layer{}

	var err error
	layer.SendPingTime, err = thisStream.readUint64BE()

	return layer, err
}
func (layer *Packet00Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	return err
}
func (layer *Packet00Layer) String() string {
	return "ID_CONNECTED_PING"
}
func (Packet00Layer) TypeString() string {
	return "ID_CONNECTED_PING"
}
func (Packet00Layer) Type() byte {
	return 0
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

func (thisStream *extendedReader) DecodePacket03Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet03Layer{}

	var err error
	layer.SendPingTime, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisStream.readUint64BE()

	return layer, err
}

func (layer *Packet03Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
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

func (layer *Packet03Layer) String() string {
	return "ID_CONNECTED_PONG"
}
func (Packet03Layer) TypeString() string {
	return "ID_CONNECTED_PONG"
}
func (Packet03Layer) Type() byte {
	return 3
}
