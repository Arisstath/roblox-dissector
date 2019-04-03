package peer

import "fmt"

// ID_MARKER
type Packet83_04 struct {
	MarkerId uint32
}

func (thisStream *extendedReader) DecodePacket83_04(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_04{}

	inner.MarkerId, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

func (layer *Packet83_04) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.MarkerId)
}

func (Packet83_04) Type() uint8 {
	return 4
}
func (Packet83_04) TypeString() string {
	return "ID_REPLIC_MARKER"
}

func (layer *Packet83_04) String() string {
	return fmt.Sprintf("ID_REPLIC_MARKER: %d", layer.MarkerId)
}
