package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"

type Packet92Layer struct {
	UnknownValue uint32
}

func NewPacket92Layer() Packet92Layer {
	return Packet92Layer{}
}

func DecodePacket92Layer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket92Layer()
	thisBitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

	var err error
	layer.UnknownValue, err = thisBitstream.ReadUint32BE()
	return layer, err
}
