package peer

import (
	"errors"
	"fmt"
)

// Packet83_09Subpacket is a generic interface for Packet83_09 subpackets
type Packet83_09Subpacket interface {
	fmt.Stringer
	Type() byte
	TypeString() string
}

// Packet83_09 represents ID_ROCKY
type Packet83_09 struct {
	Subpacket Packet83_09Subpacket
}

// Packet83_09_00 represents a NetPMC packet the purpose of which is unknown
type Packet83_09_00 struct {
	Int1 uint32
	Int2 uint32
	Int3 uint32
	Int4 uint32
	Int5 uint32
}

// Packet83_09_04 represents a NetPMC packet the purpose of which is unknown
type Packet83_09_04 struct {
	Int1 byte
	Int2 uint32
}

// Packet83_09_05 represents an ID Challenge packet
type Packet83_09_05 struct {
	Challenge uint32
}

// Packet83_09_06 represents a response to an ID Challenge packet
type Packet83_09_06 struct {
	Challenge uint32
	Response  uint32
}

func (thisStream *extendedReader) DecodePacket83_09(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_09{}

	subpacketType, err := thisStream.readUint8()
	if err != nil {
		return inner, err
	}
	var subpacket Packet83_09Subpacket
	switch subpacketType {
	case 0: // ???
		thisSubpacket := &Packet83_09_00{}
		thisSubpacket.Int1, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int3, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int4, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int5, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 4: // ???
		thisSubpacket := &Packet83_09_04{}
		thisSubpacket.Int1, err = thisStream.ReadByte()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 5: // id challenge
		thisSubpacket := &Packet83_09_05{}
		thisSubpacket.Challenge, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 6: // id response
		thisSubpacket := &Packet83_09_06{}
		thisSubpacket.Challenge, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Response, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	default:
		layers.Root.Logger.Println("don't know rocky subpacket", subpacketType)
		return inner, errors.New("unimplemented subpacket type")
	}
	inner.Subpacket = subpacket

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_09) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	switch layer.Subpacket.(type) {
	// TODO: implement unknown subpackets
	case (*Packet83_09_00):
		subpacket := layer.Subpacket.(*Packet83_09_00)
		err = stream.WriteByte(0)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int2)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int3)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int4)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int5)
		if err != nil {
			return err
		}
	case (*Packet83_09_04):
		subpacket := layer.Subpacket.(*Packet83_09_04)
		err = stream.WriteByte(4)
		if err != nil {
			return err
		}
		err = stream.WriteByte(subpacket.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int2)
		if err != nil {
			return err
		}
	case (*Packet83_09_05):
		subpacket := layer.Subpacket.(*Packet83_09_05)
		err = stream.WriteByte(5)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Challenge)
		if err != nil {
			return err
		}
	case (*Packet83_09_06):
		subpacket := layer.Subpacket.(*Packet83_09_06)
		err = stream.WriteByte(6)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Challenge)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Response)
		if err != nil {
			return err
		}
	default:
		fmt.Printf("Tried to write rocky packet %T\n", layer.Subpacket)
		return errors.New("rocky packet not implemented")
	}

	return err
}

// Type implements Packet83Subpacket.Type()
func (Packet83_09) Type() uint8 {
	return 9
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_09) TypeString() string {
	return "ID_REPLIC_ROCKY"
}

func (layer *Packet83_09) String() string {
	return fmt.Sprintf("ID_REPLIC_ROCKY: %s", layer.Subpacket.String())
}

func (layer *Packet83_09_00) String() string {
	return fmt.Sprintf("0x00: %d, %d, %d, %d, %d", layer.Int1, layer.Int2, layer.Int3, layer.Int4, layer.Int5)
}

// Type implements Packet83_09Subpacket.Type()
func (Packet83_09_00) Type() byte {
	return 0
}

// TypeString implements Packet83_09Subpacket.TypeString()
func (Packet83_09_00) TypeString() string {
	return "ID_ROCKY_UNKNOWN_0"
}
func (layer *Packet83_09_04) String() string {
	return fmt.Sprintf("0x04: %d, %d", layer.Int1, layer.Int2)
}

// Type implements Packet83_09Subpacket.Type()
func (Packet83_09_04) Type() byte {
	return 4
}

// TypeString implements Packet83_09Subpacket.TypeString()
func (Packet83_09_04) TypeString() string {
	return "ID_ROCKY_UNKNOWN_4"
}
func (layer *Packet83_09_05) String() string {
	return fmt.Sprintf("IdChallenge %08X", layer.Challenge)
}

// Type implements Packet83_09Subpacket.Type()
func (Packet83_09_05) Type() byte {
	return 5
}

// TypeString implements Packet83_09Subpacket.TypeString()
func (Packet83_09_05) TypeString() string {
	return "ID_ROCKY_ID_CHALLENGE"
}
func (layer *Packet83_09_06) String() string {
	return fmt.Sprintf("IdResponse %08X->%08X", layer.Challenge, layer.Response)
}

// Type implements Packet83_09Subpacket.Type()
func (Packet83_09_06) Type() byte {
	return 6
}

// TypeString implements Packet83_09Subpacket.TypeString()
func (Packet83_09_06) TypeString() string {
	return "ID_ROCKY_ID_RESPONSE"
}
