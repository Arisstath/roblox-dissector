package peer

import (
	"github.com/gskartwii/rbxfile"
)

type OldChatPacket struct {
	Instance *rbxfile.Instance
	Message  string
}

func NewOldChatPacket() *OldChatPacket {
	return &OldChatPacket{}
}

func (thisBitstream *extendedReader) DecodeOldChatPacket(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	context := reader.Context()
	layer := NewOldChatPacket()
	var ref Referent

	scope, err := thisBitstream.readLengthAndString()
	if err != nil {
		return layer, err
	}
	id, err := thisBitstream.readUint32BE() // Yes, big-endian
	if err != nil {
		return layer, err
	}

	ref = objectToRef(scope, id)
	layer.Instance, err = context.InstancesByReferent.TryGetInstance(ref)
	if err != nil {
		return layer, err
	}

	layer.Message, err = thisBitstream.readLengthAndString()
	return layer, err
}

func (layer *OldChatPacket) Serialize(writer PacketWriter, stream *extendedWriter) error {
	scope, id := refToObject(Referent(layer.Instance.Reference))
	err := stream.writeUint32AndString(scope)
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(id)
	if err != nil {
		return err
	}
	return stream.writeUint32AndString(layer.Message)
}
