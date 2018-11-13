package main

import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket92(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet92Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)
	layerLayout.AddWidget(NewQLabelF("Place id: %d", MainLayer.PlaceId), 0, 0)
}
