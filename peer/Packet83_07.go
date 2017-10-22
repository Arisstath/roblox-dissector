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
	isClient := context.IsClient(packet.Source)
	layer := &Packet83_07{}
	thisBitstream := packet.Stream
    referent, err := thisBitstream.ReadObject(isClient, false, context)
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

func (layer *Packet83_07) Serialize(isClient bool, context *CommunicationContext, stream *ExtendedWriter) error {
	err := stream.WriteObject(isClient, layer.Instance, false, context)
	if err != nil {
		return err
	}

	eventSchemaID, ok := context.StaticSchema.EventsByName[layer.Instance.ClassName + "." + layer.EventName]
	if !ok {
		return errors.New("Invalid event: " + layer.Instance.ClassName + "." + layer.EventName)
	}
	err = stream.WriteUint16BE(uint16(eventSchemaID))
	if err != nil {
		return err
	}

	schema := context.StaticSchema.Events[uint16(eventSchemaID)]
	println("Writing event", schema.Name, schema.InstanceSchema.Name)

	return schema.Serialize(isClient, layer.Event, context, stream)
}
