package peer
import "github.com/gskartwii/rbxfile"

// ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	Child *rbxfile.Instance
}

func decodePacket83_02(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	result, err := decodeReplicationInstance(context.IsClient(packet.Source), false, packet, context)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    return serializeReplicationInstance(isClient, layer.Child, false, context, stream)
}
