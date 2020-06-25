package main

import (
	"fmt"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func capabilitiesToString(cap uint64) string {
	var builder strings.Builder
	if cap&peer.CapabilityBasic == peer.CapabilityBasic {
		cap ^= peer.CapabilityBasic
		builder.WriteString("Basic, ")
	}
	if cap&peer.CapabilityServerCopiesPlayerGui3 != 0 {
		cap ^= peer.CapabilityServerCopiesPlayerGui3
		builder.WriteString("ServerCopiesPlayerGui, ")
	}
	if cap&peer.CapabilityDebugForceStreamingEnabled != 0 {
		cap ^= peer.CapabilityDebugForceStreamingEnabled
		builder.WriteString("DebugForceStreamingEnabled, ")
	}
	if cap&peer.CapabilityIHasMinDistToUnstreamed != 0 {
		cap ^= peer.CapabilityIHasMinDistToUnstreamed
		builder.WriteString("IHasMinDistToUnstreamed, ")
	}
	if cap&peer.CapabilityReplicateLuau != 0 {
		cap ^= peer.CapabilityReplicateLuau
		builder.WriteString("ReplicateLuau, ")
	}
	if cap&peer.CapabilityPositionBasedStreaming != 0 {
		cap ^= peer.CapabilityPositionBasedStreaming
		builder.WriteString("PositionBasedStreaming, ")
	}
	if cap&peer.CapabilityVersionedIDSync != 0 {
		cap ^= peer.CapabilityVersionedIDSync
		builder.WriteString("VersionedIDSync, ")
	}
	if cap&peer.CapabilityPubKeyExchange != 0 {
		cap ^= peer.CapabilityPubKeyExchange
		builder.WriteString("PubKeyExchange, ")
	}
	if cap&peer.CapabilitySystemAddressIsPeerId != 0 {
		cap ^= peer.CapabilitySystemAddressIsPeerId
		builder.WriteString("SystemAddressIsPeerId, ")
	}
	if cap&peer.CapabilityStreamingPrefetch != 0 {
		cap ^= peer.CapabilityStreamingPrefetch
		builder.WriteString("StreamingPrefetch, ")
	}
	if cap&peer.CapabilityTerrainReplicationUseLargerChunks != 0 {
		cap ^= peer.CapabilityTerrainReplicationUseLargerChunks
		builder.WriteString("TerrainReplicationUseLargerChunks , ")
	}
	if cap&peer.CapabilityUseBlake2BHashInSharedString != 0 {
		cap ^= peer.CapabilityUseBlake2BHashInSharedString
		builder.WriteString("UseBlake2BHashInSharedString, ")
	}
	if cap&peer.CapabilityUseSharedStringForScriptReplication != 0 {
		cap ^= peer.CapabilityUseSharedStringForScriptReplication
		builder.WriteString("UseSharedStringForScriptReplication, ")
	}

	if cap != 0 {
		fmt.Fprintf(&builder, "Unknown capabilities: %8X, ", cap)
	}

	if builder.Len() == 0 {
		return ""
	}
	return builder.String()[:builder.Len()-2]
}

func ShowPacket05(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet05Layer)

	versionLabel := NewQLabelF("Version: %d", MainLayer.ProtocolVersion)
	layerLayout.AddWidget(versionLabel, 0, 0)
}

func ShowPacket06(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet06Layer)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)
}

func ShowPacket07(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet07Layer)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	supportedVersionLabel := NewQLabelF("Supported version: %d", MainLayer.SupportedVersion)
	layerLayout.AddWidget(supportedVersionLabel, 0, 0)

	capabilitiesLabel := NewQLabelF("Capabilities: %s", capabilitiesToString(MainLayer.Capabilities))
	layerLayout.AddWidget(capabilitiesLabel, 0, 0)
}

func ShowPacket08(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet08Layer)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	mtuLabel := NewQLabelF("MTU Size: %d", MainLayer.MTU)
	layerLayout.AddWidget(mtuLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)

	supportedVersionLabel := NewQLabelF("Supported version: %d", MainLayer.SupportedVersion)
	layerLayout.AddWidget(supportedVersionLabel, 0, 0)

	capabilitiesLabel := NewQLabelF("Capabilities: %s", capabilitiesToString(MainLayer.Capabilities))
	layerLayout.AddWidget(capabilitiesLabel, 0, 0)
}

func ShowPacket09(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet09Layer)

	guidLabel := NewQLabelF("GUID: %08X", MainLayer.GUID)
	layerLayout.AddWidget(guidLabel, 0, 0)

	timeLabel := NewQLabelF("Timestamp: %d", MainLayer.Timestamp)
	layerLayout.AddWidget(timeLabel, 0, 0)

	useSecurityLabel := NewQLabelF("Use security: %v", MainLayer.UseSecurity)
	layerLayout.AddWidget(useSecurityLabel, 0, 0)

	passwordLabel := NewQLabelF("Password: %X", MainLayer.Password)
	layerLayout.AddWidget(passwordLabel, 0, 0)
}

func ShowPacket10(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet10Layer)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	labelForIPAddressList := NewLabel("Remote IP addresses:")
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

func ShowPacket13(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet13Layer)

	addressLabel := NewQLabelF("IP address: %s", MainLayer.IPAddress.String())
	layerLayout.AddWidget(addressLabel, 0, 0)

	labelForIPAddressList := NewLabel("Remote IP addresses:")
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
