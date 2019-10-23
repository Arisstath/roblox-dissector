package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
	//_ "net/http/pprof"
)

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
	0x8D: ShowPacket8D,
	0x8F: ShowPacket8F,
	0x90: ShowPacket90,
	0x92: ShowPacket92,
	0x93: ShowPacket93,
	0x96: ShowPacket96,
	0x97: ShowPacket97,
	0x9B: ShowPacket9B,
}

func main() {
	/*go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/
	GUIMain()
}
