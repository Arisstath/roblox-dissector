package peer
import "errors"
import "strconv"
import "io"

// List of string names for all 0x83 subpackets
var Packet83Subpackets map[uint8]string = map[uint8]string{
	0xFF: "ID_REPLIC_???",
	0x00: "ID_REPLIC_END",
	0x01: "ID_REPLIC_DELETE_INSTANCE",
	0x02: "ID_REPLIC_NEW_INSTANCE",
	0x03: "ID_REPLIC_PROP",
	0x04: "ID_REPLIC_MARKER",
	0x05: "ID_REPLIC_PING",
	0x06: "ID_REPLIC_PING_BACK",
	0x07: "ID_REPLIC_EVENT",
	0x08: "ID_REPLIC_REQUEST_CHAR",
	0x09: "ID_REPLIC_CHEATER",
	0x0A: "ID_REPLIC_PROP_ACK",
	0x0B: "ID_REPLIC_GZIP_JOINDATA",
	0x0C: "ID_REPLIC_UPDATE_CLIENT_QUOTA",
	0x0D: "ID_REPLIC_STREAM_DATA",
	0x0E: "ID_REPLIC_REGION_REMOVAL",
	0x0F: "ID_REPLIC_INSTANCE_REMOVAL",
	0x10: "ID_REPLIC_TAG",
	0x11: "ID_REPLIC_STATS",
	0x12: "ID_REPLIC_HASH",
}

// A subpacket contained within a 0x83 (ID_DATA) packet
type Packet83Subpacket interface{
    serialize(bool, *CommunicationContext, *extendedWriter) error
}

// Extracts a type identifier from a packet
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
		case *Packet83_06:
			return 6
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

// Looks up a string name for a packet
func Packet83ToTypeString(this Packet83Subpacket) string {
	return Packet83Subpackets[Packet83ToType(this)]
}

// ID_DATA - client <-> server
type Packet83Layer struct {
	SubPackets []Packet83Subpacket
}

func NewPacket83Layer() *Packet83Layer {
	return &Packet83Layer{}
}

func extractPacketType(stream *extendedReader) (uint8, error) {
	ret, err := stream.bits(2)
	if err != nil {
		return 0, err
	} else if ret != 0 {
		return uint8(ret), err
	}

	ret, err = stream.bits(5)
	if err != nil {
		return 0, err
	}
	return uint8(ret), err
}

func decodePacket83Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket83Layer()
	thisBitstream := packet.stream

	packetType, err := extractPacketType(thisBitstream)
	if err != nil {
		return layer, err
	}

	var inner interface{}

	for packetType != 0 {
		switch packetType {
		case 0x04:
			inner, err = decodePacket83_04(packet, context)
			break
		case 0x10:
			inner, err = decodePacket83_10(packet, context)
			break
		case 0x05:
			inner, err = decodePacket83_05(packet, context)
			break
		case 0x06:
			inner, err = decodePacket83_06(packet, context)
			break
		case 0x0B:
			inner, err = decodePacket83_0B(packet, context)
			break
		case 0x02:
			inner, err = decodePacket83_02(packet, context)
			break
		case 0x01:
			inner, err = decodePacket83_01(packet, context)
			break
		case 0x03:
			inner, err = decodePacket83_03(packet, context)
			break
		case 0x07:
			inner, err = decodePacket83_07(packet, context)
			break
		case 0x09:
			inner, err = decodePacket83_09(packet, context)
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

func (layer *Packet83Layer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    var err error
    err = stream.WriteByte(0x83)
    if err != nil {
        return err
    }
    for _, subpacket := range layer.SubPackets {
        thisType := Packet83ToType(subpacket)
        if thisType < 4 {
            err = stream.bits(2, uint64(thisType))
        } else {
            err = stream.bits(2, 0)
            if err != nil {
                return err
            }
            err = stream.bits(5, uint64(thisType))
        }
        if err != nil {
            return err
        }
        err = subpacket.serialize(isClient, context, stream)
        if err != nil {
            return err
        }
    }
    return stream.bits(2, 0)
}
