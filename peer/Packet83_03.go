package peer
import "errors"
import "fmt"
import "github.com/gskartwii/rbxfile"

type Packet83_03 struct {
	Instance *rbxfile.Instance
	Bool1 bool
	PropertyName string
	Value rbxfile.Value
}

func DecodePacket83_03(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	layer := &Packet83_03{}
	thisBitstream := packet.Stream
    referent, err := thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

    propertyIDx, err := thisBitstream.ReadUint16BE()
    if err != nil {
        return layer, err
    }

    if int(propertyIDx) == int(len(context.StaticSchema.Properties)) { // explicit Parent property system
		layer.Bool1, err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}

        var referent Referent
        referent, err = thisBitstream.ReadObject(false, context)
        instance := context.InstancesByReferent.TryGetInstance(referent)
		result := rbxfile.ValueReference{instance}
		layer.Value = result

		context.InstancesByReferent.OnAddInstance(referent, func(instance *rbxfile.Instance) {
			result.AddChild(instance)
		})
		return layer, err
    }

    if int(propertyIDx) > int(len(context.StaticSchema.Properties)) {
        return layer, errors.New(fmt.Sprintf("prop idx %d is higher than %d", propertyIDx, len(context.StaticSchema.Properties)))
    }
    schema := context.StaticSchema.Properties[propertyIDx]
    layer.PropertyName = schema.Name

    layer.Bool1, err = thisBitstream.ReadBool()
    if err != nil {
        return layer, err
    }

    layer.Value, err = schema.Decode(ROUND_UPDATE, packet, context)

    context.InstancesByReferent.OnAddInstance(referent, func(instance *rbxfile.Instance) {
        layer.Instance = instance
        instance.Properties[layer.PropertyName] = layer.Value
    })

    return layer, err
}
