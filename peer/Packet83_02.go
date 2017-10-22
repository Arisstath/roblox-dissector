package peer
import "github.com/gskartwii/rbxfile"

type Packet83_02 struct {
	Child *rbxfile.Instance
}

func DecodePacket83_02(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	result, err := DecodeReplicationInstance(context.IsClient(packet.Source), false, packet, context)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) Serialize(isClient bool, context *CommunicationContext, stream *ExtendedWriter) error {
    return SerializeReplicationInstance(isClient, layer.Child, false, context, stream)
}
