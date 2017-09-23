package peer
import "io/ioutil"

type Packet1BLayer struct {
	Timestamp uint64
	Stream *ExtendedReader
}

func NewPacket1BLayer() *Packet1BLayer {
	return &Packet1BLayer{}
}

func DecodePacket1BLayer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket1BLayer()
	thisBitstream := packet.Stream

	var err error
	layer.Timestamp, err = thisBitstream.Bits(64)
	if err != nil {
		return layer, err
	}
	layer.Stream = thisBitstream

	return layer, err
}

func (layer *Packet1BLayer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
	var err error
	err = stream.WriteByte(0x1B)
	if err != nil {
		return err
	}
	err = stream.Bits(64, layer.Timestamp)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(layer.Stream.GetReader())
	if err != nil {
		return err
	}
	err = stream.AllBytes(content)
	return err
}
