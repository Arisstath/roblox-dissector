package peer

import (
	"errors"
	"io"
	"strconv"
)

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
	0x09: "ID_REPLIC_ROCKY",
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
type Packet83Subpacket interface {
	Serialize(writer PacketWriter, stream *extendedWriter) error
	Type() uint8
	TypeString() string
}

// ID_DATA - client <-> server
type Packet83Layer struct {
	SubPackets []Packet83Subpacket
}

func NewPacket83Layer() *Packet83Layer {
	return &Packet83Layer{}
}

func extractPacketType(stream *extendedReader) (uint8, error) {
	return stream.readUint8()
}

func DecodePacket83Layer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	layer := NewPacket83Layer()
	thisBitstream := packet.stream

	packetType, err := extractPacketType(thisBitstream)
	if err != nil {
		return layer, err
	}

	var inner interface{}

	for packetType != 0 {
		//println("parsing subpacket", packetType)
		switch packetType {
		case 0x04:
			inner, err = DecodePacket83_04(reader, packet)
			break
		case 0x10:
			inner, err = DecodePacket83_10(reader, packet)
			break
		case 0x05:
			inner, err = DecodePacket83_05(reader, packet)
			break
		case 0x06:
			inner, err = DecodePacket83_06(reader, packet)
			break
		case 0x0B:
			inner, err = DecodePacket83_0B(reader, packet)
			break
		case 0x02:
			inner, err = DecodePacket83_02(reader, packet)
			break
		case 0x01:
			inner, err = DecodePacket83_01(reader, packet)
			break
		case 0x03:
			inner, err = DecodePacket83_03(reader, packet)
			break
		case 0x07:
			inner, err = DecodePacket83_07(reader, packet)
			break
		case 0x12:
			inner, err = DecodePacket83_12(reader, packet)
			break
		case 0x09:
			inner, err = DecodePacket83_09(reader, packet)
			break
		case 0x0A:
			inner, err = DecodePacket83_0A(reader, packet)
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

func (layer *Packet83Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x83)
	if err != nil {
		return err
	}
	for _, subpacket := range layer.SubPackets {
		thisType := subpacket.Type()
		stream.WriteByte(uint8(thisType))
		if err != nil {
			return err
		}
		err = subpacket.Serialize(writer, stream)
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0)
}
