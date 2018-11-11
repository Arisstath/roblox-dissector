package peer

import (
	"errors"

	"github.com/gskartwii/rbxfile"
)

// ID_DELETE_INSTANCE
type DeleteInstance struct {
	// Instance to be deleted
	Instance *rbxfile.Instance
}

func (thisBitstream *extendedReader) DecodeDeleteInstance(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &DeleteInstance{}

	// NULL deletion is actually legal. Who would have known?
	referent, err := thisBitstream.readObject(reader.Caches())
	inner.Instance, err = reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return inner, err
	}

	inner.Instance.SetParent(nil)

	return inner, err
}

func (layer *DeleteInstance) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("Instance to delete can't be nil!")
	}
	return stream.writeObject(layer.Instance, writer.Caches())
}

func (DeleteInstance) Type() uint8 {
	return 1
}
func (DeleteInstance) TypeString() string {
	return "ID_REPLIC_DELETE_INSTANCE"
}
