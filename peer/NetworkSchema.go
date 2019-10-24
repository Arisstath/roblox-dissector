package peer

const (
	// PropertyTypeNil is the type for nil values
	PropertyTypeNil uint8 = iota
	// PropertyTypeString is the type for string values
	PropertyTypeString = iota
	// PropertyTypeStringNoCache is the type for string values that
	// should never be cached
	PropertyTypeStringNoCache = iota
	// PropertyTypeProtectedString0 is a type the purpose of which is unknown
	PropertyTypeProtectedString0 = iota
	// PropertyTypeProtectedString1 is a type the purpose of which is unknown
	PropertyTypeProtectedString1 = iota
	// PropertyTypeProtectedString2 is the type for encrypted
	// ProtectedStrings
	PropertyTypeProtectedString2 = iota
	// PropertyTypeProtectedString3 is the type for unencrypted
	// ProtectedStrings
	PropertyTypeProtectedString3 = iota
	// PropertyTypeEnum is the type for enum values
	PropertyTypeEnum = iota
	// PropertyTypeBinaryString is the type for BinaryString values
	PropertyTypeBinaryString = iota
	// PropertyTypeBool is the type for boolean values
	PropertyTypeBool = iota
	// PropertyTypeInt is the type for 32-bit signed integer values
	PropertyTypeInt = iota
	// PropertyTypeFloat is the type for single-precision floating point values
	PropertyTypeFloat = iota
	// PropertyTypeDouble is the type for double-precision floating point values
	PropertyTypeDouble = iota
	// PropertyTypeUDim is the type for UDim values
	PropertyTypeUDim = iota
	// PropertyTypeUDim2 is the type for UDim2 values
	PropertyTypeUDim2 = iota
	// PropertyTypeRay is the type for Ray values
	PropertyTypeRay = iota
	// PropertyTypeFaces is the type for Faces values
	PropertyTypeFaces = iota
	// PropertyTypeAxes is the type for Axes values
	PropertyTypeAxes = iota
	// PropertyTypeBrickColor is the type for BrickColor values
	PropertyTypeBrickColor = iota
	// PropertyTypeColor3 is the type for floating-point Color3 values
	PropertyTypeColor3 = iota
	// PropertyTypeColor3uint8 is the type for uint8 Color3 values
	PropertyTypeColor3uint8 = iota
	// PropertyTypeVector2 is the type for Vector2 values
	PropertyTypeVector2 = iota
	// PropertyTypeSimpleVector3 is the type for most Vector3 values
	PropertyTypeSimpleVector3 = iota
	// PropertyTypeComplicatedVector3 is the type for Vector3 values that use
	// the "complicated" schema (usage and purpose currently unknown)
	PropertyTypeComplicatedVector3 = iota
	// PropertyTypeVector2int16 is the type for Vector2int16 values
	PropertyTypeVector2int16 = iota
	// PropertyTypeVector3int16 is the type for Vector2int16 values
	PropertyTypeVector3int16 = iota
	// PropertyTypeSimpleCFrame is the type for CFrame values that use
	// the "simple" schema (usage and purpose currently unknown)
	PropertyTypeSimpleCFrame = iota
	// PropertyTypeComplicatedCFrame is the type for most CFrame values
	PropertyTypeComplicatedCFrame = iota
	// PropertyTypeInstance is the type for Reference values
	PropertyTypeInstance = iota
	// PropertyTypeTuple is the type for Tuple values
	PropertyTypeTuple = iota
	// PropertyTypeArray is the type for Array values
	PropertyTypeArray = iota
	// PropertyTypeDictionary is the type for Dictionary values
	PropertyTypeDictionary = iota
	// PropertyTypeMap is the type for Map values
	PropertyTypeMap = iota
	// PropertyTypeContent is the type for Content values
	PropertyTypeContent = iota
	// PropertyTypeSystemAddress is the type for SystemAddress values
	PropertyTypeSystemAddress = iota
	// PropertyTypeNumberSequence is the type for NumberSequence values
	PropertyTypeNumberSequence = iota
	// PropertyTypeNumberSequenceKeypoint is the type for NumberSequenceKeypoint values
	PropertyTypeNumberSequenceKeypoint = iota
	// PropertyTypeNumberRange is the type for NumberRange values
	PropertyTypeNumberRange = iota
	// PropertyTypeColorSequence is the type for ColorSequence values
	PropertyTypeColorSequence = iota
	// PropertyTypeColorSequenceKeypoint is the type for ColorSequenceKeypoint values
	PropertyTypeColorSequenceKeypoint = iota
	// PropertyTypeRect2D is the type for Rect2D values
	PropertyTypeRect2D = iota
	// PropertyTypePhysicalProperties is the type for PhysicalProperties values
	PropertyTypePhysicalProperties = iota
	// PropertyTypeRegion3 is the type for Region3 values
	PropertyTypeRegion3 = iota
	// PropertyTypeRegion3int16 is the type for Region3int16 values
	PropertyTypeRegion3int16 = iota
	// PropertyTypeInt64 is the type for 64-bit signed integer values
	PropertyTypeInt64 = iota
	// PropertyTypePathWaypoint is the type for path waypoints
	PropertyTypePathWaypoint = iota
	// PropertyTypeSharedString is the type for shared strings
	PropertyTypeSharedString = iota
	// PropertyTypeLuauString is the type for Luau ProtectedStrings
	PropertyTypeLuauString = iota
)

