package packets

import "github.com/gskartwii/roblox-dissector/util"

// ID_CONNECTED_PING - client <-> server
type RakPing struct {
	// Timestamp (seconds)
	SendPingTime uint64
}

func NewRakPing() *RakPing {
	return &RakPing{}
}

func (thisBitstream *PacketReaderBitstream) DecodeRakPing(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewRakPing()

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}
func (layer *RakPing) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPingTime)
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

func (thisBitstream *PacketReaderBitstream) DecodeRakPong(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewRakPong()

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}

func (layer *RakPong) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
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
