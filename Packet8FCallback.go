package main
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket8F(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8FLayer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Initial spawn name: %s", MainLayer.SpawnName), 0, 0)
}
