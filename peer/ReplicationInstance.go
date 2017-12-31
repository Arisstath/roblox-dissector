package peer
import "fmt"
import "errors"
import "github.com/gskartwii/rbxfile"

func decodeReplicationInstance(isClient bool, isJoinData bool, packet *UDPPacket, context *CommunicationContext) (*rbxfile.Instance, error) {
	var err error
	thisBitstream := packet.stream

    referent, err := thisBitstream.readObject(isClient, isJoinData, context)
	if err != nil {
        return nil, errors.New("while parsing self: " + err.Error())
	}
    thisInstance := context.InstancesByReferent.TryGetInstance(referent)

    schemaIDx, err := thisBitstream.readUint16BE()
    if int(schemaIDx) > len(context.StaticSchema.Instances) {
        return thisInstance, errors.New(fmt.Sprintf("class idx %d is higher than %d", schemaIDx, len(context.StaticSchema.Instances)))
    }
    schema := context.StaticSchema.Instances[schemaIDx]
    thisInstance.ClassName = schema.Name
	if DEBUG && isJoinData && isClient {
		println("will parse", referent, schema.Name, isJoinData, len(schema.Properties))
	}

    _, err = thisBitstream.readBool()
    if err != nil {
        return thisInstance, err
    }
    thisInstance.Properties = make(map[string]rbxfile.Value, len(schema.Properties))

    if isJoinData {
        for i := 0; i < len(schema.Properties); i++ {
            propertyName := schema.Properties[i].Name
			value, err := schema.Properties[i].Decode(isClient, ROUND_JOINDATA, packet, context)
			if err != nil {
				return thisInstance, err
			}
			if value != nil {
				thisInstance.Properties[propertyName] = value
			}
        }
    } else {
        for i := 0; i < len(schema.Properties); i++ {
            isStringObject := false
            if  schema.Properties[i].Type == 0x21 ||
                schema.Properties[i].Type == 0x01 ||
                schema.Properties[i].Type == 0x1C ||
                schema.Properties[i].Type == 0x22 ||
                schema.Properties[i].Type == 0x06 ||
                schema.Properties[i].Type == 0x04 ||
                schema.Properties[i].Type == 0x05 ||
                schema.Properties[i].Type == 0x03 {
                    isStringObject = true
            }
            if isStringObject {
                propertyName := schema.Properties[i].Name
				value, err := schema.Properties[i].Decode(isClient, ROUND_STRINGS, packet, context)
				if err != nil {
					return thisInstance, err
				}
				if value != nil {
					thisInstance.Properties[propertyName] = value
				}
            }
        }
        for i := 0; i < len(schema.Properties); i++ {
            isStringObject := false
            if  schema.Properties[i].Type == 0x21 ||
                schema.Properties[i].Type == 0x01 ||
                schema.Properties[i].Type == 0x1C ||
                schema.Properties[i].Type == 0x22 ||
                schema.Properties[i].Type == 0x06 ||
                schema.Properties[i].Type == 0x04 ||
                schema.Properties[i].Type == 0x05 ||
                schema.Properties[i].Type == 0x03 {
                    isStringObject = true
            }
            if !isStringObject {
                propertyName := schema.Properties[i].Name
				value, err := schema.Properties[i].Decode(isClient, ROUND_OTHER, packet, context)
				if err != nil {
					return thisInstance, err
				}
				if value != nil {
					thisInstance.Properties[propertyName] = value
				}
            }
        }
    }
    referent, err = thisBitstream.readObject(isClient, isJoinData, context)
    if err != nil {
        return thisInstance, errors.New("while parsing parent: " + err.Error())
    }
	if DEBUG && isJoinData && isClient {
		if len(referent) > 0x50 {
			println("Parent: (invalid), ", len(referent), len(thisInstance.Get("Source").(rbxfile.ValueProtectedString)))
		} else {
			println("Parent: ", referent)
		}
	}

    context.InstancesByReferent.AddInstance(Referent(thisInstance.Reference), thisInstance)
    parent := context.InstancesByReferent.TryGetInstance(referent)
    err = parent.AddChild(thisInstance)

    return thisInstance, err
}

func serializeReplicationInstance(isClient bool, instance *rbxfile.Instance, isJoinData bool, context *CommunicationContext, stream *extendedWriter) error {
    var err error
    err = stream.writeObject(isClient, instance, isJoinData, context)
    if err != nil {
        return err
    }

    schemaIdx := uint16(context.StaticSchema.ClassesByName[instance.ClassName])
    err = stream.writeUint16BE(schemaIdx)
    if err != nil {
        return err
    }
    err = stream.writeBool(false) // ???
    if err != nil {
        return err
    }

    schema := context.StaticSchema.Instances[schemaIdx]
    if isJoinData {
        for i := 0; i < len(schema.Properties); i++ {
            value := instance.Get(schema.Properties[i].Name)
            if value == nil {
				println(schema.Properties[i].Name, "was nil")
                value = rbxfile.DefaultValue
            }
			println("serializing", schema.Properties[i].Name)
            err = schema.Properties[i].serialize(isClient, value, ROUND_JOINDATA, context, stream)
			println("ser done")
            if err != nil {
                return err
            }
        }
    } else {
        for i := 0; i < len(schema.Properties); i++ {
            isStringObject := false
            if  schema.Properties[i].Type == 0x21 ||
                schema.Properties[i].Type == 0x01 ||
                schema.Properties[i].Type == 0x1C ||
                schema.Properties[i].Type == 0x22 ||
                schema.Properties[i].Type == 0x06 ||
                schema.Properties[i].Type == 0x04 ||
                schema.Properties[i].Type == 0x05 ||
                schema.Properties[i].Type == 0x03 {
                    isStringObject = true
            }
            if isStringObject {
				value := instance.Get(schema.Properties[i].Name)
				err = schema.Properties[i].serialize(isClient, value, ROUND_STRINGS, context, stream)
				if err != nil {
					return err
				}
            }
        }
        for i := 0; i < len(schema.Properties); i++ {
            isStringObject := false
            if  schema.Properties[i].Type == 0x21 ||
                schema.Properties[i].Type == 0x01 ||
                schema.Properties[i].Type == 0x1C ||
                schema.Properties[i].Type == 0x22 ||
                schema.Properties[i].Type == 0x06 ||
                schema.Properties[i].Type == 0x04 ||
                schema.Properties[i].Type == 0x05 ||
                schema.Properties[i].Type == 0x03 {
                    isStringObject = true
            }
            if !isStringObject {
				value := instance.Get(schema.Properties[i].Name)
				err := schema.Properties[i].serialize(isClient, value, ROUND_OTHER, context, stream)
				if err != nil {
					return err
				}
            }
        }
	}

    err = stream.writeObject(isClient, instance.Parent(), isJoinData, context)
    if err != nil {
        return err
    }
    return nil
}
