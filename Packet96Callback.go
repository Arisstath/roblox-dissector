package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket96(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet96Layer)

	layerLayout.AddWidget(NewQLabelF("Request: %v", MainLayer.Request), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Version: %d", MainLayer.Version), 0, 0)
}
