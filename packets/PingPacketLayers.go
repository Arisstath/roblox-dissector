package packets

// ID_CONNECTED_PING - client <-> server
type RakPing struct {
	// Timestamp (seconds)
	SendPingTime uint64
}

func NewRakPing() *RakPing {
	return &RakPing{}
}

func (thisBitstream *PacketReaderBitstream) DecodeRakPing(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewRakPing()

	var err error
	layer.SendPingTime, err = thisBitstream.readUint64BE()

	return layer, err
}
func (layer *RakPing) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	return err
}

// ID_CONNECTED_PONG - client <-> server
type RakPong struct {
	// Timestamp from ID_CONNECTED_PING
	SendPingTime uint64
	// Timestamp of reply (seconds)
	SendPongTime uint64
}

func NewRakPong() *RakPong {
	return &RakPong{}
}

func (thisBitstream *PacketReaderBitstream) DecodeRakPong(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewRakPong()

	var err error
	layer.SendPingTime, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.readUint64BE()

	return layer, err
}

func (layer *RakPong) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
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
