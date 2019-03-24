package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket8F(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8FLayer)

	layerLayout.AddWidget(NewQLabelF("Initial spawn name: %s", MainLayer.SpawnName), 0, 0)
}
