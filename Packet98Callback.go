package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket98(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet98Layer)

	layerLayout.AddWidget(NewQLabelF("Kick message: %s", MainLayer.Message), 0, 0)
}
