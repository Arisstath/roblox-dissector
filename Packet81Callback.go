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
	referenceStringLabel := NewQLabelF("Top replication scope: %s", MainLayer.ReferenceString)
	layerLayout.AddWidget(referenceStringLabel, 0, 0)
	layerLayout.AddWidget(NewQLabelF("Script key: %X", MainLayer.ScriptKey), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Core script key: %X", MainLayer.CoreScriptKey), 0, 0)

	containerList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Class name", "Reference", "Watch properties", "Watch children"})

	containerListRootNode := standardModel.InvisibleRootItem()
	for i, item := range MainLayer.Items {
		indexItem := NewUintItem(i)
		classNameItem := NewStringItem(item.Schema.Name)
		referenceItem := NewStringItem(item.Instance.Ref.String())
		watchChangesItem := NewQStandardItemF("%v", item.WatchChanges)
		watchChildrenItem := NewQStandardItemF("%v", item.WatchChildren)
		containerListRootNode.AppendRow([]*gui.QStandardItem{indexItem, classNameItem, referenceItem, watchChangesItem, watchChildrenItem})
	}

	containerList.SetModel(standardModel)
	containerList.SetSelectionMode(0)
	containerList.SetSortingEnabled(true)
	containerList.SortByColumn(0, core.Qt__AscendingOrder)

	layerLayout.AddWidget(containerList, 0, 0)
}
