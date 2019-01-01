package main

import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket96(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet96Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)
	layerLayout.AddWidget(NewQLabelF("Request: %v", MainLayer.Request), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Version: %d", MainLayer.Version), 0, 0)
}
