package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/pcap"
import "github.com/google/gopacket/layers"
import "github.com/fatih/color"
import "flag"
import "fmt"
import "github.com/gskartwii/go-bitstream"
import "bytes"

type PacketLayers struct {
	RakNet *RakNetLayer
	Reliability *ReliablePacket
	Main interface{}
}

var PacketNames map[byte]string = map[byte]string{
	0xFF: "???",
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
type DecoderFunc func(*ExtendedReader, *CommunicationContext, gopacket.Packet) (interface{}, error)

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

type ActivationCallback func(byte, gopacket.Packet, *CommunicationContext, *PacketLayers)
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
	//0x83: ShowPacket83, // Doesn't exist yet
}

func HandleSimple(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	layers := &PacketLayers{}
	layers.RakNet = layer

	packetType, err := layer.Payload.ReadByte()
	if err != nil {
		color.Red("Failed to decode simple packet: %s", err.Error())
		return
	}
	decoder := PacketDecoders[packetType]
	if decoder != nil {
		layers.Main, err = decoder(layer.Payload, context, packet)
		if err != nil {
			color.Red("Failed to decode simple packet %02X: %s", packetType, err.Error())
			return
		}
	}
	packetViewer.AddFullPacket(packetType, packet, context, layers, ActivationCallbacks[packetType])
}

func HandleACK(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	for _, ACK := range layer.ACKs {
		packetViewer.AddACK(ACK, packet, context, layer, func() {})
	}
}

func HandleGeneric(layer *RakNetLayer, packet gopacket.Packet, context *CommunicationContext, packetViewer *MyPacketListView) {
	reliabilityLayer, err := DecodeReliabilityLayer(layer.Payload, context, packet, layer)
	if err != nil {
		color.Red("Failed to decode reliable packet: %s", err.Error())
		return
	}

	for _, subPacket := range reliabilityLayer.Packets {
		layers := &PacketLayers{}
		layers.RakNet = layer
		layers.Reliability = subPacket
		packetViewer.AddSplitPacket(subPacket.PacketType, packet, context, layers)

		if subPacket.HasPacketType && !subPacket.HasBeenDecoded {
			subPacket.HasBeenDecoded = true
			go func(subPacket *ReliablePacket) {
				packetType := subPacket.PacketType
				_, err = subPacket.FullDataReader.ReadByte() // Void first byte, since we can get it the other way
				if err != nil {
					color.Red("Failed to decode reliablePacket %02X: %s", packetType, err.Error())
					return
				}

				decoder := PacketDecoders[packetType]
				if decoder != nil {
					layers.Main, err = decoder(subPacket.FullDataReader, context, packet)
					if err != nil {
						color.Red("Failed to decode reliable packet %02X: %s", packetType, err.Error())
						return
					}
				}

				packetViewer.BindCallback(packetType, packet, context, layers, ActivationCallbacks[packetType])
			}(subPacket)
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
	live := flag.String("live", "", "Live interface to capture from")
	promisc := flag.Bool("promisc", false, "Capture from live interface in promisc. mode")
	flag.Parse()

	var handle *pcap.Handle
	var err error
	if *live == "" {
		fmt.Printf("Will capture from file %s\n", *packetName)
		handle, err = pcap.OpenOffline(*packetName)
	} else {
		fmt.Printf("Will capture from live device %s\n", *live)
		handle, err = pcap.OpenLive(*live, 2000, *promisc, pcap.BlockForever)
	}
	if err == nil {
		handle.SetBPFFilter("udp")
		var packetSource *gopacket.PacketSource
		if *ipv4 {
			packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
		} else {
			packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
		}
		context := NewCommunicationContext()
		for packet := range packetSource.Packets() {
			if packet.ApplicationLayer() == nil {
				color.Red("Ignoring packet because ApplicationLayer can't be decoded")
				continue
			}
			payload := packet.ApplicationLayer().Payload()
			if len(payload) == 0 {
				color.Red("Had 0 size payload")
				continue
			}

			thisBitstream := &ExtendedReader{bitstream.NewReader(bytes.NewReader(payload))}

			rakNetLayer, err := DecodeRakNetLayer(payload[0], thisBitstream, context, packet)
			if err != nil {
				color.Red("Failed to decode RakNet layer: %s", err.Error())
				continue
			}
			if rakNetLayer.IsSimple {
				HandleSimple(rakNetLayer, packet, context, packetViewer)
			} else if !rakNetLayer.IsValid {
				color.New(color.FgRed).Printf("Sent invalid packet (packet header %x)\n", payload[0])
			} else if rakNetLayer.IsACK {
				HandleACK(rakNetLayer, packet, context, packetViewer)
			} else if !rakNetLayer.IsNAK {
				HandleGeneric(rakNetLayer, packet, context, packetViewer)
			}
		}
	} else {
		color.Red("Failed to create packet source: %s", err.Error())
	}
	<- done
}
