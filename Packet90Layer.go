package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"

type Packet90Layer struct {
	SchemaVersion uint32
}

func NewPacket90Layer() Packet90Layer {
	return Packet90Layer{}
}

func DecodePacket90Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket90Layer()

	var err error
	layer.SchemaVersion, err = thisBitstream.ReadUint32BE()
	return layer, err
}
