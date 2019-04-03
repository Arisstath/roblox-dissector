package peer

import "fmt"

// ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	*ReplicationInstance
}

func (thisStream *extendedReader) DecodePacket83_02(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	result, err := decodeReplicationInstance(reader, thisStream, layers)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return layer.ReplicationInstance.Serialize(writer, stream)
}

func (Packet83_02) Type() uint8 {
	return 2
}
func (Packet83_02) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}

func (layer *Packet83_02) String() string {
	return fmt.Sprintf("ID_REPLIC_NEW_INSTANCE: %s (%s)", layer.Instance.GetFullName(), layer.Schema.Name)
}
