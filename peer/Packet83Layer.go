package peer

import (
	"errors"
	"fmt"
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

var Packet83Decoders = map[uint8](func(*extendedReader, PacketReader, *PacketLayers) (Packet83Subpacket, error)){
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
	0x10: (*extendedReader).DecodePacket83_10,
	0x11: (*extendedReader).DecodePacket83_11,
	0x12: (*extendedReader).DecodePacket83_12,
}

// A subpacket contained within a 0x83 (ID_DATA) packet
type Packet83Subpacket interface {
	fmt.Stringer
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

func (thisBitstream *extendedReader) DecodePacket83Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket83Layer()

	packetType, err := thisBitstream.readUint8()
	if err != nil {
		return layer, err
	}

	var inner Packet83Subpacket
	for packetType != 0 {
		//println("parsing subpacket", packetType)
		decoder, ok := Packet83Decoders[packetType]
		if !ok {
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		inner, err = decoder(thisBitstream, reader, layers)
		if err != nil {
			return layer, errors.New("parsing subpacket " + Packet83Subpackets[packetType] + ": " + err.Error())
		}

		layer.SubPackets = append(layer.SubPackets, inner)

		packetType, err = thisBitstream.readUint8()
		if err == io.EOF {
			println("DEPRECATED_WARNING: ignoring packettype read past end")
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

func (layer *Packet83Layer) String() string {
	return fmt.Sprintf("ID_REPLICATION_DATA: %d items", len(layer.SubPackets))
}
