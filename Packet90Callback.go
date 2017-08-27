package main
import "./peer"

func ShowPacket90(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(peer.Packet90Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Schema version: %d", MainLayer.SchemaVersion), 0, 0)
}
