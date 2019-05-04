package peer

import "fmt"

// ID_MARKER - server -> client
type Packet84Layer struct {
	MarkerId uint32
}

func NewPacket84Layer() *Packet84Layer {
	return &Packet84Layer{}
}

func (thisStream *extendedReader) DecodePacket84Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet84Layer{}

	var err error
	layer.MarkerId, err = thisStream.readUint32BE()
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

func (Packet84Layer) TypeString() string {
	return "ID_MARKER"
}
func (Packet84Layer) Type() byte {
	return 0x84
}
