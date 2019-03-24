package main

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket90(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet90Layer)

	layerLayout.AddWidget(NewQLabelF("Schema version: %d", MainLayer.SchemaVersion), 0, 0)

	requestList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Flag name"})

	requestListRootNode := standardModel.InvisibleRootItem()
	for i := 0; i < len(MainLayer.RequestedFlags); i++ {
		requestListRootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("%s", MainLayer.RequestedFlags[i])})
	}

	requestList.SetModel(standardModel)
	requestList.SetSelectionMode(0)
	requestList.SetSortingEnabled(true)

	layerLayout.AddWidget(NewQLabelF("Requested flags:"), 0, 0)
	layerLayout.AddWidget(requestList, 0, 0)

	layerLayout.AddWidget(widgets.NewQTextEdit2(fmt.Sprintf("Join data: %s", MainLayer.JoinData), nil), 0, 0)
}
