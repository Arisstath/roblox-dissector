package peer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/DataDog/zstd"
	bitstream "github.com/gskartwii/go-bitstream"
)

const (
	PROP_TYPE_INVALID                uint8 = iota
	PROP_TYPE_STRING                       = iota
	PROP_TYPE_STRING_NO_CACHE              = iota
	PROP_TYPE_PROTECTEDSTRING_0            = iota
	PROP_TYPE_PROTECTEDSTRING_1            = iota
	PROP_TYPE_PROTECTEDSTRING_2            = iota
	PROP_TYPE_PROTECTEDSTRING_3            = iota
	PROP_TYPE_ENUM                         = iota
	PROP_TYPE_BINARYSTRING                 = iota
	PROP_TYPE_PBOOL                        = iota
	PROP_TYPE_PSINT                        = iota
	PROP_TYPE_PFLOAT                       = iota
	PROP_TYPE_PDOUBLE                      = iota
	PROP_TYPE_UDIM                         = iota
	PROP_TYPE_UDIM2                        = iota
	PROP_TYPE_RAY                          = iota
	PROP_TYPE_FACES                        = iota
	PROP_TYPE_AXES                         = iota
	PROP_TYPE_BRICKCOLOR                   = iota
	PROP_TYPE_COLOR3                       = iota
	PROP_TYPE_COLOR3UINT8                  = iota
	PROP_TYPE_VECTOR2                      = iota
	PROP_TYPE_VECTOR3_SIMPLE               = iota
	PROP_TYPE_VECTOR3_COMPLICATED          = iota
	PROP_TYPE_VECTOR2UINT16                = iota
	PROP_TYPE_VECTOR3UINT16                = iota
	PROP_TYPE_CFRAME_SIMPLE                = iota
	PROP_TYPE_CFRAME_COMPLICATED           = iota
	PROP_TYPE_INSTANCE                     = iota
	PROP_TYPE_TUPLE                        = iota
	PROP_TYPE_ARRAY                        = iota
	PROP_TYPE_DICTIONARY                   = iota
	PROP_TYPE_MAP                          = iota
	PROP_TYPE_CONTENT                      = iota
	PROP_TYPE_SYSTEMADDRESS                = iota
	PROP_TYPE_NUMBERSEQUENCE               = iota
	PROP_TYPE_NUMBERSEQUENCEKEYPOINT       = iota
	PROP_TYPE_NUMBERRANGE                  = iota
	PROP_TYPE_COLORSEQUENCE                = iota
	PROP_TYPE_COLORSEQUENCEKEYPOINT        = iota
	PROP_TYPE_RECT2D                       = iota
	PROP_TYPE_PHYSICALPROPERTIES           = iota
	PROP_TYPE_REGION3                      = iota
	PROP_TYPE_REGION3INT16                 = iota
	PROP_TYPE_INT64                        = iota
)

