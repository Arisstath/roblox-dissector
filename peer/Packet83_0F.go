package peer

import (
	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet83_0F represents ID_REPLIC_INSTANCE_REMOVAL
// How is this different from ID_REPLIC_DELETE_INSTANCE?
// Does the latter force GC?
// Is this explicitly for streaming?
type Packet83_0F struct {
	Instance *datamodel.Instance
}

func (thisStream *extendedReader) DecodePacket83_0F(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_0F{}

	reference, err := thisStream.ReadObject(reader)
	if err != nil {
		return inner, err
	}
	inner.Instance, err = reader.Context().InstancesByReference.TryGetInstance(reference)

	return inner, err
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_0F) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return stream.WriteObject(layer.Instance, writer)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_0F) Type() uint8 {
	return 0xF
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_0F) TypeString() string {
	return "ID_REPLIC_INSTANCE_REMOVAL"
}

func (layer *Packet83_0F) String() string {
	return "ID_REPLIC_INSTANCE_REMOVAL: " + layer.Instance.GetFullName()
}
