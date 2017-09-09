package peer
import "sync"
import "github.com/gskartwii/rbxfile"
import "net"

type Descriptor map[string]uint32
type Cache [0x80]interface{}

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor Descriptor
	PropertyDescriptor Descriptor
	EventDescriptor Descriptor
	TypeDescriptor Descriptor
	ReplicatorStringCache Cache
	ReplicatorObjectCache Cache
	ReplicatorContentCache Cache
	ReplicatorSystemAddressCache Cache
	ReplicatorProtectedStringCache Cache
	ReplicatorRebindObjectCache Cache

	DataModel *rbxfile.Root
    InstancesByReferent InstanceList
    RefStringsByReferent map[string]string

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
        RefStringsByReferent: make(map[string]string),
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

