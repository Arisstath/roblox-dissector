package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/pcap"
import "github.com/google/gopacket/layers"
import "github.com/fatih/color"
import "flag"

type PacketLayers struct {
	RakNet *RakNetLayer
	Reliability *ReliablePacket
	Main interface{}
}

var PacketNames map[byte]string = map[byte]string{
	0x05: "ID_OPEN_CONNECTION_REQUEST_1",
	0x06: "ID_OPEN_CONNECTION_REPLY_1",
	0x07: "ID_OPEN_CONNECTION_REQUEST_2",
	0x08: "ID_OPEN_CONNECTION_REPLY_2",
	0x00: "ID_CONNECTED_PING",
	0x01: "ID_UNCONNECTED_PING",
	0x03: "ID_CONNECTED_PONG",
	0x04: "ID_DETECT_LOST_CONNECTIONS",
	0x09: "ID_CONNECTION_REQUEST",
	0x10: "ID_CONNECTION_REQUEST_ACCEPTED",
	0x13: "ID_NEW_INCOMING_CONNECTION",
	0x1B: "ID_ROBLOX_PHYSICS",
	0x1C: "ID_UNCONNECTED_PONG",
	0x82: "ID_ROBLOX_DICTIONARIES",
	0x91: "ID_ROBLOX_NETWORK_SCHEMA",
	0x8A: "ID_ROBLOX_AUTH",
	0x93: "ID_ROBLOX_NETWORK_PARAMS",
	0x92: "ID_ROBLOX_START_AUTH_THREAD",
	0x8F: "ID_ROBLOX_INITIAL_SPAWN_NAME",
	0x90: "ID_ROBLOX_SCHEMA_VERSION",
	0x81: "ID_ROBLOX_PRESCHEMA",
	0x83: "ID_ROBLOX_REPLICATION",
}
type DecoderFunc func([]byte, *CommunicationContext, gopacket.Packet) (interface{}, error)

var PacketDecoders map[byte]DecoderFunc = map[byte]DecoderFunc{
	0x05: DecodePacket05Layer,
	0x06: DecodePacket06Layer,
	0x07: DecodePacket07Layer,
	0x08: DecodePacket08Layer,
	0x00: DecodePacket00Layer,
	0x03: DecodePacket03Layer,
	0x09: DecodePacket09Layer,
	0x10: DecodePacket10Layer,
	0x13: DecodePacket13Layer,

	//0x8A: DecodePacket8ALayer,
	0x82: DecodePacket82Layer,
	0x93: DecodePacket93Layer,
	0x91: DecodePacket91Layer,
	0x92: DecodePacket92Layer,
	0x90: DecodePacket90Layer,
	0x8F: DecodePacket8FLayer,
	0x81: DecodePacket81Layer,
	0x83: DecodePacket83Layer,
}

type ActivationCallback func([]byte, gopacket.Packet, *CommunicationContext, *PacketLayers)
var ActivationCallbacks map[byte]ActivationCallback = map[byte]ActivationCallback{
	0x05: ShowPacket05,
	0x06: ShowPacket06,
	0x07: ShowPacket07,
	0x08: ShowPacket08,
	0x09: ShowPacket09,
	0x10: ShowPacket10,
	0x13: ShowPacket13,
	0x00: ShowPacket00,
	0x03: ShowPacket03,

	0x93: ShowPacket93,
	//0x8A: ShowPacket8A,
	0x82: ShowPacket82,
	0x91: ShowPacket91,
	0x92: ShowPacket92,
	0x90: ShowPacket90,
	0x8F: ShowPacket8F,
	0x81: ShowPacket81,
	0x83: ShowPacket83,
}

func HandleSimple(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	layers := &PacketLayers{}
	layers.RakNet = layer

	decoder := PacketDecoders[layer.Payload[0]]
	var err error
	if decoder != nil {
		layers.Main, err = decoder(layer.Payload, context, packet)
		if err != nil {
			color.Red("Failed to decode packet %02X: %s", layer.Payload[0], err.Error())
			return
		}
	}
	packetViewer.Add(layer.Payload, packet, context, layers, ActivationCallbacks[layer.Payload[0]])
}

func HandleACK(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	for _, ACK := range layer.ACKs {
		packetViewer.AddACK(ACK, packet, context, layer, func() {})
	}
}

func HandleGeneric(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	reliabilityLayer, err := DecodeReliabilityLayer(layer.Payload, context, packet, layer)
	if err != nil {
		color.Red("Failed to decode packet: %s", err.Error())
		return
	}

	for _, subPacket := range reliabilityLayer.Packets {
		if subPacket.IsFinal {
			finalData := subPacket.FinalData

			layers := &PacketLayers{}
			layers.RakNet = layer
			layers.Reliability = subPacket

			decoder := PacketDecoders[finalData[0]]
			var err error
			if decoder != nil {
				layers.Main, err = decoder(finalData, context, packet)
				if err != nil {
					color.Red("Failed to decode packet %02X: %s", finalData[0], err.Error())
					return
				}
			}
			packetViewer.Add(finalData, packet, context, layers, ActivationCallbacks[finalData[0]])
		}
	}
}

func main() {
	done := make(chan bool)
	packetViewerChan := make(chan *MyPacketListView)
	go GUIMain(done, packetViewerChan)
	packetViewer := <- packetViewerChan
	packetName := flag.String("name", "", "pcap filename")
	ipv4 := flag.Bool("ipv4", false, "Use IPv4 as initial frame type")
	flag.Parse()

	if handle, err := pcap.OpenOffline(*packetName); err == nil {
		handle.SetBPFFilter("udp")
		var packetSource *gopacket.PacketSource
		if *ipv4 {
			packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
		} else {
			packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
		}
		context := NewCommunicationContext()
		for packet := range packetSource.Packets() {
			payload := packet.ApplicationLayer().Payload()
			if len(payload) == 0 {
				continue
			}
			rakNetLayer, err := DecodeRakNetLayer(payload, context, packet)
			if err != nil {
				color.Red("Failed to decode RakNet layer: %s", err.Error())
				continue
			}
			if rakNetLayer.IsSimple {
				HandleSimple(&rakNetLayer, packet, context, packetViewer)
			} else if !rakNetLayer.IsValid {
				color.New(color.FgRed).Printf("Sent invalid packet (packet header %x)\n", payload[0])
			} else if rakNetLayer.IsACK {
				HandleACK(&rakNetLayer, packet, context, packetViewer)
			} else if !rakNetLayer.IsNAK {
				HandleGeneric(&rakNetLayer, packet, context, packetViewer)
			}
		}
	} else {
		color.Red("Failed to create packet source: %s", err.Error())
	}
	<- done
}
