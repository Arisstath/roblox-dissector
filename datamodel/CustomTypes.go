package datamodel

import (
	"bytes"
	"fmt"

	"github.com/robloxapi/rbxfile"
)

type Reference struct {
	IsNull bool
	Scope  string
	Id     uint32
	PeerId uint32
}

var NullReference = Reference{
	IsNull: true,
	Scope:  "null",
}

func (ref Reference) String() string {
	if ref.IsNull {
		return "null"
	}
	return fmt.Sprintf("%s_%d", ref.Scope, ref.Id)
}

func (ref Reference) Equal(other *Reference) bool {
	if ref.IsNull {
		return other.IsNull
	}
	return ref.Id == other.Id && ref.PeerId == other.PeerId
}

const (
	TypeNumberSequenceKeypoint rbxfile.Type = rbxfile.TypeSharedString + 1 + iota
	TypeColorSequenceKeypoint               = rbxfile.TypeSharedString + 1 + iota
	TypeNumberSequence                      = rbxfile.TypeSharedString + 1 + iota
	TypeColorSequence                       = rbxfile.TypeSharedString + 1 + iota
	TypeSystemAddress                       = rbxfile.TypeSharedString + 1 + iota
	TypeMap                                 = rbxfile.TypeSharedString + 1 + iota
	TypeDictionary                          = rbxfile.TypeSharedString + 1 + iota
	TypeArray                               = rbxfile.TypeSharedString + 1 + iota
	TypeTuple                               = rbxfile.TypeSharedString + 1 + iota
	TypeRegion3                             = rbxfile.TypeSharedString + 1 + iota
	TypeRegion3int16                        = rbxfile.TypeSharedString + 1 + iota
	TypeReference                           = rbxfile.TypeSharedString + 1 + iota
	TypeToken                               = rbxfile.TypeSharedString + 1 + iota
	// used for terrain
	TypeVector3int32          = rbxfile.TypeSharedString + 1 + iota
	TypePathWaypoint          = rbxfile.TypeSharedString + 1 + iota
	TypeDeferredString        = rbxfile.TypeSharedString + 1 + iota
	TypeSignedProtectedString = rbxfile.TypeSharedString + 1 + iota
)

var CustomTypeNames = map[rbxfile.Type]string{
	TypeNumberSequenceKeypoint: "NumberSequenceKeypoint",
	TypeColorSequenceKeypoint:  "ColorSequenceKeypoint",
	TypeNumberSequence:         "NumberSequence",
	TypeColorSequence:          "ColorSequence",
	TypeSystemAddress:          "SystemAddress",
	TypeMap:                    "Map",
	TypeDictionary:             "Dictionary",
	TypeArray:                  "Array",
	TypeTuple:                  "Tuple",
	TypeRegion3:                "Region3",
	TypeRegion3int16:           "Region3int16",
	TypeReference:              "Reference",
	TypeToken:                  "Token",
	TypeVector3int32:           "Vector3int32",
	TypePathWaypoint:           "PathWaypoint",
	TypeDeferredString:         "SharedString (deferred)",
}

type ValueColorSequenceKeypoint rbxfile.ValueColorSequenceKeypoint
type ValueNumberSequenceKeypoint rbxfile.ValueNumberSequenceKeypoint
type ValueColorSequence []ValueColorSequenceKeypoint
type ValueNumberSequence []ValueNumberSequenceKeypoint
type ValueTuple []rbxfile.Value
type ValueArray []rbxfile.Value
type ValueDictionary map[string]rbxfile.Value
type ValueMap map[string]rbxfile.Value
type ValueRegion3 struct {
	Start rbxfile.ValueVector3
	End   rbxfile.ValueVector3
}
type ValueRegion3int16 struct {
	Start rbxfile.ValueVector3int16
	End   rbxfile.ValueVector3int16
}
type ValueSystemAddress uint64
type ValueReference struct {
	Reference Reference
	Instance  *Instance
}
type ValuePathWaypoint struct {
	Position rbxfile.ValueVector3
	Action   uint32
}

type ValueToken struct {
	ID    uint16
	Value uint32
}

type ValueVector3int32 struct {
	X int32
	Y int32
	Z int32
}

type ValueDeferredString struct {
	Hash  string
	Value rbxfile.ValueSharedString
}

type ValueSignedProtectedString struct {
	Signature []byte
	Value     []byte
}

func TypeString(val rbxfile.Value) string {
	if val == nil {
		return "nil"
	}
	thisType := val.Type()
	if thisType < TypeNumberSequenceKeypoint {
		return thisType.String()
	}
	return CustomTypeNames[thisType]
}

func (x ValueRegion3) String() string {
	return fmt.Sprintf("{%s}, {%s}", x.Start.String(), x.End.String())
}
func (x ValueRegion3int16) String() string {
	return fmt.Sprintf("{%s}, {%s}", x.Start.String(), x.End.String())
}

func (ValueSystemAddress) Type() rbxfile.Type {
	return TypeSystemAddress
}
func (t ValueSystemAddress) String() string {
	return fmt.Sprintf("%d", t)
}
func (t ValueSystemAddress) Copy() rbxfile.Value {
	return t
}

// The following types should never be copied
func (x ValueTuple) String() string {
	var ret bytes.Buffer
	ret.WriteString("[")

	for _, y := range x {
		if y == nil {
			ret.WriteString("nil, ")
			continue
		}
		ret.WriteString(fmt.Sprintf("(%s) %s, ", TypeString(y), y.String()))
	}

	ret.WriteString("]")
	return ret.String()
}

