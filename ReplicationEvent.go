package main
import "github.com/google/gopacket"
import "github.com/davecgh/go-spew/spew"
//import "errors"

type ReplicationEvent struct {
	Schema *EventSchemaItem
	Arguments []*PropertyValue
	IsDefault bool
}

func (schema *EventSchemaItem) Decode(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (*ReplicationEvent, error) {
	//var err error

	event := &ReplicationEvent{Schema: schema}
//	switch schema.Type {
//	case "bool":
//		event.Value, err = thisBitstream.ReadPBool()
//	case "string":
//		event.Value, err = thisBitstream.ReadPString(false, context)
//		break
//	case "ProtectedString":
//		event.Value, err = thisBitstream.ReadProtectedString(false, context)
//		break
//	case "BinaryString":
//		event.Value, err = thisBitstream.ReadBinaryString()
//		break
//	case "int":
//		event.Value, err = thisBitstream.ReadPInt()
//		break
//	case "float":
//		event.Value, err = thisBitstream.ReadPFloat()
//		break
//	case "double":
//		event.Value, err = thisBitstream.ReadPDouble()
//		break
//	case "Axes":
//		event.Value, err = thisBitstream.ReadAxes()
//		break
//	case "Faces":
//		event.Value, err = thisBitstream.ReadFaces()
//		break
//	case "BrickColor":
//		event.Value, err = thisBitstream.ReadBrickColor()
//		break
//	case "Object":
//		event.Value, err = thisBitstream.ReadObject(false, context)
//		break
//	case "UDim":
//		event.Value, err = thisBitstream.ReadUDim()
//		break
//	case "UDim2":
//		event.Value, err = thisBitstream.ReadUDim2()
//		break
//	case "Vector2":
//		event.Value, err = thisBitstream.ReadVector2()
//		break
//	case "Vector3":
//		event.Value, err = thisBitstream.ReadVector3()
//		break
//	case "Vector2uint16":
//		event.Value, err = thisBitstream.ReadVector2uint16()
//		break
//	case "Vector3uint16":
//		event.Value, err = thisBitstream.ReadVector3uint16()
//		break
//	case "Ray":
//		event.Value, err = thisBitstream.ReadRay()
//		break
//	case "Color3":
//		event.Value, err = thisBitstream.ReadColor3()
//		break
//	case "Color3uint8":
//		event.Value, err = thisBitstream.ReadColor3uint8()
//		break
//	case "CoordinateFrame":
//		event.Value, err = thisBitstream.ReadCFrame()
//		break
//	case "Content":
//		event.Value, err = thisBitstream.ReadContent()
//		break
//	default:
//		return event, errors.New("event parser encountered unknown type")
//	}
	println(DebugInfo2(context, packet, false), "Read", schema.Name, spew.Sdump(event.Arguments))
	return event, nil
}
