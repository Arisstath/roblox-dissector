package peer
import "github.com/gskartwii/rbxfile"

type Packet83_02 struct {
	Child *rbxfile.Instance
}

func DecodePacket83_02(packet *UDPPacket, context *CommunicationContext, instanceSchema []*InstanceSchemaItem) (interface{}, error) {
	result, err := DecodeReplicationInstance(false, packet, context, instanceSchema)
	return &Packet83_02{result}, err
}
