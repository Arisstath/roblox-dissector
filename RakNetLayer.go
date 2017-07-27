package main
import "bytes"
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "github.com/google/gopacket/layers"
import "strconv"

type ACKRange struct {
	Min uint32
	Max uint32
}

type RakNetLayer struct {
	Contents []byte
	Payload []byte
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

type CommunicationContext struct {
	Server string
	Client string
	ClassDescriptor map[uint32]string
	PropertyDescriptor map[uint32]string
	EventDescriptor map[uint32]string
	TypeDescriptor map[uint32]string
	EnumSchema []*EnumSchemaItem
	InstanceSchema []*InstanceSchemaItem
	ReplicatorStringCache [0x80][]byte
}

func NewCommunicationContext() *CommunicationContext {
	return &CommunicationContext{
		ClassDescriptor: make(map[uint32]string),
		PropertyDescriptor: make(map[uint32]string),
		EventDescriptor: make(map[uint32]string),
		TypeDescriptor: make(map[uint32]string),
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

func NewRakNetLayer() RakNetLayer {
	return RakNetLayer{Payload: make([]byte, 0), Contents: make([]byte, 0)}
}

func DecodeRakNetLayer(data []byte, context *CommunicationContext, packet gopacket.Packet) (RakNetLayer, error) {
	layer := NewRakNetLayer()

	if data[0] == 0x5 {
		context.SetClient(SourceInterfaceFromPacket(packet))
		context.SetServer(DestInterfaceFromPacket(packet))
		layer.SimpleLayerID = data[0]
		layer.Payload = data
		layer.IsSimple = true
		return layer, nil
	} else if data[0] >= 0x6 && data[0] <= 0x8 {
		layer.IsSimple = true
		layer.Payload = data
		layer.SimpleLayerID = data[0]
		return layer, nil
	}

	bitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data))}

	var err error
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
		layer.Contents = data[:4]
		layer.Payload = data[4:]
		return layer, nil
	}
}
