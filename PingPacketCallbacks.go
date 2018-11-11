package main

import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket00(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.RakPing)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
}

func ShowPacket03(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.RakPong)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	sendPongTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPongTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
	layerLayout.AddWidget(sendPongTimeLabel, 0, 0)
}
