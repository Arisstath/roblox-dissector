package main
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket00(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet00Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
}

func ShowPacket03(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet03Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	sendPongTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPongTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
	layerLayout.AddWidget(sendPongTimeLabel, 0, 0)
}
