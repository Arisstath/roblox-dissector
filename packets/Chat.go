package packets

import (
	"github.com/gskartwii/rbxfile"
	"github.com/gskartwii/roblox-dissector/util"
)

type OldChatPacket struct {
	Instance util.DeserializedInstance
	Message  string
}

func NewOldChatPacket() *OldChatPacket {
	return &OldChatPacket{}
}

func (thisBitstream *PacketReaderBitstream) DecodeOldChatPacket(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewOldChatPacket()

	scope, err := thisBitstream.ReadUint16AndString()
	if err != nil {
		return layer, err
	}
	id, err := thisBitstream.ReadUint32BE() // Yes, big-endian
	if err != nil {
		return layer, err
	}

    ref := util.NewReference(scope, id)
	layer.Instance, err = reader.TryGetInstance(ref)
	if err != nil {
		return layer, err
	}

	layer.Message, err = thisBitstream.ReadUint16AndString()
	return layer, err
}

func (layer *OldChatPacket) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteUint16AndString(layer.Instance.Scope)
	if err != nil {
		return err
	}
	err = stream.WriteUint32BE(layer.Instance.Id)
	if err != nil {
		return err
	}
	return stream.WriteUint32AndString(layer.Message)
}
