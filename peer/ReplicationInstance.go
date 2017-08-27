package peer
import "fmt"
import "errors"

type ReplicationInstance struct {
	Object1 Object
	ClassName string
	Bool1 bool
	Object2 Object
	Properties []*ReplicationProperty
}

func (this *ReplicationInstance) findName() string {
	for _, property := range this.Properties {
		if property.Name == "Name" {
			return property.Show()
		}
	}

	return ""
}

func DecodeReplicationInstance(isJoinData bool, packet *UDPPacket, context *CommunicationContext, instanceSchema []*InstanceSchemaItem) (*ReplicationInstance, error) {
	var err error
	thisBitstream := packet.Stream
	thisInstance := &ReplicationInstance{}
	thisInstance.Object1, err = thisBitstream.ReadObject(isJoinData, context)
	if err != nil {
		return thisInstance, err
	}

	if !context.UseStaticSchema {
		classIDx, err := thisBitstream.Bits(9)
		if err != nil {
			return thisInstance, err
		}
		realIDx := (classIDx & 1 << 8) | classIDx >> 1
		if int(realIDx) > int(len(instanceSchema)) {
			return thisInstance, errors.New(fmt.Sprintf("idx %d is higher than %d", realIDx, len(context.InstanceSchema)))
		}
		thisInstance.ClassName = instanceSchema[realIDx].Name

		thisPropertySchema := instanceSchema[realIDx].PropertySchema

		thisInstance.Bool1, err = thisBitstream.ReadBool()
		if err != nil {
			return thisInstance, err
		}

		if isJoinData {
			for _, schema := range thisPropertySchema {
				property, err := schema.Decode(ROUND_JOINDATA, packet, context)
				if err != nil {
					return thisInstance, err
				}
				if property != nil {
					thisInstance.Properties = append(thisInstance.Properties, property)
				}
			}
		} else {
			for _, schema := range thisPropertySchema {
				property, err := schema.Decode(ROUND_STRINGS, packet, context)
				if err != nil {
					return thisInstance, err
				}
				if property != nil {
					thisInstance.Properties = append(thisInstance.Properties, property)
				}
			}
			for _, schema := range thisPropertySchema {
				property, err := schema.Decode(ROUND_OTHER, packet, context)
				if err != nil {
					return thisInstance, err
				}
				if property != nil {
					thisInstance.Properties = append(thisInstance.Properties, property)
				}
			}
		}

		thisInstance.Object2, err = thisBitstream.ReadObject(isJoinData, context)
		if err != nil {
			return thisInstance, err
		}

		return thisInstance, nil
	} else {
		serialized := formatBindable(thisInstance.Object1)
		_, wasRebind := context.Rebindables[serialized]
		context.Rebindables[serialized] = struct{}{}
		schemaIDx, err := thisBitstream.ReadUint16BE()
		if int(schemaIDx) > len(context.StaticSchema.Instances) {
			return thisInstance, errors.New(fmt.Sprintf("class idx %d is higher than %d", schemaIDx, len(context.StaticSchema.Instances)))
		}
		schema := context.StaticSchema.Instances[schemaIDx]

		thisInstance.Bool1, err = thisBitstream.ReadBool()
		if err != nil {
			return thisInstance, err
		}

		thisInstance.ClassName = schema.Name
		thisInstance.Properties = make([]*ReplicationProperty, len(schema.Properties))

		if isJoinData {
			for i := 0; i < len(thisInstance.Properties); i++ {
				thisInstance.Properties[i], err = schema.Properties[i].Decode(ROUND_JOINDATA, packet, context)
				if err != nil {
					return thisInstance, err
				}
			}
		} else {
			for i := 0; i < len(thisInstance.Properties); i++ {
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
					thisInstance.Properties[i], err = schema.Properties[i].Decode(ROUND_STRINGS, packet, context)
				}
			}
			for i := 0; i < len(thisInstance.Properties); i++ {
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
					thisInstance.Properties[i], err = schema.Properties[i].Decode(ROUND_STRINGS, packet, context)
				}
			}
		}
		thisInstance.Object2, err = thisBitstream.ReadObject(isJoinData, context)
		return thisInstance, err
	}
}