var TypeNames = map[uint8]string{
	PROP_TYPE_INVALID:                "???",
	PROP_TYPE_STRING:                 "string",
	PROP_TYPE_STRING_NO_CACHE:        "stringnc",
	PROP_TYPE_PROTECTEDSTRING_0:      "ProtectedString0",
	PROP_TYPE_PROTECTEDSTRING_1:      "ProtectedString1",
	PROP_TYPE_PROTECTEDSTRING_2:      "ProtectedString2",
	PROP_TYPE_PROTECTEDSTRING_3:      "ProtectedString3",
	PROP_TYPE_ENUM:                   "Enum",
	PROP_TYPE_BINARYSTRING:           "BinaryString",
	PROP_TYPE_PBOOL:                  "bool",
	PROP_TYPE_PSINT:                  "sint",
	PROP_TYPE_PFLOAT:                 "float",
	PROP_TYPE_PDOUBLE:                "double",
	PROP_TYPE_UDIM:                   "UDim",
	PROP_TYPE_UDIM2:                  "UDim2",
	PROP_TYPE_RAY:                    "Ray",
	PROP_TYPE_FACES:                  "Faces",
	PROP_TYPE_AXES:                   "Axes",
	PROP_TYPE_BRICKCOLOR:             "BrickColor",
	PROP_TYPE_COLOR3:                 "Color3",
	PROP_TYPE_COLOR3UINT8:            "Color3uint8",
	PROP_TYPE_VECTOR2:                "Vector2",
	PROP_TYPE_VECTOR3_SIMPLE:         "Vector3simp",
	PROP_TYPE_VECTOR3_COMPLICATED:    "Vector3comp",
	PROP_TYPE_VECTOR2UINT16:          "Vector2uint16",
	PROP_TYPE_VECTOR3UINT16:          "Vector3uint16",
	PROP_TYPE_CFRAME_SIMPLE:          "CFramesimp",
	PROP_TYPE_CFRAME_COMPLICATED:     "CFramecomp",
	PROP_TYPE_INSTANCE:               "Instance",
	PROP_TYPE_TUPLE:                  "Tuple",
	PROP_TYPE_ARRAY:                  "Array",
	PROP_TYPE_DICTIONARY:             "Dictionary",
	PROP_TYPE_MAP:                    "Map",
	PROP_TYPE_CONTENT:                "Content",
	PROP_TYPE_SYSTEMADDRESS:          "SystemAddress",
	PROP_TYPE_NUMBERSEQUENCE:         "NumberSequence",
	PROP_TYPE_NUMBERSEQUENCEKEYPOINT: "NumberSequenceKeypoint",
	PROP_TYPE_NUMBERRANGE:            "NumberRange",
	PROP_TYPE_COLORSEQUENCE:          "ColorSequence",
	PROP_TYPE_COLORSEQUENCEKEYPOINT:  "ColorSequenceKeypoint",
	PROP_TYPE_RECT2D:                 "Rect2D",
	PROP_TYPE_PHYSICALPROPERTIES:     "PhysicalProperties",
	PROP_TYPE_INT64:                  "sint64",
}

type StaticArgumentSchema struct {
	Type       uint8
	TypeString string
	EnumID     uint16
}

type StaticEnumSchema struct {
	Name      string
	BitSize   uint8
	NetworkID uint16
}

type StaticEventSchema struct {
	Name           string
	Arguments      []*StaticArgumentSchema
	InstanceSchema *StaticInstanceSchema
	NetworkID      uint16
}

type StaticPropertySchema struct {
	Name           string
	Type           uint8
	TypeString     string
	EnumID         uint16
	InstanceSchema *StaticInstanceSchema
	NetworkID      uint16
}

type StaticInstanceSchema struct {
	Name       string
	Unknown    uint16
	Properties []*StaticPropertySchema
	Events     []*StaticEventSchema
	NetworkID  uint16
}

func (schema *StaticInstanceSchema) LocalPropertyIndex(name string) int {
	for i := 0; i < len(schema.Properties); i++ {
		if schema.Properties[i].Name == name {
			return i
		}
	}
	return -1
}
func (schema *StaticInstanceSchema) SchemaForProp(name string) *StaticPropertySchema {
	idx := schema.LocalPropertyIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Properties[idx]
}
func (schema *StaticInstanceSchema) LocalEventIndex(name string) int {
	for i := 0; i < len(schema.Events); i++ {
		if schema.Events[i].Name == name {
			return i
		}
	}
	return -1
}
func (schema *StaticInstanceSchema) SchemaForEvent(name string) *StaticEventSchema {
	idx := schema.LocalEventIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Events[idx]
}

func (schema *StaticSchema) SchemaForClass(instance string) *StaticInstanceSchema {
	for _, inst := range schema.Instances {
		if inst.Name == instance {
			return inst
		}
	}
	return nil
}
func (schema *StaticSchema) SchemaForEnum(enum string) *StaticEnumSchema {
	for _, enumVal := range schema.Enums {
		if enum == enumVal.Name {
			return enumVal
		}
	}
	return nil
}

