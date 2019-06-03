package peer

import "fmt"

// Packet83_0C represents ID_UPDATE_CLIENT_QUOTA
type Packet83_0C struct {
	QuotaDiff       int32
	MaxRegionRadius int16
}

func (thisStream *extendedReader) DecodePacket83_0C(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_0C{}

	quotaDiff, err := thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	inner.QuotaDiff = int32(quotaDiff)
	maxRegionRadius, err := thisStream.readUint16BE()
	if err != nil {
		return inner, err
	}
	inner.MaxRegionRadius = int16(maxRegionRadius)

	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_0C) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeUint32BE(uint32(layer.QuotaDiff))
	if err != nil {
		return err
	}
	return stream.writeUint16BE(uint16(layer.MaxRegionRadius))
}

// Type implements Packet83Subpacket.Type()
func (Packet83_0C) Type() uint8 {
	return 0xC
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_0C) TypeString() string {
	return "ID_REPLIC_UPDATE_CLIENT_QUOTA"
}

func (layer *Packet83_0C) String() string {
	return fmt.Sprintf("ID_REPLIC_UPDATE_CLIENT_QUOTA: diff %d, max radius %d", layer.QuotaDiff, layer.MaxRegionRadius)
}
