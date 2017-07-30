package main
import "github.com/google/gopacket"

func ShowPacket00(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet00Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
}

func ShowPacket03(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet03Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	sendPongTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPongTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
	layerLayout.AddWidget(sendPongTimeLabel, 0, 0)
}
