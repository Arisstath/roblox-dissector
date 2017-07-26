package main
import "github.com/google/gopacket"

func ShowPacket92(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet92Layer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Unknown int: %08X", MainLayer.UnknownValue), 0, 0)
}
