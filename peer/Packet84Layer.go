package peer

import "fmt"

// ID_MARKER - server -> client
type Packet84Layer struct {
	MarkerId uint32
}

func NewPacket84Layer() *Packet84Layer {
	return &Packet84Layer{}
}

func (thisBitstream *extendedReader) DecodePacket84Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket84Layer()

	var err error
	layer.MarkerId, err = thisBitstream.readUint32BE()
	return layer, err
}

func (layer *Packet84Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x84)
	if err != nil {
		return err
	}
	return stream.writeUint32BE(layer.MarkerId)
}

func (layer *Packet84Layer) String() string {
	return fmt.Sprintf("ID_MARKER: %d", layer.MarkerId)
}
