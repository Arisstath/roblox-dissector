package main

import "github.com/Gskartwii/roblox-dissector/peer"
import "github.com/therecipe/qt/widgets"
import "fmt"

func ShowPacket8A(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8ALayer)

	int1Label := widgets.NewQTextEdit2(fmt.Sprintf("Player Id: %d", MainLayer.PlayerId), nil)
	string1Label := widgets.NewQTextEdit2(fmt.Sprintf("Client ticket: %s", MainLayer.ClientTicket), nil)
	string2Label := widgets.NewQTextEdit2(fmt.Sprintf("Data model hash: %s", MainLayer.DataModelHash), nil)
	int2Label := widgets.NewQTextEdit2(fmt.Sprintf("Protocol version: %d", MainLayer.ProtocolVersion), nil)
	string3Label := widgets.NewQTextEdit2(fmt.Sprintf("Security key: %s", MainLayer.SecurityKey), nil)
	string4Label := widgets.NewQTextEdit2(fmt.Sprintf("Platform: %s", MainLayer.Platform), nil)
	string5Label := widgets.NewQTextEdit2(fmt.Sprintf("Roblox product name: %s", MainLayer.RobloxProductName), nil)
	string6Label := widgets.NewQTextEdit2(fmt.Sprintf("Session Id: %s", MainLayer.SessionId), nil)
	int3Label := widgets.NewQTextEdit2(fmt.Sprintf("Golden hash: %8X", MainLayer.GoldenHash), nil)
	layerLayout.AddWidget(int1Label, 0, 0)
	layerLayout.AddWidget(string1Label, 0, 0)
	layerLayout.AddWidget(string2Label, 0, 0)
	layerLayout.AddWidget(int2Label, 0, 0)
	layerLayout.AddWidget(string3Label, 0, 0)
	layerLayout.AddWidget(string4Label, 0, 0)
	layerLayout.AddWidget(string5Label, 0, 0)
	layerLayout.AddWidget(string6Label, 0, 0)
	layerLayout.AddWidget(int3Label, 0, 0)
}
