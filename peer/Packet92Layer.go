package peer

import "fmt"

// Packet92Layer represents ID_PLACEID_VERIFICATION - client -> server
type Packet92Layer struct {
	PlaceID int64
}

func (thisStream *extendedReader) DecodePacket92Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet92Layer{}

	var err error
	layer.PlaceID, err = thisStream.readVarsint64()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet92Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x92)
	if err != nil {
		return err
	}
	return stream.writeVarsint64(layer.PlaceID)
}

func (layer *Packet92Layer) String() string {
	return fmt.Sprintf("ID_PLACEID_VERIFICATION: %d", layer.PlaceID)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet92Layer) TypeString() string {
	return "ID_PLACEID_VERIFICATION"
}

// Type implements RakNetPacket.Type()
func (Packet92Layer) Type() byte {
	return 0x92
}
