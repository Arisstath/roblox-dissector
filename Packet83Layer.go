package main
import "github.com/google/gopacket"
import "errors"
import "strconv"
import "io"
//import "github.com/therecipe/qt/widgets"

type Packet83Subpacket interface {
	//Show() *widgets.QWidget_ITF
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
	defer context.FinishSchema()
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
		default:
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		if err != nil {
			return layer, err
		}

		layer.SubPackets = append(layer.SubPackets, inner)

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
