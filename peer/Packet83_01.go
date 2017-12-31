package peer
import "github.com/gskartwii/rbxfile"

// ID_DELETE_INSTANCE
type Packet83_01 struct {
	// Instance to be deleted
	Instance *rbxfile.Instance
}

func decodePacket83_01(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_01{}
	thisBitstream := packet.stream
    referent, err := thisBitstream.readObject(context.IsClient(packet.Source), false, context)
    inner.Instance = context.InstancesByReferent.TryGetInstance(referent)
	inner.Instance.SetParent(nil)

	return inner, err
}

func (layer *Packet83_01) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    return stream.writeObject(isClient, layer.Instance, false, context)
}
