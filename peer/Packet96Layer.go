package peer

import "fmt"

type Packet96Layer struct {
	Request bool
	Version uint32
}

func NewPacket96Layer() *Packet96Layer {
	return &Packet96Layer{}
}

func (thisStream *extendedReader) DecodePacket96Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket96Layer()
	var err error

	layer.Request, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	if !layer.Request {
		return layer, err
	}

	layer.Version, err = thisStream.readUint32BE()
	return layer, err
}

func (layer *Packet96Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeBoolByte(layer.Request)
	if err != nil {
		return err
	}
	if !layer.Request {
		return nil
	}

	return stream.writeUint32BE(layer.Version)
}

func (layer *Packet96Layer) String() string {
	return fmt.Sprintf("ID_REQUEST_STATS: Version %d", layer.Version)
}

func (Packet96Layer) TypeString() string {
	return "ID_REQUEST_STATS"
}