func (x ValueArray) String() string {
	return ValueTuple(x).String()
}

func (x ValueDictionary) String() string {
	var ret bytes.Buffer
	ret.WriteString("{")

	for k, v := range x {
		var thisValue string
		if v == nil {
			thisValue = "nil"
		} else {
			thisValue = fmt.Sprintf("(%s) %s", TypeString(v), v.String())
		}
		ret.WriteString(fmt.Sprintf("%s: %s, ", k, thisValue))
	}

	ret.WriteString("}")
	return ret.String()
}

func (x ValueMap) String() string {
	return ValueDictionary(x).String()
}

func (x ValueTuple) Copy() rbxfile.Value {
	return x // nop
}

func (x ValueTuple) Type() rbxfile.Type {
	return TypeTuple
}

func (x ValueArray) Copy() rbxfile.Value {
	return x
}
func (x ValueArray) Type() rbxfile.Type {
	return TypeArray
}

func (x ValueMap) Copy() rbxfile.Value {
	return x
}

func (x ValueMap) Type() rbxfile.Type {
	return TypeMap
}

func (x ValueDictionary) Copy() rbxfile.Value {
	return x
}

func (x ValueDictionary) Type() rbxfile.Type {
	return TypeDictionary
}

func (x ValueRegion3) Copy() rbxfile.Value {
	return x
}

func (x ValueRegion3) Type() rbxfile.Type {
	return TypeRegion3
}

func (x ValueRegion3int16) Copy() rbxfile.Value {
	return x
}

func (x ValueRegion3int16) Type() rbxfile.Type {
	return TypeRegion3int16
}

// WARNING: Remember to set val.Instance yourself when copying
func (x ValueReference) Copy() rbxfile.Value {
	x.Instance = nil
	return x
}
func (x ValueReference) Type() rbxfile.Type {
	return TypeReference
}
func (x ValueReference) String() string {
	return fmt.Sprintf("%s: %s", x.Reference.String(), x.Instance.GetFullName())
}

func (x ValueColorSequenceKeypoint) Type() rbxfile.Type {
	return TypeColorSequenceKeypoint
}
func (x ValueColorSequenceKeypoint) Copy() rbxfile.Value {
	return x
}
func (x ValueColorSequenceKeypoint) String() string {
	return (rbxfile.ValueColorSequenceKeypoint)(x).String()
}

func (x ValueNumberSequenceKeypoint) Type() rbxfile.Type {
	return TypeNumberSequenceKeypoint
}
func (x ValueNumberSequenceKeypoint) Copy() rbxfile.Value {
	return x
}
func (x ValueNumberSequenceKeypoint) String() string {
	return (rbxfile.ValueNumberSequenceKeypoint)(x).String()
}

func (x ValueColorSequence) Type() rbxfile.Type {
	return TypeColorSequence
}
func (x ValueColorSequence) Copy() rbxfile.Value {
	c := make(ValueColorSequence, len(x))
	copy(c, x)
	return c
}
func (x ValueColorSequence) String() string {
	b := make([]byte, 0, 64)
	for _, v := range x {
		b = append(b, []byte(v.String())...)
		b = append(b, ' ')
	}
	return string(b)
}

func (x ValueNumberSequence) Type() rbxfile.Type {
	return TypeNumberSequence
}
func (x ValueNumberSequence) Copy() rbxfile.Value {
	c := make(ValueNumberSequence, len(x))
	copy(c, x)
	return c
}
func (x ValueNumberSequence) String() string {
	b := make([]byte, 0, 64)
	for _, v := range x {
		b = append(b, []byte(v.String())...)
		b = append(b, ' ')
	}
	return string(b)
}

func (x ValueToken) Type() rbxfile.Type {
	return TypeToken
}
func (x ValueToken) Copy() rbxfile.Value {
	return x
}
func (x ValueToken) String() string {
	return rbxfile.ValueToken(x.Value).String()
}

func (x ValueVector3int32) Type() rbxfile.Type {
	return TypeVector3int32
}

func (x ValueVector3int32) Copy() rbxfile.Value {
	return x
}

func (x ValueVector3int32) String() string {
	return fmt.Sprintf("%d, %d, %d", x.X, x.Y, x.Z)
}

func (x ValuePathWaypoint) Type() rbxfile.Type {
	return TypePathWaypoint
}

func (x ValuePathWaypoint) String() string {
	return fmt.Sprintf("%s, action %d", x.Position, x.Action)
}

func (x ValuePathWaypoint) Copy() rbxfile.Value {
	return x
}

func (x *ValueDeferredString) Type() rbxfile.Type {
	return TypeDeferredString
}
func (x *ValueDeferredString) String() string {
	stringLen := len(x.Value)
	if stringLen == 0 {
		return fmt.Sprintf("Deferred; MD5=%X", x.Hash)
	}
	return fmt.Sprintf("%X (len %d)", x.Hash, stringLen)
}
func (x *ValueDeferredString) Copy() rbxfile.Value {
	return x
}

func (x ValueSignedProtectedString) Type() rbxfile.Type {
	return TypeSignedProtectedString
}
func (x ValueSignedProtectedString) String() string {
	return fmt.Sprintf("Signed string, len %d", len(x.Value))
}
func (x ValueSignedProtectedString) Copy() rbxfile.Value {
	newString := new(ValueSignedProtectedString)
	newString.Signature = append(newString.Signature, x.Signature...)
	newString.Value = append(newString.Value, x.Value...)
	return newString
}
