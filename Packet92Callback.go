package main
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket92(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet92Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Unknown int: %08X", MainLayer.UnknownValue), 0, 0)
}
