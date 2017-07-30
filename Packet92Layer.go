package main
import "github.com/google/gopacket"

type Packet92Layer struct {
	UnknownValue uint32
}

func NewPacket92Layer() Packet92Layer {
	return Packet92Layer{}
}

func DecodePacket92Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket92Layer()

	var err error
	layer.UnknownValue, err = thisBitstream.ReadUint32BE()
	return layer, err
}
