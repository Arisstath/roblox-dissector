package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet83_13 represents ID_REPLIC_ATOMIC
type Packet83_13 struct {
	Instance *datamodel.Instance
	Parent   *datamodel.Instance
}

func (thisStream *extendedReader) DecodePacket83_13(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_13{}

	ref1, err := thisStream.readObject(reader.Context())
	if err != nil {
		return inner, err
	}
	inner.Instance, err = reader.Context().InstancesByReference.TryGetInstance(ref1)
	if err != nil {
		return inner, err
	}

	ref2, err := thisStream.readObject(reader.Context())
	if err != nil {
		return inner, err
	}
	inner.Parent, err = reader.Context().InstancesByReference.TryGetInstance(ref2)
	if err != nil && err != datamodel.ErrNullInstance {
		return inner, err
	}

	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_13) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeObject(layer.Instance, writer.Context())
	if err != nil {
		return err
	}
	return stream.writeObject(layer.Parent, writer.Context())
}

// Type implements Packet83Subpacket.Type()
func (Packet83_13) Type() uint8 {
	return 0x13
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_13) TypeString() string {
	return "ID_REPLIC_ATOMIC"
}

func (layer *Packet83_13) String() string {
	return fmt.Sprintf("ID_REPLIC_ATOMIC: %s: %s parented to %s: %s", layer.Instance.Ref.String(), layer.Instance.Name(), layer.Parent.Ref.String(), layer.Parent.Name())
}
