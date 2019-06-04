package peer

import (
	"fmt"
)

// StreamInfo represents a streaming region id
type StreamInfo struct {
	X int32
	Y int32
	Z int32
}

func (thisStream *extendedReader) readStreamInfo() (StreamInfo, error) {
	info := StreamInfo{}

	is32Bit, err := thisStream.readBoolByte()
	if err != nil {
		return info, err
	}
	if is32Bit {
		x, err := thisStream.readUint32BE()
		if err != nil {
			return info, err
		}
		y, err := thisStream.readUint32BE()
		if err != nil {
			return info, err
		}
		z, err := thisStream.readUint32BE()
		if err != nil {
			return info, err
		}

		info.X = int32(x)
		info.Y = int32(y)
		info.Z = int32(z)
	} else {
		x, err := thisStream.ReadByte()
		if err != nil {
			return info, err
		}
		y, err := thisStream.ReadByte()
		if err != nil {
			return info, err
		}
		z, err := thisStream.ReadByte()
		if err != nil {
			return info, err
		}

		info.X = int32(int8(x))
		info.Y = int32(int8(y))
		info.Z = int32(int8(z))
	}

	return info, nil
}

func isIn8BitRange(val int32) bool {
	return -127 < val && val < 127
}

// Serialize writes the StreamInfo into a network stream
func (info StreamInfo) Serialize(stream *extendedWriter) error {
	is32Bit := !(isIn8BitRange(info.X) && isIn8BitRange(info.Y) && isIn8BitRange(info.Z))
	err := stream.writeBoolByte(is32Bit)
	if err != nil {
		return err
	}
	if is32Bit {
		err = stream.writeUint32BE(uint32(info.X))
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(info.Y))
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(info.Z))
		if err != nil {
			return err
		}
	} else {
		err = stream.WriteByte(byte(info.X))
		if err != nil {
			return err
		}
		err = stream.WriteByte(byte(info.Y))
		if err != nil {
			return err
		}
		err = stream.WriteByte(byte(info.Z))
		if err != nil {
			return err
		}
	}

	return nil
}

func (info StreamInfo) String() string {
	return fmt.Sprintf("%d, %d, %d", info.X, info.Y, info.Z)
}

// Packet83_0D represents ID_REPLIC_STREAM_DATA
type Packet83_0D struct {
	Bool1 bool
	Bool2 bool

	Region StreamInfo

	Instances []*ReplicationInstance
}

func (thisStream *extendedReader) DecodePacket83_0D(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_0D{}
	var err error
	inner.Bool1, err = thisStream.readBoolByte()
	if err != nil {
		return inner, err
	}
	inner.Bool2, err = thisStream.readBoolByte()
	if err != nil {
		return inner, err
	}

	if !inner.Bool1 && !inner.Bool2 {
		inner.Region, err = thisStream.readStreamInfo()
		if err != nil {
			return inner, err
		}
	}

	joinData, err := thisStream.DecodePacket83_0B(reader, layers)
	if err != nil {
		return inner, err
	}

	inner.Instances = joinData.(*Packet83_0B).Instances

	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_0D) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeBoolByte(layer.Bool1)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.Bool2)
	if err != nil {
		return err
	}

	if !layer.Bool1 && !layer.Bool2 {
		err = layer.Region.Serialize(stream)
		if err != nil {
			return err
		}
	}

	return (&Packet83_0B{Instances: layer.Instances}).Serialize(writer, stream)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_0D) Type() uint8 {
	return 0x0D
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_0D) TypeString() string {
	return "ID_REPLIC_STREAM_DATA"
}

func (layer *Packet83_0D) String() string {
	return fmt.Sprintf("ID_REPLIC_STREAM_DATA: %s, %d instances", layer.Region, len(layer.Instances))
}
