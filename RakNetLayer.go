package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/layers"
import "strconv"
import "sync"

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

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor Descriptor
	PropertyDescriptor Descriptor
	EventDescriptor Descriptor
	TypeDescriptor Descriptor
	EnumSchema []*EnumSchemaItem
	InstanceSchema []*InstanceSchemaItem
	ReplicatorStringCache [0x80][]byte

	MClassDescriptor *sync.Mutex
	MPropertyDescriptor *sync.Mutex
	MEventDescriptor *sync.Mutex
	MTypeDescriptor *sync.Mutex
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		ClassDescriptor: make(map[uint32]string),
		PropertyDescriptor: make(map[uint32]string),
		EventDescriptor: make(map[uint32]string),
		TypeDescriptor: make(map[uint32]string),

		MClassDescriptor: &sync.Mutex{},
		MPropertyDescriptor: &sync.Mutex{},
		MEventDescriptor: &sync.Mutex{},
		MTypeDescriptor: &sync.Mutex{},
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

func PacketFromClient(packet gopacket.Packet, c *CommunicationContext) bool {
	return SourceInterfaceFromPacket(packet) == c.Client
}
func PacketFromServer(packet gopacket.Packet, c *CommunicationContext) bool {
	return SourceInterfaceFromPacket(packet) == c.Server
}

func NewRakNetLayer() *RakNetLayer {
	return &RakNetLayer{}
}

func DecodeRakNetLayer(packetType byte, bitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (*RakNetLayer, error) {
	layer := NewRakNetLayer()

	var err error
	if packetType == 0x5 {
		context.SetClient(SourceInterfaceFromPacket(packet))
		context.SetServer(DestInterfaceFromPacket(packet))
		layer.SimpleLayerID = packetType
		layer.Payload = bitstream
		layer.IsSimple = true
		return layer, nil
	} else if packetType >= 0x6 && packetType <= 0x8 {
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
