package main
import "github.com/google/gopacket"

func ShowPacket90(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet90Layer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Schema version: %d", MainLayer.SchemaVersion), 0, 0)
}
