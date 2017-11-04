package peer
import "errors"

type Packet83_09Subpacket interface{}

type Packet83_09 struct {
	Subpacket Packet83_09Subpacket
	Type uint8
}

type Packet83_09_01 struct {
	Int1 uint8
	Int2 uint32
	Int3 uint32
	Int4 uint32
	Int5 uint64
}

type Packet83_09_05 struct {
	Int uint32
}

type Packet83_09_default struct {
	Int1 uint8
	Int2 uint32
}

type Packet83_09_07 struct{}

func DecodePacket83_09(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_09{}
	thisBitstream := packet.Stream
	inner.Type, err = thisBitstream.ReadUint8()
	if err != nil {
		return inner, err
	}
	var subpacket interface{}
	switch inner.Type {
	case 1:
		thisSubpacket := Packet83_09_01{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int3, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int4, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int5, err = thisBitstream.ReadUint64BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 5:
		thisSubpacket := Packet83_09_05{}
		thisSubpacket.Int, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 7:
		thisSubpacket := Packet83_09_07{}
		subpacket = thisSubpacket
	default:
		thisSubpacket := Packet83_09_default{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	}
	inner.Subpacket = subpacket

	return inner, err
}

func (layer *Packet83_09) Serialize(isClient bool, context *CommunicationContext, stream *ExtendedWriter) error {
    return errors.New("packet 83_09 not implemented!")
}
