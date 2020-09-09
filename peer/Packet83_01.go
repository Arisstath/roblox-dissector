package peer

import (
	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet83_01 represents ID_DELETE_INSTANCE
type Packet83_01 struct {
	// Instance to be deleted
	Instance *datamodel.Instance
}

func (thisStream *extendedReader) DecodePacket83_01(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_01{}

	// NULL deletion is actually legal. Who would have known?
	reference, err := thisStream.readObject(reader.Context())
	if err != nil {
		return inner, err
	}
	inner.Instance, err = reader.Context().InstancesByReference.TryGetInstance(reference)

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_01) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.writeObject(layer.Instance, writer.Context())
}

// Type implements Packet83Subpacket.Type()
func (Packet83_01) Type() uint8 {
	return 1
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_01) TypeString() string {
	return "ID_REPLIC_DELETE_INSTANCE"
}

func (layer *Packet83_01) String() string {
	return "ID_REPLIC_DELETE_INSTANCE: " + layer.Instance.Ref.String() + ": " + layer.Instance.Name()
}
