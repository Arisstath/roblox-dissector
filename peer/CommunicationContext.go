package peer

import (
	"bytes"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

// Cache represents a network cache that stores repeatable objects such as strings.
type Cache interface {
	// Get fetches the object matching a cache index.
	Get(uint8) (interface{}, bool)
	// Put creates a new object in the caches at the specific index.
	Put(interface{}, uint8)
	// Equal compares a cached object and another object and returns if
	// they are equal, and if the indexed object exists.
	Equal(uint8, interface{}) (bool, bool)
	// LastWrite returns the cache index that was last written to.
	LastWrite() uint8
}

// StringCache represents a cache that stores strings.
type StringCache struct {
	// Values contains the stored strings.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get()
func (c *StringCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put()
func (c *StringCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(string)
	c.lastWrite = index
}

// Equal implements Cache.Equal()
func (c *StringCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(string) == val.(string), val1 != nil
}

// LastWrite implements Cache.LastWrite()
func (c *StringCache) LastWrite() uint8 {
	return c.lastWrite
}

// SysAddrCache is a Cache that stores SystemAddress values.
type SysAddrCache struct {
	// Values contains the stores SystemAddresses.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *SysAddrCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *SysAddrCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(datamodel.ValueSystemAddress)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *SysAddrCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(datamodel.ValueSystemAddress).String() == val.(datamodel.ValueSystemAddress).String(), val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *SysAddrCache) LastWrite() uint8 {
	return c.lastWrite
}

// ByteSliceCache is a Cache that stores []byte objects.
type ByteSliceCache struct {
	// Values contains the []byte objects.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *ByteSliceCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *ByteSliceCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.([]byte)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *ByteSliceCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return bytes.Compare(val1.([]byte), val.([]byte)) == 0, val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *ByteSliceCache) LastWrite() uint8 {
	return c.lastWrite
}

type Caches struct {
	String          StringCache
	Object          StringCache
	Content         StringCache
	SystemAddress   SysAddrCache
	ProtectedString ByteSliceCache
}

// TODO: Remove CommunicationContext, or at least
// make it so that it's only used for PCAP captures
// where the server and client must be stored somewhere
type CommunicationContext struct {
	// TODO: Move this to reader and writer?
	InstanceTopScope string

	DataModel           *datamodel.DataModel
	InstancesByReference *datamodel.InstanceList

	// TODO: Can we do better?
	UniqueID uint32

	StaticSchema *StaticSchema

	IsStudio bool

	ScriptKey     uint32
	CoreScriptKey uint32
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		DataModel:           datamodel.New(),
		InstancesByReference: datamodel.NewInstanceList(),
		// TODO: Report an error if top scope is accessed before being assigned
		InstanceTopScope: "WARNING_UNASSIGNED_TOP_SCOPE",
	}
}
