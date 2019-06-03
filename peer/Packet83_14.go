package peer

import (
	"fmt"
)

// Packet83_14 represents ID_REPLIC_STREAM_DATA_INFO
type Packet83_14 struct {
	Int1 int32
	Int2 int32
	Int3 int32
	Int4 int32
}

func (thisStream *extendedReader) DecodePacket83_14(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_14{}

	is32Bit, err := thisStream.readBoolByte()
	if err != nil {
		return inner, err
	}
	if is32Bit {
		int1, err := thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		int2, err := thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		int3, err := thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}

		inner.Int1 = int32(int1)
		inner.Int2 = int32(int2)
		inner.Int3 = int32(int3)
	} else {
		int1, err := thisStream.ReadByte()
		if err != nil {
			return inner, err
		}
		int2, err := thisStream.ReadByte()
		if err != nil {
			return inner, err
		}
		int3, err := thisStream.ReadByte()
		if err != nil {
			return inner, err
		}

		inner.Int1 = int32(int8(int1))
		inner.Int2 = int32(int8(int2))
		inner.Int3 = int32(int8(int3))
	}

	int4, err := thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	return inner, nil
}

func isIn8BitRange(val int32) bool {
	return -127 < val < 127
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_14) Serialize(writer PacketWriter, stream *extendedWriter) error {
	is32Bit := !(isIn8BitRange(layer.Int1) && isIn8BitRange(layer.Int2) && isIn8BitRange(layer.Int3))
	err := stream.writeBoolByte(is32Bit)
	if err != nil {
		return err
	}
	if is32Bit {
		err = stream.writeUint32BE(uint32(layer.Int1))
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(layer.Int2))
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(layer.Int3))
		if err != nil {
			return err
		}
	} else {
		err = stream.WriteByte(byte(layer.Int1))
		if err != nil {
			return err
		}
		err = stream.WriteByte(byte(layer.Int2))
		if err != nil {
			return err
		}
		err = stream.WriteByte(byte(layer.Int3))
		if err != nil {
			return err
		}
	}

	return stream.writeUint32BE(uint32(layer.Int4))
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
	return fmt.Sprintf("ID_REPLIC_STREAM_DATA_INFO: %d, %d, %d, %d", layer.Int1, layer.Int2, layer.Int3, layer.Int4)
}
