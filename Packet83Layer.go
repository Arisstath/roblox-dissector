package main
import "github.com/google/gopacket"
import "errors"
import "strconv"
import "io"
import "github.com/therecipe/qt/widgets"

var Packet83Subpackets map[uint8]string = map[uint8]string{
	0xFF: "ID_REPLIC_???",
	0x00: "ID_REPLIC_END",
	0x01: "ID_REPLIC_INIT_REFERENT",
	0x02: "ID_REPLIC_NEW_INSTANCE",
	0x03: "ID_REPLIC_PROP",
	0x04: "ID_REPLIC_MARKER",
	0x05: "ID_REPLIC_UNK_05",
	0x07: "ID_REPLIC_EVENT",
	0x0B: "ID_REPLIC_GZIP_JOINDATA",
	0x10: "ID_REPLIC_TAG",
	0x11: "ID_REPLIC_STATS",
}

type packet83Subpacket_ interface {
	Show() widgets.QWidget_ITF
}

type Packet83Subpacket struct {
	child packet83Subpacket_
}

func NewPacket83Subpacket(child packet83Subpacket_) Packet83Subpacket {
	return Packet83Subpacket{child}
}

func (this Packet83Subpacket) Type() uint8 {
	switch this.child.(type) {
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

func (this Packet83Subpacket) TypeString() string {
	return Packet83Subpackets[this.Type()]
}

func (this Packet83Subpacket) Show() widgets.QWidget_ITF {
	return this.child.Show()
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

func DebugInfo(context *CommunicationContext, packet gopacket.Packet) string {
	str := ""
	if context.PacketFromClient(packet) {
		str = "[C->S]"
	} else {
		str = "[S->C]"
	}

	return str
}

func DebugInfo2(context *CommunicationContext, packet gopacket.Packet, isJoinData bool) string {
	if isJoinData {
		return DebugInfo(context, packet) + " J"
	} else {
		return DebugInfo(context, packet)
	}
}

func DecodePacket83Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket83Layer()

	packetType, err := extractPacketType(thisBitstream)
	if err != nil {
		return layer, err
	}
	context.WaitForSchema()
	context.WaitForDescriptors()
	defer context.FinishSchema()
	defer context.FinishDescriptors()
	instanceSchema := context.InstanceSchema

	var inner interface{}

	for packetType != 0 {
		switch packetType {
		case 0x04:
			inner, err = DecodePacket83_04(thisBitstream, context, packet)
			break
		case 0x10:
			inner, err = DecodePacket83_10(thisBitstream, context, packet)
			break
		case 0x05:
			inner, err = DecodePacket83_05(thisBitstream, context, packet)
			break
		case 0x06:
			inner, err = DecodePacket83_05(thisBitstream, context, packet) // Yes, I know it's 05
			break
		case 0x11:
			inner, err = DecodePacket83_11(thisBitstream, context, packet)
			break
		case 0x0B:
			inner, err = DecodePacket83_0B(thisBitstream, context, packet, instanceSchema)
			break
		case 0x02:
			inner, err = DecodePacket83_02(thisBitstream, context, packet, instanceSchema)
			break
		case 0x01:
			inner, err = DecodePacket83_01(thisBitstream, context, packet)
			break
		case 0x03:
			inner, err = DecodePacket83_03(thisBitstream, context, packet, context.PropertySchema)
			break
		//case 0x07:
		//	inner, err = DecodePacket83_07(thisBitstream, context, packet, context.EventSchema)
		//	break
		default:
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		if err != nil {
			return layer, errors.New("parsing subpacket " + Packet83Subpackets[packetType] + ": " + err.Error())
		}

		layer.SubPackets = append(layer.SubPackets, NewPacket83Subpacket(inner.(packet83Subpacket_)))

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
