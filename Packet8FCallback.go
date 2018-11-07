package main

import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket8F(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8FLayer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)
	layerLayout.AddWidget(NewQLabelF("Initial spawn name: %s", MainLayer.SpawnName), 0, 0)
}