type StaticSchema struct {
	Instances  []*StaticInstanceSchema
	Properties []*StaticPropertySchema
	Events     []*StaticEventSchema
	Enums      []*StaticEnumSchema
}

// ID_NEW_SCHEMA - server -> client
// Negotiates a network schema with the client
type Packet97Layer struct {
	Schema *StaticSchema
}

func NewStaticShema() *StaticSchema {
	return &StaticSchema{}
}
func NewPacket97Layer() *Packet97Layer {
	return &Packet97Layer{Schema: NewStaticShema()}
}

func (thisBitstream *extendedReader) DecodePacket97Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket97Layer()

	var err error
	stream, err := thisBitstream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	enumArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	if enumArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum enum array len")
	}
	layer.Schema.Enums = make([]StaticEnumSchema, enumArrayLen)
	for i := 0; i < int(enumArrayLen); i++ {
		stringLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i] = new(StaticEnumSchema)
		layer.Schema.Enums[i].Name, err = stream.readASCII(int(stringLen))
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].BitSize, err = stream.readUint8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].NetworkID = i
	}

	classArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	propertyArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	eventArrayLen, err := stream.readUintUTF8()
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
	layer.Schema.Instances = make([]*StaticInstanceSchema, classArrayLen)
	layer.Schema.Properties = make([]*StaticPropertySchema, propertyArrayLen)
	layer.Schema.Events = make([]*StaticEventSchema, eventArrayLen)
	propertyGlobalIndex := 0
	classGlobalIndex := 0
	eventGlobalIndex := 0
	for i := 0; i < int(classArrayLen); i++ {
		classNameLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		thisInstance := new(StaticInstanceSchema)
		layer.Schema.Instances[i] = thisInstance
		thisInstance.Name, err = stream.readASCII(int(classNameLen))
		if err != nil {
			return layer, err
		}
		thisInstance.NetworkID = i

		propertyCount, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		if propertyCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum property count")
		}
		thisInstance.Properties = make([]*StaticPropertySchema, propertyCount)

		for j := 0; j < int(propertyCount); j++ {
			thisProperty := new(StaticPropertySchema)
			thisInstance.Properties[j] = thisProperty
			propNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisProperty.Name, err = stream.readASCII(int(propNameLen))
			if err != nil {
				return layer, err
			}
			thisProperty.NetworkID = uint16(propertyGlobalIndex)

			thisProperty.Type, err = stream.readUint8()
			if err != nil {
				return layer, err
			}
			thisProperty.TypeString = TypeNames[thisProperty.Type]

			if thisProperty.Type == 7 {
				thisProperty.EnumID, err = stream.readUint16BE()
				if err != nil {
					return layer, err
				}
			}

			if int(propertyGlobalIndex) >= int(propertyArrayLen) {
				return layer, errors.New("property global index too high")
			}

			thisProperty.InstanceSchema = thisInstance
			layer.Schema.Properties[propertyGlobalIndex] = thisProperty
			propertyGlobalIndex++
		}

		thisInstance.Unknown, err = stream.readUint16BE()
		if err != nil {
			return layer, err
		}
		if int(classGlobalIndex) >= int(classArrayLen) {
			return layer, errors.New("class global index too high")
		}

		eventCount, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		if eventCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum event count")
		}
		thisInstance.Events = make([]*StaticEventSchema, eventCount)

		for j := 0; j < int(eventCount); j++ {
			thisEvent := new(StaticEventSchema)
			thisInstance.Events[j] = thisEvent
			eventNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisEvent.Name, err = stream.readASCII(int(eventNameLen))
			if err != nil {
				return layer, err
			}
			thisEvent.NetworkID = uint16(eventGlobalIndex)

			countArguments, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			if countArguments > 0x10000 {
				return layer, errors.New("sanity check: exceeded maximum argument count")
			}
			thisEvent.Arguments = make([]*StaticArgumentSchema, countArguments)

			for k := 0; k < int(countArguments); k++ {
				thisArgument := new(StaticArgumentSchema)
				thisEvent.Arguments[k] = thisArgument
				thisArgument.Type, err = stream.readUint8()
				if err != nil {
					return layer, err
				}
				thisArgument.TypeString = TypeNames[thisArgument.Type]
				thisArgument.EnumID, err = stream.readUint16BE()
				if err != nil {
					return layer, err
				}
			}
			thisEvent.InstanceSchema = thisInstance
			layer.Schema.Events[eventGlobalIndex] = thisEvent
			eventGlobalIndex++
		}
		layer.Schema.Instances[classGlobalIndex] = thisInstance
		classGlobalIndex++
	}

	return layer, err
}

