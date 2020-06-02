package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket87(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet87Layer)

	layerLayout.AddWidget(NewQLabelF("Instance: %s", MainLayer.Instance.GetFullName()), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Message: %s", MainLayer.Message), 0, 0)
}
