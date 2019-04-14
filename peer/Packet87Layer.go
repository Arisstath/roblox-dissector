package peer

import (
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

type Packet87Layer struct {
	Instance *datamodel.Instance
	Message  string
}

func NewPacket87Layer() *Packet87Layer {
	return &Packet87Layer{}
}

func (thisStream *extendedReader) DecodePacket87Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	context := reader.Context()
	layer := NewPacket87Layer()
	var ref datamodel.Reference

	scope, err := thisStream.readLengthAndString()
	if err != nil {
		return layer, err
	}
	id, err := thisStream.readUint32BE() // Yes, big-endian
	if err != nil {
		return layer, err
	}

	// This reference will never be null
	ref = datamodel.Reference{Scope: scope, Id: id}
	layer.Instance, err = context.InstancesByReference.TryGetInstance(ref)
	if err != nil {
		return layer, err
	}

	layer.Message, err = thisStream.readLengthAndString()
	return layer, err
}

func (layer *Packet87Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeUint32AndString(layer.Instance.Ref.Scope)
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(layer.Instance.Ref.Id)
	if err != nil {
		return err
	}
	return stream.writeUint32AndString(layer.Message)
}

func (layer *Packet87Layer) String() string {
	return fmt.Sprintf("ID_CHAT_ALL: <%s>", layer.Instance.GetFullName())
}

func (Packet87Layer) TypeString() string {
	return "ID_CHAT_ALL"
}

func (Packet87Layer) Type() byte {
	return 0x87
}
