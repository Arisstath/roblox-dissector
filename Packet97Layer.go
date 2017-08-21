package main
import "github.com/google/gopacket"
import "errors"

const (
	PROP_TYPE_INVALID uint8 = iota
	PROP_TYPE_STRING = iota
	PROP_TYPE_STRING_NO_CACHE = iota
	PROP_TYPE_PROTECTEDSTRING_0 = iota
	PROP_TYPE_PROTECTEDSTRING_1 = iota
	PROP_TYPE_PROTECTEDSTRING_2 = iota
	PROP_TYPE_PROTECTEDSTRING_3 = iota
	PROP_TYPE_ENUM = iota
	PROP_TYPE_BINARYSTRING = iota
	PROP_TYPE_PBOOL = iota
	PROP_TYPE_PSINT = iota
	PROP_TYPE_PFLOAT = iota
	PROP_TYPE_PDOUBLE = iota
	PROP_TYPE_UDIM = iota
	PROP_TYPE_UDIM2 = iota
	PROP_TYPE_RAY = iota
	PROP_TYPE_FACES = iota
	PROP_TYPE_AXES = iota
	PROP_TYPE_BRICKCOLOR = iota
	PROP_TYPE_COLOR3 = iota
	PROP_TYPE_COLOR3UINT8 = iota
	PROP_TYPE_VECTOR2 = iota
	PROP_TYPE_VECTOR3_SIMPLE = iota
	PROP_TYPE_VECTOR3_COMPLICATED = iota
	PROP_TYPE_VECTOR2UINT16 = iota
	PROP_TYPE_VECTOR3UINT16 = iota
	PROP_TYPE_CFRAME_SIMPLE = iota
	PROP_TYPE_CFRAME_COMPLICATED = iota
	PROP_TYPE_INSTANCE = iota
	PROP_TYPE_TUPLE = iota
	PROP_TYPE_ARRAY = iota
	PROP_TYPE_DICTIONARY = iota
	PROP_TYPE_MAP = iota
	PROP_TYPE_CONTENT = iota
	PROP_TYPE_SYSTEMADDRESS = iota
	PROP_TYPE_NUMBERSEQUENCE = iota
	PROP_TYPE_NUMBERSEQUENCEKEYPOINT = iota
	PROP_TYPE_NUMBERRANGE = iota
	PROP_TYPE_COLORSEQUENCE = iota
	PROP_TYPE_COLORSEQUENCEKEYPOINT = iota
	PROP_TYPE_RECT2D = iota
	PROP_TYPE_PHYSICALPROPERTIES = iota
)

var TypeNames = map[uint8]string{
	PROP_TYPE_INVALID: "???",
	PROP_TYPE_STRING: "string",
	PROP_TYPE_STRING_NO_CACHE: "string",
	PROP_TYPE_PROTECTEDSTRING_0: "ProtectedString",
	PROP_TYPE_PROTECTEDSTRING_1: "ProtectedString",
	PROP_TYPE_PROTECTEDSTRING_2: "ProtectedString",
	PROP_TYPE_PROTECTEDSTRING_3: "ProtectedString",
	PROP_TYPE_ENUM: "Enum",
	PROP_TYPE_BINARYSTRING: "BinaryString",
	PROP_TYPE_PBOOL: "bool",
	PROP_TYPE_PSINT: "sint",
	PROP_TYPE_PFLOAT: "float",
	PROP_TYPE_PDOUBLE: "double",
	PROP_TYPE_UDIM: "UDim",
	PROP_TYPE_UDIM2: "UDim2",
	PROP_TYPE_RAY: "Ray",
	PROP_TYPE_FACES: "Faces",
	PROP_TYPE_AXES: "Axes",
	PROP_TYPE_BRICKCOLOR: "BrickColor",
	PROP_TYPE_COLOR3: "Color3",
	PROP_TYPE_COLOR3UINT8: "Color3uint8",
	PROP_TYPE_VECTOR2: "Vector2",
	PROP_TYPE_VECTOR3_SIMPLE: "Vector3",
	PROP_TYPE_VECTOR3_COMPLICATED: "Vector3",
	PROP_TYPE_VECTOR2UINT16: "Vector2uint16",
	PROP_TYPE_VECTOR3UINT16: "Vector3uint16",
	PROP_TYPE_CFRAME_SIMPLE: "CFrame",
	PROP_TYPE_CFRAME_COMPLICATED: "CFrame",
	PROP_TYPE_INSTANCE: "Instance",
	PROP_TYPE_TUPLE: "Tuple",
	PROP_TYPE_ARRAY: "Array",
	PROP_TYPE_DICTIONARY: "Dictionary",
	PROP_TYPE_MAP: "Map",
	PROP_TYPE_CONTENT: "Content",
	PROP_TYPE_SYSTEMADDRESS: "SystemAddress",
	PROP_TYPE_NUMBERSEQUENCE: "NumberSequence",
	PROP_TYPE_NUMBERSEQUENCEKEYPOINT: "NumberSequenceKeypoint",
	PROP_TYPE_NUMBERRANGE: "NumberRange",
	PROP_TYPE_COLORSEQUENCE: "ColorSequence",
	PROP_TYPE_COLORSEQUENCEKEYPOINT: "ColorSequenceKeypoint",
	PROP_TYPE_RECT2D: "Rect2D",
	PROP_TYPE_PHYSICALPROPERTIES: "PhysicalProperties",
}

type StaticArgumentSchema struct {
	Type uint8
	TypeString string
}

type StaticEnumSchema struct {
	Name string
	BitSize uint8
}

