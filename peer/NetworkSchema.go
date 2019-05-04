package peer

const (
	PropertyTypeNil                    uint8 = iota
	PropertyTypeString                       = iota
	PropertyTypeStringNoCache                = iota
	PropertyTypeProtectedString0             = iota
	PropertyTypeProtectedString1             = iota
	PropertyTypeProtectedString2             = iota
	PropertyTypeProtectedString3             = iota
	PropertyTypeEnum                         = iota
	PropertyTypeBinaryString                 = iota
	PropertyTypeBool                         = iota
	PropertyTypeInt                          = iota
	PropertyTypeFloat                        = iota
	PropertyTypeDouble                       = iota
	PropertyTypeUDim                         = iota
	PropertyTypeUDim2                        = iota
	PropertyTypeRay                          = iota
	PropertyTypeFaces                        = iota
	PropertyTypeAxes                         = iota
	PropertyTypeBrickColor                   = iota
	PropertyTypeColor3                       = iota
	PropertyTypeColor3uint8                  = iota
	PropertyTypeVector2                      = iota
	PropertyTypeSimpleVector3                = iota
	PropertyTypeComplicatedVector3           = iota
	PropertyTypeVector2int16                 = iota
	PropertyTypeVectorint16                  = iota
	PropertyTypeSimpleCFrame                 = iota
	PropertyTypeComplicatedCFrame            = iota
	PropertyTypeInstance                     = iota
	PropertyTypeTuple                        = iota
	PropertyTypeArray                        = iota
	PropertyTypeDictionary                   = iota
	PropertyTypeMap                          = iota
	PropertyTypeContent                      = iota
	PropertyTypeSystemAddress                = iota
	PropertyTypeNumberSequence               = iota
	PropertyTypeNumberSequenceKeypoint       = iota
	PropertyTypeNumberRange                  = iota
	PropertyTypeColorSequence                = iota
	PropertyTypeColorSequenceKeypoint        = iota
	PropertyTypeRect2D                       = iota
	PropertyTypePhysicalProperties           = iota
	PropertyTypeREGION3                      = iota
	PropertyTypeREGION3INT16                 = iota
	PropertyTypeInt64                        = iota
)

var TypeNames = map[uint8]string{
	PropertyTypeNil:                    "nil",
	PropertyTypeString:                 "string",
	PropertyTypeStringNoCache:          "stringnc",
	PropertyTypeProtectedString0:       "ProtectedString0",
	PropertyTypeProtectedString1:       "ProtectedString1",
	PropertyTypeProtectedString2:       "ProtectedString2",
	PropertyTypeProtectedString3:       "ProtectedString3",
	PropertyTypeEnum:                   "Enum",
	PropertyTypeBinaryString:           "BinaryString",
	PropertyTypeBool:                   "bool",
	PropertyTypeInt:                    "sint",
	PropertyTypeFloat:                  "float",
	PropertyTypeDouble:                 "double",
	PropertyTypeUDim:                   "UDim",
	PropertyTypeUDim2:                  "UDim2",
	PropertyTypeRay:                    "Ray",
	PropertyTypeFaces:                  "Faces",
	PropertyTypeAxes:                   "Axes",
	PropertyTypeBrickColor:             "BrickColor",
	PropertyTypeColor3:                 "Color3",
	PropertyTypeColor3uint8:            "Color3uint8",
	PropertyTypeVector2:                "Vector2",
	PropertyTypeSimpleVector3:          "Vector3simp",
	PropertyTypeComplicatedVector3:     "Vector3comp",
	PropertyTypeVector2int16:           "Vector2uint16",
	PropertyTypeVectorint16:            "Vector3uint16",
	PropertyTypeSimpleCFrame:           "CFramesimp",
	PropertyTypeComplicatedCFrame:      "CFramecomp",
	PropertyTypeInstance:               "Instance",
	PropertyTypeTuple:                  "Tuple",
	PropertyTypeArray:                  "Array",
	PropertyTypeDictionary:             "Dictionary",
	PropertyTypeMap:                    "Map",
	PropertyTypeContent:                "Content",
	PropertyTypeSystemAddress:          "SystemAddress",
	PropertyTypeNumberSequence:         "NumberSequence",
	PropertyTypeNumberSequenceKeypoint: "NumberSequenceKeypoint",
	PropertyTypeNumberRange:            "NumberRange",
	PropertyTypeColorSequence:          "ColorSequence",
	PropertyTypeColorSequenceKeypoint:  "ColorSequenceKeypoint",
	PropertyTypeRect2D:                 "Rect2D",
	PropertyTypePhysicalProperties:     "PhysicalProperties",
	PropertyTypeInt64:                  "sint64",
}

type NetworkArgumentSchema struct {
	Type       uint8
	TypeString string
	EnumID     uint16
}

type NetworkEnumSchema struct {
	Name      string
	BitSize   uint8
	NetworkID uint16
}

type NetworkEventSchema struct {
	Name           string
	Arguments      []*NetworkArgumentSchema
	InstanceSchema *NetworkInstanceSchema
	NetworkID      uint16
}

type NetworkPropertySchema struct {
	Name           string
	Type           uint8
	TypeString     string
	EnumID         uint16
	InstanceSchema *NetworkInstanceSchema
	NetworkID      uint16
}

type NetworkInstanceSchema struct {
	Name       string
	Unknown    uint16
	Properties []*NetworkPropertySchema
	Events     []*NetworkEventSchema
	NetworkID  uint16
}

func (schema *NetworkInstanceSchema) LocalPropertyIndex(name string) int {
	for i := 0; i < len(schema.Properties); i++ {
		if schema.Properties[i].Name == name {
			return i
		}
	}
	return -1
}
func (schema *NetworkInstanceSchema) SchemaForProp(name string) *NetworkPropertySchema {
	idx := schema.LocalPropertyIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Properties[idx]
}
func (schema *NetworkInstanceSchema) LocalEventIndex(name string) int {
	for i := 0; i < len(schema.Events); i++ {
		if schema.Events[i].Name == name {
			return i
		}
	}
	return -1
}
func (schema *NetworkInstanceSchema) SchemaForEvent(name string) *NetworkEventSchema {
	idx := schema.LocalEventIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Events[idx]
}

func (schema *NetworkSchema) SchemaForClass(instance string) *NetworkInstanceSchema {
	for _, inst := range schema.Instances {
		if inst.Name == instance {
			return inst
		}
	}
	return nil
}
func (schema *NetworkSchema) SchemaForEnum(enum string) *NetworkEnumSchema {
	for _, enumVal := range schema.Enums {
		if enum == enumVal.Name {
			return enumVal
		}
	}
	return nil
}

type NetworkSchema struct {
	Instances  []*NetworkInstanceSchema
	Properties []*NetworkPropertySchema
	Events     []*NetworkEventSchema
	Enums      []*NetworkEnumSchema
}
