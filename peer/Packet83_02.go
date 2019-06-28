package peer

import (
	"fmt"
)

// Packet83_02 represents ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	*ReplicationInstance
}

func (thisStream *extendedReader) DecodePacket83_02(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	deferred := newDeferredStrings(reader)
	result, err := decodeReplicationInstance(reader, thisStream, layers, deferred)
	if err != nil {
		return nil, err
	}

	err = thisStream.resolveDeferredStrings(deferred)
	if err != nil {
		return nil, err
	}
	return &Packet83_02{result}, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_02) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return layer.ReplicationInstance.Serialize(writer, stream)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_02) Type() uint8 {
	return 2
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_02) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}

func (layer *Packet83_02) String() string {
	return fmt.Sprintf("ID_REPLIC_NEW_INSTANCE: %s (%s)", layer.Instance.GetFullName(), layer.Schema.Name)
}
