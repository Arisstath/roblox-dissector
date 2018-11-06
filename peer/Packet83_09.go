package peer
import "errors"

type Packet83_09Subpacket interface{}

type Packet83_09 struct {
	Subpacket Packet83_09Subpacket
	Type uint8
}

type Packet83_09_00 struct {
	Values [5]uint32
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

type Packet83_09_06 struct {
	Int1 uint32
	Int2 uint32
}

type Packet83_09_07 struct{}

func DecodePacket83_09(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_09{}
	thisBitstream := packet.stream
	inner.Type, err = thisBitstream.readUint8()
	if err != nil {
		return inner, err
	}
	var subpacket interface{}
	switch inner.Type {
	case 0: // Rocky
		thisSubpacket := &Packet83_09_00{}
		for i := 0; i < 5; i++ {
			thisSubpacket.Values[i], err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
		}
		subpacket = thisSubpacket
	case 1:
		thisSubpacket := &Packet83_09_01{}
		thisSubpacket.Int1, err = thisBitstream.readUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int3, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int4, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int5, err = thisBitstream.readUint64BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 2: // net pmc response
		
	case 5:
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
	case 7:
		thisSubpacket := &Packet83_09_07{}
		subpacket = thisSubpacket
	default:
		packet.Logger.Println("don't know rocky subpacket", inner.Type)
		return inner, errors.New("unimplemented subpacket type")
	}
	inner.Subpacket = subpacket

	return inner, err
}

func (layer *Packet83_09) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(layer.Type)
	if err != nil {
		return err
	}

	switch layer.Type {
	case 6:
		subpacket := layer.Subpacket.(*Packet83_09_06)
		err = stream.writeUint32BE(subpacket.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(subpacket.Int2)
		break
	default:
		println("Tried to write rocky packet", layer.Type)
		return errors.New("rocky packet not implemented!")
	}

	return err
}
