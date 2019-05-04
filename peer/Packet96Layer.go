package peer

import "fmt"

// Packet96Layer represents ID_REQUEST_STATS
type Packet96Layer struct {
	Request bool
	Version uint32
}

func (thisStream *extendedReader) DecodePacket96Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet96Layer{}
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

// Serialize implements RakNetPacket.Serialize
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

// TypeString implements RakNetPacket.TypeString()
func (Packet96Layer) TypeString() string {
	return "ID_REQUEST_STATS"
}

// Type implements RakNetPacket.Type()
func (Packet96Layer) Type() byte {
	return 0x96
}
