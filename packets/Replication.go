package packets

import (
	"errors"
	"io"
	"strconv"
)
import "github.com/gskartwii/roblox-dissector/util"

// List of string names for all 0x83 subpackets
var ReplicationSubpackets map[uint8]string = map[uint8]string{
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

var ReplicationDecoders = map[uint8](func(*PacketReaderBitstream, util.PacketReader, *PacketLayers) (ReplicationSubpacket, error)){
	0x01: (*PacketReaderBitstream).DecodeDeleteInstance,
	0x02: (*PacketReaderBitstream).DecodeNewInstance,
	0x03: (*PacketReaderBitstream).DecodeChangeProperty,
	0x04: (*PacketReaderBitstream).DecodeReplicationMarker,
	0x05: (*PacketReaderBitstream).DecodeDataPing,
	0x06: (*PacketReaderBitstream).DecodeDataPingBack,
	0x07: (*PacketReaderBitstream).DecodeReplicateEvent,
	0x09: (*PacketReaderBitstream).DecodeReplicRocky,
	0x0A: (*PacketReaderBitstream).DecodeAckProperty,
	0x0B: (*PacketReaderBitstream).DecodeReplicateJoinData,
	0x10: (*PacketReaderBitstream).DecodeReplicationTag,
	0x11: (*PacketReaderBitstream).DecodeStats,
	0x12: (*PacketReaderBitstream).DecodeReplicateHash,
}

// A subpacket contained within a 0x83 (ID_DATA) packet
type ReplicationSubpacket interface {
	Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error
	Type() uint8
	TypeString() string
}

// ID_DATA - client <-> server
type ReplicatorPacket struct {
	SubPackets []ReplicationSubpacket
}

func NewReplicatorPacket() *ReplicatorPacket {
	return &ReplicatorPacket{}
}

func (thisBitstream *PacketReaderBitstream) DecodeReplicatorPacket(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewReplicatorPacket()

	packetType, err := thisBitstream.ReadUint8()
	if err != nil {
		return layer, err
	}

	var inner ReplicationSubpacket
	for packetType != 0 {
		//println("parsing subpacket", packetType)
		decoder, ok := ReplicationDecoders[packetType]
		if !ok {
			return layer, errors.New("don't know how to parse replication subpacket: " + strconv.Itoa(int(packetType)))
		}
		inner, err = decoder(thisBitstream, reader, layers)
		if err != nil {
			return layer, errors.New("parsing subpacket " + ReplicationSubpackets[packetType] + ": " + err.Error())
		}

		layer.SubPackets = append(layer.SubPackets, inner)

		packetType, err = thisBitstream.ReadUint8()
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

func (layer *ReplicatorPacket) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
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
