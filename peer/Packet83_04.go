package peer

import "fmt"

// Packet83_04 represents ID_MARKER
type Packet83_04 struct {
	MarkerID uint32
}

func (thisStream *extendedReader) DecodePacket83_04(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_04{}

	inner.MarkerID, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_04) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.MarkerID)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_04) Type() uint8 {
	return 4
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_04) TypeString() string {
	return "ID_REPLIC_MARKER"
}

func (layer *Packet83_04) String() string {
	return fmt.Sprintf("ID_REPLIC_MARKER: %d", layer.MarkerID)
}
