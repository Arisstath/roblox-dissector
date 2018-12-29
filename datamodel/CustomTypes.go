package datamodel

import (
	"bytes"
	"fmt"
	"net"

	"github.com/robloxapi/rbxfile"
)

type Reference struct {
	IsNull bool
	Scope  string
	Id     uint32
}

func (ref Reference) String() string {
	if ref.IsNull {
		return "null"
	}
	return fmt.Sprintf("%s_%d", ref.Scope, ref.Id)
}

const (
	TypeNumberSequenceKeypoint rbxfile.Type = rbxfile.TypeInt64 + 1
	TypeColorSequenceKeypoint               = iota
	TypeSystemAddress
	TypeMap
	TypeDictionary
	TypeArray
	TypeTuple
	TypeRegion3
	TypeRegion3int16
	TypeReference
)

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
type ValueSystemAddress net.UDPAddr
type ValueReference struct {
	Reference Reference
	Instance  *Instance
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
	return t.String()
}
func (t ValueSystemAddress) Copy() rbxfile.Value {
	c := new(net.UDPAddr)
	copy(c.IP, t.IP)
	c.Port = t.Port
	return ValueSystemAddress(*c)
}

// The following types should never be copied
func (x ValueTuple) String() string {
	var ret bytes.Buffer
	ret.WriteString("[")

	for _, y := range x {
		if y == nil {
			ret.WriteString("nil, ")
		}
		ret.WriteString(fmt.Sprintf("(%s) %s, ", y.Type().String(), y.String()))
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
		ret.WriteString(fmt.Sprintf("%s: (%s) %s, ", k, v.Type().String(), v.String()))
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

func (x ValueReference) Copy() rbxfile.Value {
	return x
}
func (x ValueReference) Type() rbxfile.Type {
	return TypeReference
}
func (x ValueReference) String() string {
	return fmt.Sprintf("%s: %s", x.Reference.String(), x.Instance.GetFullName())
}
