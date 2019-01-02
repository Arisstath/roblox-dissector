package peer

import (
	"errors"
	"fmt"
)

type Packet83_09Subpacket interface {
	fmt.Stringer
}

type Packet83_09 struct {
	Subpacket     Packet83_09Subpacket
	SubpacketType uint8
}

type Packet83_09_05 struct {
	Int uint32
}

type Packet83_09_06 struct {
	Int1 uint32
	Int2 uint32
}

func (thisBitstream *extendedReader) DecodePacket83_09(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_09{}

	inner.SubpacketType, err = thisBitstream.readUint8()
	if err != nil {
		return inner, err
	}
	var subpacket Packet83_09Subpacket
	switch inner.SubpacketType {
	case 5: // id challenge
		thisSubpacket := &Packet83_09_05{}
		thisSubpacket.Int, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 6: // id response
		thisSubpacket := &Packet83_09_06{}
		thisSubpacket.Int1, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	default:
		layers.Root.Logger.Println("don't know rocky subpacket", inner.Type())
		return inner, errors.New("unimplemented subpacket type")
	}
	inner.Subpacket = subpacket

	return inner, err
}

func (layer *Packet83_09) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(layer.SubpacketType)
	if err != nil {
		return err
	}

	switch layer.SubpacketType {
	case 5:
		subpacket := layer.Subpacket.(*Packet83_09_05)
		err = stream.writeUint32BE(subpacket.Int)
	case 6:
		subpacket := layer.Subpacket.(*Packet83_09_06)
		err = stream.writeUint32BE(subpacket.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int2)
	default:
		println("Tried to write rocky packet", layer.Type())
		return errors.New("rocky packet not implemented")
	}

	return err
}

func (Packet83_09) Type() uint8 {
	return 9
}
func (Packet83_09) TypeString() string {
	return "ID_REPLIC_ROCKY"
}

func (layer *Packet83_09) String() string {
	return fmt.Sprintf("ID_REPLIC_ROCKY: %s", layer.Subpacket.String())
}

func (Packet83_09_05) String() string {
	return "IdChallenge"
}
func (Packet83_09_06) String() string {
	return "IdResponse"
}
