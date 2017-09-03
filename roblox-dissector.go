package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/pcap"
import "github.com/google/gopacket/layers"
import "github.com/fatih/color"
import "fmt"
import "github.com/gskartwii/roblox-dissector/peer"
import "time"

const DEBUG bool = false

var PacketNames map[byte]string = map[byte]string{
	0xFF: "???",
	0x00: "ID_CONNECTED_PING",
	0x01: "ID_UNCONNECTED_PING",
	0x03: "ID_CONNECTED_PONG",
	0x04: "ID_DETECT_LOST_CONNECTIONS",
	0x05: "ID_OPEN_CONNECTION_REQUEST_1",
	0x06: "ID_OPEN_CONNECTION_REPLY_1",
	0x07: "ID_OPEN_CONNECTION_REQUEST_2",
	0x08: "ID_OPEN_CONNECTION_REPLY_2",
	0x09: "ID_CONNECTION_REQUEST",
	0x10: "ID_CONNECTION_REQUEST_ACCEPTED",
	0x11: "ID_CONNECTION_ATTEMPT_FAILED",
	0x13: "ID_NEW_INCOMING_CONNECTION",
	0x15: "ID_DISCONNECTION_NOTIFICATION",
	0x18: "ID_INVALID_PASSWORD",
	0x1B: "ID_TIMESTAMP",
	0x1C: "ID_UNCONNECTED_PONG",
	0x81: "ID_ROBLOX_INIT_INSTANCES",
	0x82: "ID_ROBLOX_DICTIONARIES",
	0x83: "ID_ROBLOX_REPLICATION",
	0x89: "ID_ROBLOX_REPORT_ABUSE",
	0x8A: "ID_ROBLOX_AUTH",
	0x8E: "ID_ROBLOX_PROTOCOL_MISMATCH",
	0x8F: "ID_ROBLOX_INITIAL_SPAWN_NAME",
	0x90: "ID_ROBLOX_SCHEMA_VERSION",
	0x91: "ID_ROBLOX_NETWORK_SCHEMA",
	0x92: "ID_ROBLOX_START_AUTH_THREAD",
	0x93: "ID_ROBLOX_NETWORK_PARAMS",
	0x94: "ID_ROBLOX_HASH_REJECTED",
	0x95: "ID_ROBLOX_SECURITY_KEY_REJECTED",
	0x97: "ID_ROBLOX_NEW_SCHEMA",
}

type ActivationCallback func(byte, *peer.UDPPacket, *peer.CommunicationContext, *peer.PacketLayers)
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
    0x97: ShowPacket97,
}

func captureJob(handle *pcap.Handle, useIPv4 bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	handle.SetBPFFilter("udp")
	var packetSource *gopacket.PacketSource
	if useIPv4 {
		packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
	} else {
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	}
	packetChannel := make(chan gopacket.Packet, 0x100) // Buffer the packets to avoid dropping them

	go func() {
		for packet := range packetSource.Packets() {
			packetChannel <- packet
		}
	}()

	packetReader := &peer.PacketReader{
		SimpleHandler: func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
			if context.IsValid {
				packetViewer.AddFullPacket(packetType, packet, context, layers, ActivationCallbacks[packetType])
			}
		},
		ReliableHandler: func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
			if context.IsValid {
				packetViewer.AddSplitPacket(packetType, packet, context, layers)
			}
		},
		FullReliableHandler: func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
			if context.IsValid {
				packetViewer.BindCallback(packetType, packet, context, layers, ActivationCallbacks[packetType])
			}
		},
		ErrorHandler: func(err error) {
			println(err.Error())
		},
		Context: context,
	}

	for true {
		select {
		case _ = <- stopCaptureJob:
			context.IsValid = false
			return
		case packet := <- packetChannel:
			if packet.ApplicationLayer() == nil {
				color.Red("Ignoring packet because ApplicationLayer can't be decoded")
				continue
			}
			payload := packet.ApplicationLayer().Payload()
			if len(payload) == 0 {
				println("Ignoring 0 payload")
				continue
			}
			if context.Client == "" && payload[0] != 5 {
				println("Ignoring non5")
				continue
			}
			newPacket := peer.UDPPacketFromGoPacket(packet)
			if context.Client != "" && !context.IsClient(newPacket.Source) && !context.IsServer(newPacket.Source) {
				continue
			}

			if newPacket != nil {
				packetReader.ReadPacket(payload, newPacket)
			}
		}
	}
	return
}

func captureFromFile(filename string, useIPv4 bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from file %s\n", filename)
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, stopCaptureJob, packetViewer, context)
}

func captureFromLive(livename string, useIPv4 bool, usePromisc bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from live device %s\n", livename)
	handle, err := pcap.OpenLive(livename, 2000, usePromisc, 10 * time.Second)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, stopCaptureJob, packetViewer, context)
}

func main() {
	GUIMain()
}
