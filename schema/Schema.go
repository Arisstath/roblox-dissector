package schema

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
	Name    string
	BitSize uint8
}

type StaticEventSchema struct {
	Name           string
	Arguments      []StaticArgumentSchema
	InstanceSchema *StaticInstanceSchema
}

type StaticPropertySchema struct {
	Name           string
	Type           uint8
	TypeString     string
	EnumID         uint16
	InstanceSchema *StaticInstanceSchema
}

type StaticInstanceSchema struct {
	Name       string
	Unknown    uint16
	Properties []StaticPropertySchema
	Events     []StaticEventSchema
}

func (schema *StaticInstanceSchema) FindPropertyIndex(name string) int {
	for i := 0; i < len(schema.Properties); i++ {
		if schema.Properties[i].Name == name {
			return i
		}
	}
	return -1
}
func (schema *StaticInstanceSchema) FindEventIndex(name string) int {
	for i := 0; i < len(schema.Events); i++ {
		if schema.Events[i].Name == name {
			return i
		}
	}
	return -1
}

type StaticSchema struct {
	Instances  []StaticInstanceSchema
	Properties []StaticPropertySchema
	Events     []StaticEventSchema
	Enums      []StaticEnumSchema
	// TODO: Improve this
	ClassesByName    map[string]int
	PropertiesByName map[string]int
	EventsByName     map[string]int
	EnumsByName      map[string]int
}

