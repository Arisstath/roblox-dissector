package main

import "github.com/Gskartwii/roblox-dissector/peer"
import "github.com/therecipe/qt/widgets"
import "fmt"

func ShowPacket8A(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8ALayer)

	// We use QTextEdits here so that the data can be easily copied
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Player Id: %d", MainLayer.PlayerID), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Client ticket: %s", MainLayer.ClientTicket), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Luau hash: %08X", MainLayer.LuauResponse), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Data model hash: %s", MainLayer.DataModelHash), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Protocol version: %d", MainLayer.ProtocolVersion), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Security key: %s", MainLayer.SecurityKey), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Platform: %s", MainLayer.Platform), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Roblox product name: %s", MainLayer.RobloxProductName), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Crypto Hash: %s", MainLayer.CryptoHash), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Session Id: %s", MainLayer.SessionID), nil), 0, 0)
	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Golden hash: %08X", MainLayer.GoldenHash), nil), 0, 0)
}
