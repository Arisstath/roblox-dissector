package peer

import (
	"errors"
	"fmt"
	"strconv"
)

// Packet83Subpackets containts a list of string names for all 0x83 subpackets
var Packet83Subpackets = map[uint8]string{
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
	0x0A: "ID_REPLIC_CFRAME_ACK",
	0x0B: "ID_REPLIC_JOIN_DATA",
	0x0C: "ID_REPLIC_UPDATE_CLIENT_QUOTA",
	0x0D: "ID_REPLIC_STREAM_DATA",
	0x0E: "ID_REPLIC_REGION_REMOVAL",
	0x0F: "ID_REPLIC_INSTANCE_REMOVAL",
	0x10: "ID_REPLIC_TAG",
	0x11: "ID_REPLIC_STATS",
	0x12: "ID_REPLIC_HASH",
	0x13: "ID_REPLIC_ATOMIC",
	0x14: "ID_REPLIC_STREAM_DATA_INFO",
}

var packet83Decoders = map[uint8](func(*extendedReader, PacketReader, *PacketLayers) (Packet83Subpacket, error)){
	0x01: (*extendedReader).DecodePacket83_01,
	0x02: (*extendedReader).DecodePacket83_02,
	0x03: (*extendedReader).DecodePacket83_03,
	0x04: (*extendedReader).DecodePacket83_04,
	0x05: (*extendedReader).DecodePacket83_05,
	0x06: (*extendedReader).DecodePacket83_06,
	0x07: (*extendedReader).DecodePacket83_07,
	0x09: (*extendedReader).DecodePacket83_09,
	0x0A: (*extendedReader).DecodePacket83_0A,
	0x0B: (*extendedReader).DecodePacket83_0B,
	0x0D: (*extendedReader).DecodePacket83_0D,
	0x0E: (*extendedReader).DecodePacket83_0E,
	0x10: (*extendedReader).DecodePacket83_10,
	0x11: (*extendedReader).DecodePacket83_11,
	0x12: (*extendedReader).DecodePacket83_12,
	0x13: (*extendedReader).DecodePacket83_13,
	0x14: (*extendedReader).DecodePacket83_14,
}

// Packet83Subpacket is an interface implemented by
// subpackets contained within a 0x83 (ID_DATA) packet
type Packet83Subpacket interface {
	fmt.Stringer
	Serialize(writer PacketWriter, stream *extendedWriter) error
	Type() uint8
	TypeString() string
}

// Packet83Layer represents ID_DATA - client <-> server
type Packet83Layer struct {
	SubPackets []Packet83Subpacket
}

func (thisStream *extendedReader) DecodePacket83Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet83Layer{}

	packetType, err := thisStream.readUint8()
	if err != nil {
		return layer, err
	}

	var inner Packet83Subpacket
	for packetType != 0 {
		//println("parsing subpacket", packetType)
		decoder, ok := packet83Decoders[packetType]
		if !ok {
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		inner, err = decoder(thisStream, reader, layers)
		if err != nil {
			return layer, errors.New("parsing subpacket " + Packet83Subpackets[packetType] + ": " + err.Error())
		}

		layer.SubPackets = append(layer.SubPackets, inner)

		packetType, err = thisStream.readUint8()
		if err != nil {
			return layer, err
		}
	}
	return layer, nil
}

// Serialize implements RakNetPacket.Serialize
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

// TypeString implements RakNetPacket.TypeString()
func (Packet83Layer) TypeString() string {
	return "ID_DATA"
}

func (layer *Packet83Layer) String() string {
	return fmt.Sprintf("ID_DATA: %d items", len(layer.SubPackets))
}

// Type implements RakNetPacket.Type()
func (Packet83Layer) Type() byte {
	return 0x83
}
