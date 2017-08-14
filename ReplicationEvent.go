package main
import "github.com/google/gopacket"
import "github.com/davecgh/go-spew/spew"
import "errors"

type ReplicationEvent struct {
	UnknownInt uint32
	Arguments []PropertyValue
}

func decodeEventArgument(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, argType string) (PropertyValue, error) {
	var argument PropertyValue
	var err error
	switch argType {
	case "bool":
		argument, err = thisBitstream.ReadPBool()
		break
	case "string":
		argument, err = thisBitstream.ReadPString(false, context)
		break
	case "BinaryString":
		argument, err = thisBitstream.ReadBinaryString()
		break
	case "int":
		argument, err = thisBitstream.ReadPInt()
		break
	case "float":
		argument, err = thisBitstream.ReadPFloat()
		break
	case "double":
		argument, err = thisBitstream.ReadPDouble()
		break
	case "Axes":
		argument, err = thisBitstream.ReadAxes()
		break
	case "Faces":
		argument, err = thisBitstream.ReadFaces()
		break
	case "BrickColor":
		argument, err = thisBitstream.ReadBrickColor()
		break
	case "Object":
		argument, err = thisBitstream.ReadObject(false, context)
		break
	case "UDim":
		argument, err = thisBitstream.ReadUDim()
		break
	case "UDim2":
		argument, err = thisBitstream.ReadUDim2()
		break
	case "Vector2":
		argument, err = thisBitstream.ReadVector2()
		break
	case "Vector3":
		argument, err = thisBitstream.ReadVector3Simple()
		break
	case "Vector2uint16":
		argument, err = thisBitstream.ReadVector2uint16()
		break
	case "Vector3uint16":
		argument, err = thisBitstream.ReadVector3uint16()
		break
	case "Ray":
		argument, err = thisBitstream.ReadRay()
		break
	case "Color3":
		argument, err = thisBitstream.ReadColor3()
		break
	case "Color3uint8":
		argument, err = thisBitstream.ReadColor3uint8()
		break
	case "CoordinateFrame":
		argument, err = thisBitstream.ReadCFrame()
		break
	case "Content":
		argument, err = thisBitstream.ReadContent(false, context)
		break
	case "Instance":
		argument, err = thisBitstream.ReadObject(false, context)
		break
	case "long":
		argument, err = thisBitstream.ReadPInt()
		break
	case "Region3":
		argument, err = thisBitstream.ReadRegion3()
		break
	case "Region3uint16":
		argument, err = thisBitstream.ReadRegion3uint16()
		break
	case "Tuple":
		argument, err = thisBitstream.ReadTuple(context, packet)
		break
	case "Array":
		argument, err = thisBitstream.ReadArray(context, packet)
		break
	case "Dictionary":
		argument, err = thisBitstream.ReadDictionary(context, packet)
		break
	case "Map":
		argument, err = thisBitstream.ReadMap(context, packet)
		break
	default:
		if schema, ok := context.EnumSchema[argType]; ok {
			argument, err = thisBitstream.ReadEnumValue(schema.BitSize)
		} else {
			return argument, errors.New("event parser encountered unknown type " + argType)
		}
	}
	return argument, err
}

func (schema *EventSchemaItem) Decode(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (*ReplicationEvent, error) {
	var err error

	event := &ReplicationEvent{}
	event.UnknownInt, err = thisBitstream.ReadUint32BE()
	event.Arguments = make([]PropertyValue, len(schema.ArgumentTypes))
	for i, argSchemaName := range schema.ArgumentTypes {
		event.Arguments[i], err = decodeEventArgument(thisBitstream, context, packet, argSchemaName)
		if err != nil {
			return event, err
		}
	}
	println(DebugInfo2(context, packet, false), "Read", schema.Name, spew.Sdump(event.Arguments))
	return event, nil
}
