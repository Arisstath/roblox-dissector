package peer
import "sync"
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
	return bytes.Compare(val1.([]byte), val.([]byte)) == 0, val1 != nil
}
func (c *ByteSliceCache) LastWrite() uint8 {
	return c.lastWrite
}

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor Descriptor
	PropertyDescriptor Descriptor
	EventDescriptor Descriptor
	TypeDescriptor Descriptor
	ReplicatorStringCache StringCache
	ReplicatorObjectCache StringCache
	ReplicatorContentCache StringCache
	ReplicatorSystemAddressCache SysAddrCache
	ReplicatorProtectedStringCache ByteSliceCache

	DataModel *rbxfile.Root
    InstancesByReferent InstanceList

	MDescriptor *sync.Mutex
	MSchema *sync.Mutex

	UniqueID uint32

	EDescriptorsParsed *sync.Cond
	ESchemaParsed *sync.Cond

	UseStaticSchema bool
	StaticSchema *StaticSchema

	IsStudio bool
	IsValid bool

	SplitPackets SplitPacketList
	MUniques *sync.Mutex
	UniqueDatagramsClient map[uint32]struct{}
	UniqueDatagramsServer map[uint32]struct{}
}

func NewCommunicationContext() *CommunicationContext {
	MDescriptor := &sync.Mutex{}
	MSchema := &sync.Mutex{}
	return &CommunicationContext{
		ClassDescriptor: make(Descriptor),
		PropertyDescriptor: make(Descriptor),
		EventDescriptor: make(Descriptor),
		TypeDescriptor: make(Descriptor),

		MDescriptor: MDescriptor,
		MSchema: MSchema,

		EDescriptorsParsed: sync.NewCond(MDescriptor),
		ESchemaParsed: sync.NewCond(MSchema),
		IsValid: true,

		MUniques: &sync.Mutex{},
		UniqueDatagramsClient: make(map[uint32]struct{}),
		UniqueDatagramsServer: make(map[uint32]struct{}),
        InstancesByReferent: InstanceList{
            CommonMutex: MSchema,
            EAddReferent: sync.NewCond(MSchema),
            Instances: make(map[string]*rbxfile.Instance),
        },
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

func (c *CommunicationContext) WaitForDescriptors() {
	c.MDescriptor.Lock()
	for len(c.ClassDescriptor) == 0 {
		c.EDescriptorsParsed.Wait()
	}
}
func (c *CommunicationContext) WaitForSchema() {
	c.MSchema.Lock()
	for c.StaticSchema == nil {
		c.ESchemaParsed.Wait()
	}
}

func (c *CommunicationContext) FinishDescriptors() {
	c.MDescriptor.Unlock()
}
func (c *CommunicationContext) FinishSchema() {
	c.MSchema.Unlock()
}

