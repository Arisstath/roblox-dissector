package peer

import "github.com/gskartwii/rbxfile"
import "net"
import "bytes"

type CommunicationContext struct {
	Server *net.UDPAddr
	Client *net.UDPAddr

	InstanceTopScope string

	DataModel           *rbxfile.Root
	InstancesByReferent InstanceList

	// TODO: Can we do better?
	UniqueID uint32

	StaticSchema *StaticSchema

	IsStudio bool

	Int1 uint32
	Int2 uint32
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		InstancesByReferent: InstanceList{
			Instances: make(map[string]*rbxfile.Instance),
		},
		InstanceTopScope: "WARNING_UNASSIGNED_TOP_SCOPE",
	}
}

func (c *CommunicationContext) IsClient(peer *net.UDPAddr) bool {
	return c.Client.Port == peer.Port && bytes.Compare(c.Client.IP, peer.IP) == 0
}
func (c *CommunicationContext) IsServer(peer *net.UDPAddr) bool {
	return c.Server.Port == peer.Port && bytes.Compare(c.Server.IP, peer.IP) == 0
}
