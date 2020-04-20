package peer

import "fmt"

// Packet98Layer represents ID_KICK_MESSAGE - server -> client
type Packet98Layer struct {
	Message string
}

func (thisStream *extendedReader) DecodePacket98Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet98Layer{}

	message, err := thisStream.readUint32AndString()
	layer.Message = message.(string)
	return layer, err
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet98Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeUint32AndString(layer.Message)
}

func (layer *Packet98Layer) String() string {
	return fmt.Sprintf("ID_KICK_MESSAGE: %s", layer.Message)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet98Layer) TypeString() string {
	return "ID_KICK_MESSAGE"
}

// Type implements RakNetPacket.Type()
func (Packet98Layer) Type() byte {
	return 0x98
}