func (layer *Packet97Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	err = stream.WriteByte(0x97)
	if err != nil {
		return err
	}
	// TODO: General NewZStdBuf() method?
	uncompressedBuf := bytes.NewBuffer([]byte{})
	zstdBuf := bytes.NewBuffer([]byte{})
	middleStream := zstd.NewWriter(zstdBuf)
	zstdStream := &extendedWriter{bitstream.NewWriter(uncompressedBuf)}

	schema := layer.Schema
	err = zstdStream.writeUintUTF8(uint32(len(schema.Enums)))
	if err != nil {
		middleStream.Close()
		return err
	}
	for _, enum := range schema.Enums {
		err = zstdStream.writeUintUTF8(uint32(len(enum.Name)))
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeASCII(enum.Name)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.WriteByte(enum.BitSize)
		if err != nil {
			middleStream.Close()
			return err
		}
	}

	err = zstdStream.writeUintUTF8(uint32(len(schema.Instances)))
	if err != nil {
		middleStream.Close()
		return err
	}
	err = zstdStream.writeUintUTF8(uint32(len(schema.Properties)))
	if err != nil {
		middleStream.Close()
		return err
	}
	err = zstdStream.writeUintUTF8(uint32(len(schema.Events)))
	if err != nil {
		middleStream.Close()
		return err
	}
	for _, instance := range schema.Instances {
		err = zstdStream.writeUintUTF8(uint32(len(instance.Name)))
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeASCII(instance.Name)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeUintUTF8(uint32(len(instance.Properties)))
		if err != nil {
			middleStream.Close()
			return err
		}

		for _, property := range instance.Properties {
			err = zstdStream.writeUintUTF8(uint32(len(property.Name)))
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.writeASCII(property.Name)
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.WriteByte(property.Type)
			if err != nil {
				middleStream.Close()
				return err
			}
			if property.Type == 7 {
				err = zstdStream.writeUint16BE(property.EnumID)
				if err != nil {
					middleStream.Close()
					return err
				}
			}
		}

		err = zstdStream.writeUint16BE(instance.Unknown)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeUintUTF8(uint32(len(instance.Events)))
		if err != nil {
			middleStream.Close()
			return err
		}
		for _, event := range instance.Events {
			err = zstdStream.writeUintUTF8(uint32(len(event.Name)))
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.writeASCII(event.Name)
			if err != nil {
				middleStream.Close()
				return err
			}

			err = zstdStream.writeUintUTF8(uint32(len(event.Arguments)))
			if err != nil {
				middleStream.Close()
				return err
			}
			for _, argument := range event.Arguments {
				err = zstdStream.WriteByte(argument.Type)
				if err != nil {
					middleStream.Close()
					return err
				}
				err = zstdStream.writeUint16BE(argument.EnumID)
				if err != nil {
					middleStream.Close()
					return err
				}
			}
		}
	}

	err = zstdStream.Flush(bitstream.Zero)
	if err != nil {
		middleStream.Close()
		return err
	}
	_, err = middleStream.Write(uncompressedBuf.Bytes())
	if err != nil {
		middleStream.Close()
		return err
	}
	err = middleStream.Close()
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(zstdBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(uncompressedBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.allBytes(zstdBuf.Bytes())
	return err
}

func (layer *Packet97Layer) String() string {
	return fmt.Sprintf("ID_NEW_SCHEMA: %d enums, %d instances", len(layer.Schema.Enums), len(layer.Schema.Instances))
}
