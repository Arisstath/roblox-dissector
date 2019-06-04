package peer

import (
	"fmt"
)

// Packet83_14 represents ID_REPLIC_STREAM_DATA_INFO
type Packet83_14 struct {
	Region StreamInfo
	Int1   int32
}

func (thisStream *extendedReader) DecodePacket83_14(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_14{}
	var err error

	inner.Region, err = thisStream.readStreamInfo()
	int1, err := thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	inner.Int1 = int32(int1)
	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_14) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := layer.Region.Serialize(stream)
	if err != nil {
		return err
	}

	return stream.writeUint32BE(uint32(layer.Int1))
}

// Type implements Packet83Subpacket.Type()
func (Packet83_14) Type() uint8 {
	return 0x14
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_14) TypeString() string {
	return "ID_REPLIC_STREAM_DATA_INFO"
}

func (layer *Packet83_14) String() string {
	return fmt.Sprintf("ID_REPLIC_STREAM_DATA_INFO: %s, %d", layer.Region, layer.Int1)
}
