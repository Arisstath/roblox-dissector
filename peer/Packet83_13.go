package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// ID_REPLIC_ATOMIC
type Packet83_13 struct {
	Instance *datamodel.Instance
	Parent   *datamodel.Instance
}

func (thisStream *extendedReader) DecodePacket83_13(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_13{}

	ref1, err := thisStream.ReadObject(reader)
	if err != nil {
		return inner, err
	}
	inner.Instance, err = reader.Context().InstancesByReference.TryGetInstance(ref1)
	if err != nil {
		return inner, err
	}

	ref2, err := thisStream.ReadObject(reader)
	if err != nil {
		return inner, err
	}
	inner.Parent, err = reader.Context().InstancesByReference.TryGetInstance(ref2)
	if err != nil && err != datamodel.ErrNullInstance {
		return inner, err
	}

	return inner, nil
}

func (layer *Packet83_13) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteObject(layer.Instance, writer)
	if err != nil {
		return err
	}
	return stream.WriteObject(layer.Parent, writer)
}

func (Packet83_13) Type() uint8 {
	return 0x13
}
func (Packet83_13) TypeString() string {
	return "ID_REPLIC_ATOMIC"
}

func (layer *Packet83_13) String() string {
	return fmt.Sprintf("ID_REPLIC_ATOMIC: %s", layer.Instance.GetFullName())
}
