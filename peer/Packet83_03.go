package peer
import "errors"
import "fmt"
import "github.com/gskartwii/rbxfile"

// ID_CHANGE_PROPERTY
type Packet83_03 struct {
	// Instance that had the property change
	Instance *rbxfile.Instance
	Bool1 bool
	// Name of the property
	PropertyName string
	// New value
	Value rbxfile.Value
}

func decodePacket83_03(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	isClient := context.IsClient(packet.Source)

	layer := &Packet83_03{}
	thisBitstream := packet.stream
    referent, err := thisBitstream.readObject(isClient, false, context)
	if err != nil {
		return layer, err
	}

    propertyIDx, err := thisBitstream.readUint16BE()
    if err != nil {
        return layer, err
    }

    if int(propertyIDx) == int(len(context.StaticSchema.Properties)) { // explicit Parent property system
		layer.Bool1, err = thisBitstream.readBool()
		if err != nil {
			return layer, err
		}

        var referent Referent
        referent, err = thisBitstream.readObject(isClient, false, context)
        instance := context.InstancesByReferent.TryGetInstance(referent)
		result := rbxfile.ValueReference{instance}
		layer.Value = result
		layer.PropertyName = "Parent"

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

    layer.Bool1, err = thisBitstream.readBool()
    if err != nil {
        return layer, err
    }

    layer.Value, err = schema.Decode(isClient, ROUND_UPDATE, packet, context)

    context.InstancesByReferent.OnAddInstance(referent, func(instance *rbxfile.Instance) {
        layer.Instance = instance
        instance.Properties[layer.PropertyName] = layer.Value
    })

    return layer, err
}

func (layer *Packet83_03) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	err := stream.writeObject(isClient, layer.Instance, false, context)
	if err != nil {
		return err
	}

	if layer.PropertyName == "Parent" {
		err = stream.writeUint16BE(uint16(len(context.StaticSchema.Properties)))
	} else {
		err = stream.writeUint16BE(uint16(context.StaticSchema.PropertiesByName[layer.Instance.ClassName + "." + layer.PropertyName]))
	}
	if err != nil {
		return err
	}

	err = stream.writeBool(layer.Bool1)
	if err != nil {
		return err
	}

	err = context.StaticSchema.Properties[context.StaticSchema.PropertiesByName[layer.PropertyName]].serialize(isClient, layer.Value, ROUND_UPDATE, context, stream)
	return err
}
