package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket05(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.ConnectionRequest1)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	versionLabel := NewQLabelF("Version: %d", MainLayer.ProtocolVersion)
	layerLayout.AddWidget(versionLabel, 0, 0)
}

func ShowPacket06(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.ConnectionReply1)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)
}

func ShowPacket07(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.ConnectionRequest2)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)
}

func ShowPacket08(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.ConnectionReply2)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)
}

func ShowPacket09(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet09Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	timeLabel := NewQLabelF("Timestamp: %d", MainLayer.Timestamp)
	layerLayout.AddWidget(timeLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)

	passwordLabel := NewQLabelF("Password: %X", MainLayer.Password)
	layerLayout.AddWidget(passwordLabel, 0, 0)
}

func ShowPacket10(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet10Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	labelForIPAddressList := NewQLabelF("Remote IP addresses:")
	layerLayout.AddWidget(labelForIPAddressList, 0, 0)

	systemIndexLabel := NewQLabelF("System index: %d", MainLayer.SystemIndex)
	layerLayout.AddWidget(systemIndexLabel, 0, 0)

	ipAddressList := widgets.NewQListWidget(nil)
	standardModel := gui.NewQStandardItemModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"IP address"})
	ipAddressList.SetModel(standardModel)
	ipAddressList.SetSelectionMode(0)

	ipAddressStrings := make([]string, 10)
	for i, address := range MainLayer.Addresses {
		ipAddressStrings[i] = address.String()
	}
	ipAddressList.AddItems(ipAddressStrings)

	layerLayout.AddWidget(ipAddressList, 0, 0)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	sendPongTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPongTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
	layerLayout.AddWidget(sendPongTimeLabel, 0, 0)
}

func ShowPacket13(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet13Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	labelForIPAddressList := NewQLabelF("Remote IP addresses:")
	layerLayout.AddWidget(labelForIPAddressList, 0, 0)

	ipAddressList := widgets.NewQListWidget(nil)
	standardModel := gui.NewQStandardItemModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"IP address"})
	ipAddressList.SetModel(standardModel)
	ipAddressList.SetSelectionMode(0)

	ipAddressStrings := make([]string, 10)
	for i, address := range MainLayer.Addresses {
		ipAddressStrings[i] = address.String()
	}
	ipAddressList.AddItems(ipAddressStrings)

	layerLayout.AddWidget(ipAddressList, 0, 0)

	sendPingTimeLabel := NewQLabelF("Send ping time: %d", MainLayer.SendPingTime)
	sendPongTimeLabel := NewQLabelF("Send pong time: %d", MainLayer.SendPongTime)
	layerLayout.AddWidget(sendPingTimeLabel, 0, 0)
	layerLayout.AddWidget(sendPongTimeLabel, 0, 0)
}
