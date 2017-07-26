package main
import "github.com/google/gopacket"

func ShowPacket8F(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet8FLayer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Initial spawn name: %s", MainLayer.SpawnName), 0, 0)
}
