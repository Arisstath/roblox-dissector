package peer

import (
	"errors"

	"github.com/gskartwii/rbxfile"
)

// ID_DELETE_INSTANCE
type Packet83_01 struct {
	// Instance to be deleted
	Instance *rbxfile.Instance
}

func DecodePacket83_01(reader PacketReader, packet *UDPPacket) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_01{}
	thisBitstream := packet.stream
	// NULL deletion is actually legal. Who would have known?
	referent, err := thisBitstream.readObject(reader.Caches())
	inner.Instance, err = reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return inner, err
	}

	inner.Instance.SetParent(nil)

	return inner, err
}

func (layer *Packet83_01) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("Instance to delete can't be nil!")
	}
	return stream.writeObject(layer.Instance, writer.Caches())
}
