package main

import (
	"io/ioutil"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket9B(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet9BLayer)

	if layers.Root.FromClient {
		layerLayout.AddWidget(NewQLabelF("Challenge: %08X", MainLayer.Challenge), 0, 0)
		layerLayout.AddWidget(NewQLabelF("Response: %08X", MainLayer.Response), 0, 0)
	} else {
		layerLayout.AddWidget(NewQLabelF("Unknown int: %08X", MainLayer.Int1), 0, 0)
		layerLayout.AddWidget(NewQLabelF("Challenge: %08X", MainLayer.Challenge), 0, 0)
		dumpScriptButton := widgets.NewQPushButton2("Dump challenge script bytecode", nil)
		dumpScriptButton.ConnectReleased(func() {
			scriptLocation := widgets.QFileDialog_GetSaveFileName(dumpScriptButton, "Dump script bytecode...", "", "RBXC files (*.rbxc)", "", 0)
			err := ioutil.WriteFile(scriptLocation, MainLayer.Script, 0666)
			if err != nil {
				println("while dumping script: ", err.Error())
			}
		})
		layerLayout.AddWidget(dumpScriptButton, 0, 0)
	}
}
