package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket92(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet92Layer)

	layerLayout.AddWidget(NewQLabelF("Place id: %d", MainLayer.PlaceID), 0, 0)
}
