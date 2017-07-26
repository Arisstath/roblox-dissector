package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"

type Packet8FLayer struct {
	SpawnName string
}

func NewPacket8FLayer() Packet8FLayer {
	return Packet8FLayer{}
}

func DecodePacket8FLayer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket8FLayer()
	thisBitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

	var err error
	spawnName, err := thisBitstream.ReadHuffman()
	layer.SpawnName = string(spawnName)
	return layer, err
}