// TypeNames is a list of names for value types
var TypeNames = map[uint8]string{
	PropertyTypeNil:                    "nil",
	PropertyTypeString:                 "string",
	PropertyTypeStringNoCache:          "string (no cache)",
	PropertyTypeProtectedString0:       "Server ProtectedString",
	PropertyTypeProtectedString1:       "ProtectedString 1",
	PropertyTypeProtectedString2:       "Encrypted ProtectedString",
	PropertyTypeProtectedString3:       "Studio ProtectedString",
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
	PropertyTypeSimpleVector3:          "Vector3 (simple)",
	PropertyTypeComplicatedVector3:     "Vector3 (complicated)",
	PropertyTypeVector2int16:           "Vector2uint16",
	PropertyTypeVector3int16:           "Vector3uint16",
	PropertyTypeSimpleCFrame:           "CFrame (simple)",
	PropertyTypeComplicatedCFrame:      "CFrame (complicated)",
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
	PropertyTypePathWaypoint:           "PathWaypoint",
	PropertyTypeSharedString:           "SharedString",
	PropertyTypeLuauString:             "Luau ProtectedString",
}

// NetworkArgumentSchema describes the schema of one event argument
type NetworkArgumentSchema struct {
	Type       uint8
	TypeString string
	EnumID     uint16
}

// NetworkEnumSchema describes the schema of one enum
type NetworkEnumSchema struct {
	Name      string
	BitSize   uint8
	NetworkID uint16
}

// NetworkEventSchema describes the schema of one event
type NetworkEventSchema struct {
	Name           string
	Arguments      []*NetworkArgumentSchema
	InstanceSchema *NetworkInstanceSchema
	NetworkID      uint16
}

// NetworkPropertySchema describes the schema of one property
type NetworkPropertySchema struct {
	Name           string
	Type           uint8
	TypeString     string
	EnumID         uint16
	InstanceSchema *NetworkInstanceSchema
	NetworkID      uint16
}

// NetworkInstanceSchema describes the schema of one class
type NetworkInstanceSchema struct {
	Name       string
	Unknown    uint16
	Properties []*NetworkPropertySchema
	Events     []*NetworkEventSchema
	NetworkID  uint16
}

// LocalPropertyIndex finds the index of a certain property within one class
func (schema *NetworkInstanceSchema) LocalPropertyIndex(name string) int {
	for i := 0; i < len(schema.Properties); i++ {
		if schema.Properties[i].Name == name {
			return i
		}
	}
	return -1
}

// SchemaForProp finds the schema for a certain property
func (schema *NetworkInstanceSchema) SchemaForProp(name string) *NetworkPropertySchema {
	idx := schema.LocalPropertyIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Properties[idx]
}

// LocalEventIndex finds the index of a certain event within one class
func (schema *NetworkInstanceSchema) LocalEventIndex(name string) int {
	for i := 0; i < len(schema.Events); i++ {
		if schema.Events[i].Name == name {
			return i
		}
	}
	return -1
}

// SchemaForEvent finds the schema for a certain event
func (schema *NetworkInstanceSchema) SchemaForEvent(name string) *NetworkEventSchema {
	idx := schema.LocalEventIndex(name)
	if idx == -1 {
		return nil
	}
	return schema.Events[idx]
}

// SchemaForClass finds the schema for a certain class
func (schema *NetworkSchema) SchemaForClass(instance string) *NetworkInstanceSchema {
	for _, inst := range schema.Instances {
		if inst.Name == instance {
			return inst
		}
	}
	return nil
}

// SchemaForEnum finds the schema for a certain enum
func (schema *NetworkSchema) SchemaForEnum(enum string) *NetworkEnumSchema {
	for _, enumVal := range schema.Enums {
		if enum == enumVal.Name {
			return enumVal
		}
	}
	return nil
}

// NetworkSchema represents the data serialization schema
// and class/enum API for a communication as specified by the server
type NetworkSchema struct {
	Instances  []*NetworkInstanceSchema
	Properties []*NetworkPropertySchema
	Events     []*NetworkEventSchema
	Enums      []*NetworkEnumSchema
}
