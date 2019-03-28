package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket81(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet81Layer)

	layerLayout.AddWidget(NewQLabelF("Stream job: %v", MainLayer.StreamJob), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Filtering enabled: %v", MainLayer.FilteringEnabled), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Allow third party sales: %v", MainLayer.AllowThirdPartySales), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Character auto spawn: %v", MainLayer.CharacterAutoSpawn), 0, 0)
	referentStringLabel := NewQLabelF("Top replication scope: %s", MainLayer.ReferentString)
	layerLayout.AddWidget(referentStringLabel, 0, 0)
	layerLayout.AddWidget(NewQLabelF("Int 1: %X", MainLayer.Int1), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Int 2: %X", MainLayer.Int2), 0, 0)

	containerList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Class name", "Referent", "Replicate properties", "Replicate children"})

	containerListRootNode := standardModel.InvisibleRootItem()
	for i, item := range MainLayer.Items {
		indexItem := NewUintItem(i)
		classNameItem := NewStringItem(item.Schema.Name)
		referenceItem := NewStringItem(item.Instance.Ref.String())
		repPropertiesItem := NewQStandardItemF("%v", item.Bool1)
		repChildrenItem := NewQStandardItemF("%v", item.Bool2)
		containerListRootNode.AppendRow([]*gui.QStandardItem{indexItem, classNameItem, referenceItem, repPropertiesItem, repChildrenItem})
	}

	containerList.SetModel(standardModel)
	containerList.SetSelectionMode(0)
	containerList.SetSortingEnabled(true)
	containerList.SortByColumn(0, core.Qt__AscendingOrder)

	layerLayout.AddWidget(containerList, 0, 0)
}
