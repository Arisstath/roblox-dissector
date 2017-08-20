package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/layers"
import "strconv"
import "sync"
import "bytes"

type ACKRange struct {
	Min uint32
	Max uint32
}

type RakNetLayer struct {
	Payload *ExtendedReader
	IsSimple bool
	SimpleLayerID uint8
	IsValid bool
	IsACK bool
	IsNAK bool
	HasBAndAS bool
	ACKs []ACKRange
	IsPacketPair bool
	IsContinuousSend bool
	NeedsBAndAS bool
	DatagramNumber uint32
}

type Descriptor map[uint32]string
type Cache [0x80]interface{}

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor Descriptor
	PropertyDescriptor Descriptor
	EventDescriptor Descriptor
	TypeDescriptor Descriptor
	EnumSchema map[string]EnumSchemaItem
	InstanceSchema []*InstanceSchemaItem
	PropertySchema []*PropertySchemaItem
	EventSchema []*EventSchemaItem
	ReplicatorStringCache Cache
	ReplicatorObjectCache Cache
	ReplicatorContentCache Cache
	ReplicatorSystemAddressCache Cache
	ReplicatorProtectedStringCache Cache
	ReplicatorRebindObjectCache Cache
	Rebindables map[string]struct{}

	MDescriptor *sync.Mutex
	MSchema *sync.Mutex

	UniqueID uint32

	EDescriptorsParsed *sync.Cond
	ESchemaParsed *sync.Cond

	UseStaticSchema bool
	StaticInstanceSchema []StaticInstanceSchema
	StaticPropertySchema []StaticPropertySchema
	StaticEventSchema []StaticEventSchema

	IsStudio bool
	IsValid bool

	SplitPackets SplitPacketList
}

func NewCommunicationContext() *CommunicationContext {
	MDescriptor := &sync.Mutex{}
	MSchema := &sync.Mutex{}
	return &CommunicationContext{
		ClassDescriptor: make(map[uint32]string),
		PropertyDescriptor: make(map[uint32]string),
		EventDescriptor: make(map[uint32]string),
		TypeDescriptor: make(map[uint32]string),
		Rebindables: make(map[string]struct{}),

		MDescriptor: MDescriptor,
		MSchema: MSchema,

		EDescriptorsParsed: sync.NewCond(MDescriptor),
		ESchemaParsed: sync.NewCond(MSchema),
		IsValid: true,
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

func SourceInterfaceFromPacket(packet gopacket.Packet) string {
	if packet.Layer(layers.LayerTypeIPv4) == nil {
		return ""
	}
	return packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).SrcIP.String() + ":" + strconv.Itoa(int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort))
}
func DestInterfaceFromPacket(packet gopacket.Packet) string {
	if packet.Layer(layers.LayerTypeIPv4) == nil {
		return ""
	}
	return packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).DstIP.String() + ":" + strconv.Itoa(int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).DstPort))
}

func (c *CommunicationContext) PacketFromClient(packet gopacket.Packet) bool {
	return SourceInterfaceFromPacket(packet) == c.Client
}
func (c *CommunicationContext) PacketFromServer(packet gopacket.Packet) bool {
	return SourceInterfaceFromPacket(packet) == c.Server
}

func (c *CommunicationContext) WaitForDescriptors() {
	c.MDescriptor.Lock()
	for len(c.ClassDescriptor) == 0 {
		c.EDescriptorsParsed.Wait()
	}
}
func (c *CommunicationContext) WaitForSchema() {
	c.MSchema.Lock()
	for len(c.InstanceSchema) == 0 && !c.UseStaticSchema {
		c.ESchemaParsed.Wait()
	}
}

func (c *CommunicationContext) FinishDescriptors() {
	c.MDescriptor.Unlock()
}
func (c *CommunicationContext) FinishSchema() {
	c.MSchema.Unlock()
}

func NewRakNetLayer() *RakNetLayer {
	return &RakNetLayer{}
}

var OfflineMessageID = [...]byte{0x00,0xFF,0xFF,0x00,0xFE,0xFE,0xFE,0xFE,0xFD,0xFD,0xFD,0xFD,0x12,0x34,0x56,0x78}

func DecodeRakNetLayer(packetType byte, bitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (*RakNetLayer, error) {
	layer := NewRakNetLayer()

	var err error
	if packetType == 0x5 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		thisOfflineMessage := make([]byte, 0x10)
		err = bitstream.Bytes(thisOfflineMessage, 0x10)
		if err != nil {
			return layer, err
		}

		if bytes.Compare(thisOfflineMessage, OfflineMessageID[:]) != 0 {
			return layer, nil
		}

		client := SourceInterfaceFromPacket(packet)
		server := DestInterfaceFromPacket(packet)
		println("Automatically detected Roblox peers using 0x5 packet:")
		println("Client:", client)
		println("Server:", server)
		context.SetClient(client)
		context.SetServer(server)
		layer.SimpleLayerID = packetType
		layer.Payload = bitstream
		layer.IsSimple = true
		return layer, nil
	} else if packetType >= 0x6 && packetType <= 0x8 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		layer.IsSimple = true
		layer.Payload = bitstream
		layer.SimpleLayerID = packetType
		return layer, nil
	}

	layer.IsValid, err = bitstream.ReadBool()
	if !layer.IsValid {
		return layer, nil
	}
	if err != nil {
		return layer, err
	}
	layer.IsACK, err = bitstream.ReadBool()
	if err != nil {
		return layer, err
	}
	if layer.IsACK {
		layer.HasBAndAS, _ = bitstream.ReadBool()
		bitstream.Align()

		ackCount, _ := bitstream.ReadUint16BE()
		var i uint16
		for i = 0; i < ackCount; i++ {
			var min, max uint32

			minEqualToMax, _ := bitstream.ReadBoolByte()
			min, _ = bitstream.ReadUint24LE()
			if minEqualToMax {
				max = min
			} else {
				max, _ = bitstream.ReadUint24LE()
			}

			layer.ACKs = append(layer.ACKs, ACKRange{min, max})
		}
		return layer, nil
	} else {
		layer.IsNAK, _ = bitstream.ReadBool()
		layer.IsPacketPair, _ = bitstream.ReadBool()
		layer.IsContinuousSend, _ = bitstream.ReadBool()
		layer.NeedsBAndAS, _ = bitstream.ReadBool()
		bitstream.Align()

		layer.DatagramNumber, _ = bitstream.ReadUint24LE()
		layer.Payload = bitstream
		return layer, nil
	}
}