type StaticEventSchema struct {
	Name string
	Arguments []StaticArgumentSchema
	InstanceSchema *StaticInstanceSchema
}

type StaticPropertySchema struct {
	Name string
	Type uint8
	TypeString string
	InstanceSchema *StaticInstanceSchema
}

type StaticInstanceSchema struct {
	Name string
	Properties []StaticPropertySchema
	Events []StaticEventSchema
}

type StaticSchema struct {
	Instances []StaticInstanceSchema
	Properties []StaticPropertySchema
	Events []StaticEventSchema
	Enums []StaticEnumSchema
}


type Packet97Layer struct {
	Schema StaticSchema
}

func NewPacket97Layer() Packet97Layer {
	return Packet97Layer{}
}

func DecodePacket97Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket97Layer()

	context.MSchema.Lock()
	defer context.MSchema.Unlock()
	var err error
	stream, err := thisBitstream.RegionToGZipStream()
	if err != nil {
		return layer, err
	}

	enumArrayLen, err := stream.ReadUintUTF8()
	if err != nil {
		return layer, err
	}
	if enumArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum enum array len")
	}
	layer.Schema.Enums = make([]StaticEnumSchema, enumArrayLen)
	for i := 0; i < int(enumArrayLen); i++ {
		stringLen, err := stream.ReadUintUTF8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].Name, err = stream.ReadASCII(int(stringLen))
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].BitSize, err = stream.ReadUint8()
		if err != nil {
			return layer, err
		}
	}
	
	classArrayLen, err := stream.ReadUintUTF8()
	if err != nil {
		return layer, err
	}
	propertyArrayLen, err := stream.ReadUintUTF8()
	if err != nil {
		return layer, err
	}
	eventArrayLen, err := stream.ReadUintUTF8()
	if err != nil {
		return layer, err
	}
	if classArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum class array len")
	}
	if propertyArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum property array len")
	}
	if eventArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum event array len")
	}
	layer.Schema.Instances = make([]StaticInstanceSchema, classArrayLen)
	layer.Schema.Properties = make([]StaticPropertySchema, propertyArrayLen)
	layer.Schema.Events = make([]StaticEventSchema, eventArrayLen)
    propertyGlobalIndex := 0
    classGlobalIndex := 0
    eventGlobalIndex := 0
	for i := 0; i < int(classArrayLen); i++ {
		classNameLen, err := stream.ReadUintUTF8()
		if err != nil {
			return layer, err
		}
		thisInstance := layer.Schema.Instances[i]
		thisInstance.Name, err = stream.ReadASCII(int(classNameLen))
		if err != nil {
			return layer, err
		}
		propertyCount, err := stream.ReadUintUTF8()
		if err != nil {
			return layer, err
		}
		if propertyCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum property count")
		}
		thisInstance.Properties = make([]StaticPropertySchema, propertyCount)

		for j := 0; j < int(propertyCount); j++ {
			thisProperty := thisInstance.Properties[j]
			propNameLen, err := stream.ReadUintUTF8()
			if err != nil {
				return layer, err
			}
			thisProperty.Name, err = stream.ReadASCII(int(propNameLen))
			if err != nil {
				return layer, err
			}

			thisProperty.Type, err = stream.ReadUint8()
			if err != nil {
				return layer, err
			}
			thisProperty.TypeString = TypeNames[thisProperty.Type]

            if thisProperty.Type == 7 {
                _, err = stream.ReadUint16BE()
                if err != nil {
                    return layer, err
                }
            }

			if int(propertyGlobalIndex) >= int(propertyArrayLen) {
				return layer, errors.New("property global index too high")
			}

			thisProperty.InstanceSchema = &thisInstance
			layer.Schema.Properties[propertyGlobalIndex] = thisProperty
            thisInstance.Properties[j] = thisProperty
            propertyGlobalIndex++
		}

		_, err = stream.ReadUint16BE()
		if err != nil {
			return layer, err
		}
		if int(classGlobalIndex) >= int(classArrayLen) {
			return layer, errors.New("class global index too high")
		}

		eventCount, err := stream.ReadUintUTF8()
		if err != nil {
			return layer, err
		}
		if eventCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum event count")
		}
		thisInstance.Events = make([]StaticEventSchema, eventCount)

		for j := 0; j < int(eventCount); j++ {
			thisEvent := thisInstance.Events[j]
			eventNameLen, err := stream.ReadUintUTF8()
			if err != nil {
				return layer, err
			}
			thisEvent.Name, err = stream.ReadASCII(int(eventNameLen))
			if err != nil {
				return layer, err
			}

			countArguments, err := stream.ReadUintUTF8()
			if err != nil {
				return layer, err
			}
			if countArguments > 0x10000 {
				return layer, errors.New("sanity check: exceeded maximum argument count")
			}
			thisEvent.Arguments = make([]StaticArgumentSchema, countArguments)

			for k := 0; k < int(countArguments); k++ {
				thisArgument := thisEvent.Arguments[k]
				thisArgument.Type, err = stream.ReadUint8()
				if err != nil {
					return layer, err
				}
				thisArgument.TypeString = TypeNames[thisArgument.Type]
				_, err = stream.ReadUint16BE()
				if err != nil {
					return layer, err
				}
                thisEvent.Arguments[k] = thisArgument
			}
            layer.Schema.Events[eventGlobalIndex] = thisEvent
            thisInstance.Events[j] = thisEvent
            eventGlobalIndex++
		}
        layer.Schema.Instances[classGlobalIndex] = thisInstance
        classGlobalIndex++
	}
	context.StaticSchema = &layer.Schema
	context.ESchemaParsed.Broadcast()

	return layer, err
}
