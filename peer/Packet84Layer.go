package peer

import "fmt"

// Packet84Layer represents ID_MARKER - server -> client
type Packet84Layer struct {
	MarkerID uint32
}

func (thisStream *extendedReader) DecodePacket84Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet84Layer{}

	var err error
	layer.MarkerID, err = thisStream.readUint32BE()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet84Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32BE(layer.MarkerID)
}

func (layer *Packet84Layer) String() string {
	return fmt.Sprintf("ID_MARKER: %d", layer.MarkerID)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet84Layer) TypeString() string {
	return "ID_MARKER"
}

// Type implements RakNetPacket.Type()
func (Packet84Layer) Type() byte {
	return 0x84
}
