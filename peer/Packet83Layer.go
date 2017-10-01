package peer
import "errors"
import "strconv"
import "io"

var Packet83Subpackets map[uint8]string = map[uint8]string{
	0xFF: "ID_REPLIC_???",
	0x00: "ID_REPLIC_END",
	0x01: "ID_REPLIC_DELETE_INSTANCE",
	0x02: "ID_REPLIC_NEW_INSTANCE",
	0x03: "ID_REPLIC_PROP",
	0x04: "ID_REPLIC_MARKER",
	0x05: "ID_REPLIC_PING",
	0x07: "ID_REPLIC_EVENT",
	0x09: "ID_REPLIC_UNK_09",
	0x0B: "ID_REPLIC_GZIP_JOINDATA",
	0x10: "ID_REPLIC_TAG",
	0x11: "ID_REPLIC_STATS",
}

type Packet83Subpacket interface{
    Serialize(*CommunicationContext, *ExtendedWriter) error
}

func Packet83ToType(this Packet83Subpacket) uint8 {
	switch this.(type) {
		case *Packet83_01:
			return 1
		case *Packet83_02:
			return 2
		case *Packet83_03:
			return 3
		case *Packet83_04:
			return 4
		case *Packet83_05:
			return 5
		case *Packet83_07:
			return 7
		case *Packet83_0B:
			return 0xB
		case *Packet83_10:
			return 0x10
		case *Packet83_11:
			return 0x11
		default:
			return 0xFF
	}
}

func Packet83ToTypeString(this Packet83Subpacket) string {
	return Packet83Subpackets[Packet83ToType(this)]
}

type Packet83Layer struct {
	SubPackets []Packet83Subpacket
}

func NewPacket83Layer() Packet83Layer {
	return Packet83Layer{}
}

func extractPacketType(stream *ExtendedReader) (uint8, error) {
	ret, err := stream.Bits(2)
	if err != nil {
		return 0, err
	} else if ret != 0 {
		return uint8(ret), err
	}

	ret, err = stream.Bits(5)
	if err != nil {
		return 0, err
	}
	return uint8(ret), err
}

func DecodePacket83Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket83Layer()
	thisBitstream := packet.Stream

	packetType, err := extractPacketType(thisBitstream)
	if err != nil {
		return layer, err
	}
	context.WaitForSchema()
	defer context.FinishSchema()

	var inner interface{}

	for packetType != 0 {
		switch packetType {
		case 0x04:
			inner, err = DecodePacket83_04(packet, context)
			break
		case 0x10:
			inner, err = DecodePacket83_10(packet, context)
			break
		case 0x05:
			inner, err = DecodePacket83_05(packet, context)
			break
		case 0x06:
			inner, err = DecodePacket83_05(packet, context) // Yes, I know it's 05
			break
		case 0x12:
			inner, err = DecodePacket83_11(packet, context)
			break
		case 0x0B:
			inner, err = DecodePacket83_0B(packet, context)
			break
		case 0x02:
			inner, err = DecodePacket83_02(packet, context)
			break
		case 0x01:
			inner, err = DecodePacket83_01(packet, context)
			break
		case 0x03:
			inner, err = DecodePacket83_03(packet, context)
			break
		case 0x07:
			inner, err = DecodePacket83_07(packet, context)
			break
		case 0x09:
			inner, err = DecodePacket83_09(packet, context)
			break
		default:
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		if err != nil {
			return layer, errors.New("parsing subpacket " + Packet83Subpackets[packetType] + ": " + err.Error())
		}

		layer.SubPackets = append(layer.SubPackets, inner.(Packet83Subpacket))

		packetType, err = extractPacketType(thisBitstream)
		if err == io.EOF {
			return layer, nil
		}
		if err != nil {
			return layer, err
		}
	}
	return layer, nil
}

func (layer *Packet83Layer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
    var err error
    err = stream.WriteByte(0x83)
    if err != nil {
        return err
    }
    for _, subpacket := range layer.SubPackets {
        thisType := Packet83ToType(subpacket)
        if thisType < 4 {
            err = stream.Bits(2, uint64(thisType))
        } else {
            err = stream.Bits(2, 0)
            if err != nil {
                return err
            }
            err = stream.Bits(5, uint64(thisType))
        }
        if err != nil {
            return err
        }
        err = subpacket.Serialize(context, stream)
        if err != nil {
            return err
        }
    }
    return stream.Bits(2, 0)
}
