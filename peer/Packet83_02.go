package peer
import "github.com/gskartwii/rbxfile"

type Packet83_02 struct {
	Child *rbxfile.Instance
}

func DecodePacket83_02(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	result, err := DecodeReplicationInstance(false, packet, context)
	return &Packet83_02{result}, err
}
