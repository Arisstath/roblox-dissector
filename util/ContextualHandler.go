package util
import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/schema"
import "net"
import "bytes"
import "fmt"

type CommunicationContext struct {
    *InstanceList
	Server *net.UDPAddr
	Client *net.UDPAddr

	InstanceTopScope string

	DataModel           *rbxfile.Root

	// TODO: Can we do better?
	UniqueID uint32

	StaticSchema *schema.StaticSchema

	IsStudio bool

	Int1 uint32
	Int2 uint32
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		InstanceList: NewInstanceList(),
        InstanceTopScope: "WARNING_UNASSIGNED_TOP_SCOPE",
	}
}

func (c *CommunicationContext) IsClient(peer *net.UDPAddr) bool {
	return c.Client.Port == peer.Port && bytes.Compare(c.Client.IP, peer.IP) == 0
}
func (c *CommunicationContext) IsServer(peer *net.UDPAddr) bool {
	return c.Server.Port == peer.Port && bytes.Compare(c.Server.IP, peer.IP) == 0
}
type Reference struct {
    Scope string
    Id uint32
    IsNull bool
}

func NewReference(scope string, id uint32) Reference {
    return Reference{IsNull: id == 0, Scope: scope, Id: id}
}
func (ref *Reference) String() string {
    if ref.IsNull {
        return "nil"
    }
    return fmt.Sprintf("%s_%d", ref.Scope, ref.Id)
}

type ContextualHandler interface {
    SetCaches(*Caches)
	Caches() *Caches
    TryGetInstance(Reference) (*rbxfile.Instance, error)
    OnAddInstance(Reference, func(*rbxfile.Instance))
    CreateInstance(Reference) (*rbxfile.Instance, error)
    Schema() *schema.StaticSchema
}

// PacketReader is an interface that can be passed to packet decoders
type PacketReader interface {
    ContextualHandler
    SetIsClient(bool)
	IsClient() bool
    SetSchema(*schema.StaticSchema)
}

type PacketWriter interface {
    ContextualHandler
    SetToClient(bool)
    ToClient() bool
}
