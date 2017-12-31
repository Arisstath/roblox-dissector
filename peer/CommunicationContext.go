package peer
import "github.com/gskartwii/rbxfile"
import "net"
import "bytes"

type Descriptor map[string]uint32
type Cache interface{
	Get(uint8)(interface{}, bool)
	Put(interface{}, uint8)
	Equal(uint8, interface{})(bool, bool)
	LastWrite() uint8
}

type StringCache struct {
	Values [0x80]interface{}
	lastWrite uint8
}

func (c *StringCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}
func (c *StringCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(string)
	c.lastWrite = index
}
func (c *StringCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(string) == val.(string), val1 != nil
}
func (c *StringCache) LastWrite() uint8 {
	return c.lastWrite
}

type SysAddrCache struct {
	Values [0x80]interface{}
	lastWrite uint8
}
func (c *SysAddrCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}
func (c *SysAddrCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(rbxfile.ValueSystemAddress)
	c.lastWrite = index
}
func (c *SysAddrCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(rbxfile.ValueSystemAddress).String() == val.(rbxfile.ValueSystemAddress).String(), val1 != nil
}
func (c *SysAddrCache) LastWrite() uint8 {
	return c.lastWrite
}

type ByteSliceCache struct {
	Values [0x80]interface{}
	lastWrite uint8
}
func (c *ByteSliceCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}
func (c *ByteSliceCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.([]byte)
	c.lastWrite = index
}
func (c *ByteSliceCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return bytes.Compare(val1.([]byte), val.([]byte)) == 0, val1 != nil
}
func (c *ByteSliceCache) LastWrite() uint8 {
	return c.lastWrite
}

type Caches struct {
	String StringCache
	Object StringCache
	Content StringCache
	SystemAddress SysAddrCache
	ProtectedString ByteSliceCache
}

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor Descriptor
	PropertyDescriptor Descriptor
	EventDescriptor Descriptor
	TypeDescriptor Descriptor

	ClientCaches Caches
	ServerCaches Caches
	
	InstanceTopScope string

	DataModel *rbxfile.Root
    InstancesByReferent InstanceList

	UniqueID uint32

	UseStaticSchema bool
	StaticSchema *StaticSchema

	IsStudio bool
	IsValid bool

	splitPackets splitPacketList
	UniqueDatagramsClient map[uint32]struct{}
	UniqueDatagramsServer map[uint32]struct{}
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		ClassDescriptor: make(Descriptor),
		PropertyDescriptor: make(Descriptor),
		EventDescriptor: make(Descriptor),
		TypeDescriptor: make(Descriptor),

		IsValid: true,

		UniqueDatagramsClient: make(map[uint32]struct{}),
		UniqueDatagramsServer: make(map[uint32]struct{}),
        InstancesByReferent: InstanceList{
            Instances: make(map[string]*rbxfile.Instance),
        },
		InstanceTopScope: "WARNING_UNASSIGNED_TOP_SCOPE",
	}
}

func (c *CommunicationContext) SetServer(server string) {
	c.Server = server
}
func (c *CommunicationContext) SetClient(client string) {
	c.Client = client
}
func (c *CommunicationContext) GetClient() string {
	return c.Client
}
func (c *CommunicationContext) GetServer() string {
	return c.Server
}

func (c *CommunicationContext) IsClient(peer net.UDPAddr) bool {
	return peer.String() == c.Client
}
func (c *CommunicationContext) IsServer(peer net.UDPAddr) bool {
	return peer.String() == c.Server
}
