package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"

type Packet83Layer struct {
	SpawnName string
}

func NewPacket83Layer() Packet83Layer {
	return Packet83Layer{}
}

func DecodePacket83Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket83Layer()

	var err error
	spawnName, err := thisBitstream.ReadHuffman()
	layer.SpawnName = string(spawnName)
	return layer, err
}
