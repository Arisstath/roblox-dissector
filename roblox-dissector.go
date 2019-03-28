package main

import (
	"flag"
	"log"
	"net"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/therecipe/qt/widgets"

	"os"
	"regexp"
	"strconv"

	"net/http"
	_ "net/http"
	_ "net/http/pprof"
)

const DEBUG bool = false

type ActivationCallback func(*widgets.QVBoxLayout, *peer.CommunicationContext, *peer.PacketLayers)

var ActivationCallbacks map[byte]ActivationCallback = map[byte]ActivationCallback{
	0x00: ShowPacket00,
	0x03: ShowPacket03,
	0x05: ShowPacket05,
	0x06: ShowPacket06,
	0x07: ShowPacket07,
	0x08: ShowPacket08,
	0x09: ShowPacket09,
	0x10: ShowPacket10,
	0x13: ShowPacket13,
	0x15: ShowPacket15,

	0x81: ShowPacket81,
	0x83: ShowPacket83,
	0x85: ShowPacket85,
	0x86: ShowPacket86,
	0x8A: ShowPacket8A,
	0x8F: ShowPacket8F,
	0x90: ShowPacket90,
	0x92: ShowPacket92,
	0x93: ShowPacket93,
	0x96: ShowPacket96,
	0x97: ShowPacket97,
}

func SrcAndDestFromGoPacket(packet gopacket.Packet) (*net.UDPAddr, *net.UDPAddr) {
	return &net.UDPAddr{
			IP:   packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).SrcIP,
			Port: int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort),
			Zone: "udp",
		}, &net.UDPAddr{
			IP:   packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4).DstIP,
			Port: int(packet.Layer(layers.LayerTypeUDP).(*layers.UDP).DstPort),
			Zone: "udp",
		}
}

func main() {
	joinFlag := flag.String("join", "", "roblox-dissector:<authTicket>:<placeID>:<browserTrackerID>")
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	flag.Parse()
	if *joinFlag != "" {
		println("Received protocol invocation?")
		protocolRegex := regexp.MustCompile(`roblox-dissector:([0-9A-Fa-f]+):(\d+):(\d+)`)
		uri := *joinFlag
		parts := protocolRegex.FindStringSubmatch(uri)
		if len(parts) < 4 {
			println("invalid protocol invocation: ", os.Args[1])
		} else {
			customClient := peer.NewCustomClient()
			authTicket := parts[1]
			placeID, _ := strconv.Atoi(parts[2])
			browserTrackerId, _ := strconv.Atoi(parts[3])
			customClient.SecuritySettings = peer.Win10Settings()
			customClient.BrowserTrackerId = uint64(browserTrackerId)
			// No more guests! Roblox won't let us connect as one.

			customClient.Logger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
			customClient.ConnectWithAuthTicket(uint32(placeID), authTicket)
		}
		return
	}
	GUIMain(flag.Arg(0))
}
