package peer
import "errors"
import "fmt"
import "github.com/gskartwii/rbxfile"

type Packet83_07 struct {
	Instance *rbxfile.Instance
	EventName string
	Event *ReplicationEvent
}

func DecodePacket83_07(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	layer := &Packet83_07{}
	thisBitstream := packet.Stream
    referent, err := thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

    eventIDx, err := thisBitstream.ReadUint16BE()
    if err != nil {
        return layer, err
    }

    if int(eventIDx) > int(len(context.StaticSchema.Events)) {
        return layer, errors.New(fmt.Sprintf("event idx %d is higher than %d", eventIDx, len(context.StaticSchema.Events)))
    }

    schema := context.StaticSchema.Events[eventIDx]
    layer.EventName = schema.Name
    layer.Event, err = schema.Decode(packet, context)
    context.InstancesByReferent.OnAddInstance(referent, func(instance *rbxfile.Instance) {
        layer.Instance = instance
    })
    return layer, err
}
