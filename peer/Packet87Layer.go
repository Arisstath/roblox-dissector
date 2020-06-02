package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet87Layer represents ID_CHAT_ALL
type Packet87Layer struct {
	Instance *datamodel.Instance
	Message  string
}

func (thisStream *extendedReader) DecodePacket87Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	context := reader.Context()
	layer := &Packet87Layer{}
	var ref datamodel.Reference

	peerID, err := thisStream.readVarint64()
	if err != nil {
		return layer, err
	}
	id, err := thisStream.readUint32BE() // Yes, big-endian
	if err != nil {
		return layer, err
	}

	// This reference will never be null
	ref = datamodel.Reference{Scope: fmt.Sprintf("RBXPID%d", peerID), Id: id, PeerId: uint32(peerID)}
	layer.Instance, err = context.InstancesByReference.TryGetInstance(ref)
	if err != nil {
		return layer, err
	}

	messageLen, err := thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.Message, err = thisStream.readASCII(int(messageLen))
	return layer, err
}

// Serialize implements RakNetPacket.Serialize
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

// TypeString implements RakNetPacket.TypeString()
func (Packet87Layer) TypeString() string {
	return "ID_CHAT_ALL"
}

// Type implements RakNetPacket.Type()
func (Packet87Layer) Type() byte {
	return 0x87
}
