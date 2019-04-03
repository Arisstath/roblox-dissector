package peer

import (
	"fmt"
)

// ID_DISCONNECTION_NOTIFICATION - client <-> server
type Packet15Layer struct {
	Reason int32
}

func NewPacket15Layer() *Packet15Layer {
	return &Packet15Layer{}
}

func (thisStream *extendedReader) DecodePacket15Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket15Layer()

	var err error
	reason, err := thisStream.readUint32BE()
	layer.Reason = int32(reason)
	return layer, err
}

func (layer *Packet15Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x15)
	if err != nil {
		return err
	}
	return stream.writeUint32BE(uint32(layer.Reason))
}

func (Packet15Layer) TypeString() string {
	return "ID_DISCONNECTION_NOTIFICATION"
}
func (layer *Packet15Layer) String() string {
	return fmt.Sprintf("ID_DISCONNECTION_NOTIFICATION: Reason %d", layer.Reason)
}
