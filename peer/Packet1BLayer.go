package peer
import "io/ioutil"

// ID_TIMESTAMP - client <-> server
type Packet1BLayer struct {
	// Timestamp of when this packet was sent
	Timestamp uint64
	stream *extendedReader
}

func NewPacket1BLayer() *Packet1BLayer {
	return &Packet1BLayer{}
}

func decodePacket1BLayer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket1BLayer()
	thisBitstream := packet.stream

	var err error
	layer.Timestamp, err = thisBitstream.bits(64)
	if err != nil {
		return layer, err
	}
	layer.stream = thisBitstream

	return layer, err
}

func (layer *Packet1BLayer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x1B)
	if err != nil {
		return err
	}
	err = stream.bits(64, layer.Timestamp)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(layer.stream.GetReader())
	if err != nil {
		return err
	}
	err = stream.allBytes(content)
	return err
}
