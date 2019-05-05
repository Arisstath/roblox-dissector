package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket93(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet93Layer)

	layerLayout.AddWidget(NewQLabelF("Protocol schema sync: %v", MainLayer.ProtocolSchemaSync), 0, 0)
	layerLayout.AddWidget(NewQLabelF("API dictionary compression: %v", MainLayer.APIDictionaryCompression), 0, 0)

	labelForParamList := NewLabel("Network params:")
	layerLayout.AddWidget(labelForParamList, 0, 0)

	paramList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Value"})

	paramListRootNode := standardModel.InvisibleRootItem()
	for Name, Value := range MainLayer.Params {
		nameItem := NewStringItem(Name)
		valueItem := NewQStandardItemF("%v", Value)
		paramListRootNode.AppendRow([]*gui.QStandardItem{nameItem, valueItem})
	}

	paramList.SetModel(standardModel)
	paramList.SetSelectionMode(0)

	layerLayout.AddWidget(paramList, 0, 0)
}
