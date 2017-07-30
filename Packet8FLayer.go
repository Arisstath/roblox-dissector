package main
import "github.com/google/gopacket"

type Packet8FLayer struct {
	SpawnName string
}

func NewPacket8FLayer() Packet8FLayer {
	return Packet8FLayer{}
}

func DecodePacket8FLayer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket8FLayer()

	var err error
	spawnName, err := thisBitstream.ReadHuffman()
	layer.SpawnName = string(spawnName)
	return layer, err
}
