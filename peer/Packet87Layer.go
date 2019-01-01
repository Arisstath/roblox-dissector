package peer

import (
	"github.com/gskartwii/roblox-dissector/datamodel"
)

type Packet87Layer struct {
	Instance *datamodel.Instance
	Message  string
}

func NewPacket87Layer() *Packet87Layer {
	return &Packet87Layer{}
}

func (thisBitstream *extendedReader) DecodePacket87Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	context := reader.Context()
	layer := NewPacket87Layer()
	var ref datamodel.Reference

	scope, err := thisBitstream.readLengthAndString()
	if err != nil {
		return layer, err
	}
	id, err := thisBitstream.readUint32BE() // Yes, big-endian
	if err != nil {
		return layer, err
	}

	// This reference will never be null
	ref = datamodel.Reference{Scope: scope, Id: id}
	layer.Instance, err = context.InstancesByReferent.TryGetInstance(ref)
	if err != nil {
		return layer, err
	}

	layer.Message, err = thisBitstream.readLengthAndString()
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
