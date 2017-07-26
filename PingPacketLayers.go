package main
import "bytes"
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"

type Packet00Layer struct {
	SendPingTime uint64
}

func NewPacket00Layer() Packet00Layer {
	return Packet00Layer{}
}

func DecodePacket00Layer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket00Layer()
	thisBitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}

type Packet03Layer struct {
	SendPingTime uint64
	SendPongTime uint64
}

func NewPacket03Layer() Packet03Layer {
	return Packet03Layer{}
}

func DecodePacket03Layer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket03Layer()
	thisBitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

	var err error
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.ReadUint64BE()

	return layer, err
}
