package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
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

// Caches represents a collection of network caches that are used by various packets
type Caches struct {
	String StringCache
}

// CommunicationContext represents a network communication's
// contextual data
type CommunicationContext struct {
	// InstanceTopScope is the server's default scope that is
	// mainly used by JoinData
	// TODO: make this use a getter and setter instead, so
	// that an error can be reported if it is accessed prematurely
	InstanceTopScope string

	// DataModel represents the hierarchical collection of instances
	// in the context of which the communication takes place
	DataModel *datamodel.DataModel
	// InstancesByReference provides a convenient way to access the DataModel's instances
	// when only their reference is known
	InstancesByReference *datamodel.InstanceList
	// NetworkSchema is the network instance/enum schema used in this communication
	NetworkSchema *NetworkSchema
	// IsStudio
	IsStudio bool
	// SharedStrings contains a dictionary of SharedStrings indexed by their MD5 hash
	SharedStrings map[string]rbxfile.ValueSharedString

	// ScriptKey and CoreScriptKey are decryption keys for the
	// replicated scripts as reported by the server
	ScriptKey     uint32
	CoreScriptKey uint32
	ServerPeerID  uint32

	PlaceID   int64
	VersionID Packet90VersionID

	uniqueID uint64
}

// NewCommunicationContext returns a new CommunicationContext
func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		DataModel:            datamodel.New(),
		InstancesByReference: datamodel.NewInstanceList(),
	}
}

// GenerateSubmitTicketKey generates a key to be used by ID_SUBMIT_TICKET packets
func (context *CommunicationContext) GenerateSubmitTicketKey() [0x10]byte {
	var result [0x10]byte
	str := fmt.Sprintf("%d%d%d%d",
		int32(context.PlaceID),
		context.VersionID[0],
		context.VersionID[2],
		context.VersionID[1])

	copy(result[:], []byte(str)[:0x10])

	return result
}

func (context *CommunicationContext) removeInstance(instance *datamodel.Instance) {
	context.InstancesByReference.RemoveTree(instance)
}
