package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/Gskartwii/roblox-dissector/peer"

func ShowPacket86(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet86Layer)

	labelForSubpackets := NewQLabelF("Touch replication (%d entries):", len(MainLayer.SubPackets))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	subpacketList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name 1", "Reference 1", "Name 2", "Reference 2", "Touch type"})

	rootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.SubPackets {
		if item.Instance1 == nil || item.Instance2 == nil {
			rootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("nil!!")})
			continue
		}
		var name1Item, name2Item *gui.QStandardItem
		name1Item = NewQStandardItemF(item.Instance1.GetFullName())
		reference1Item := NewQStandardItemF(item.Instance1.Ref.String())
		name2Item = NewQStandardItemF(item.Instance2.GetFullName())
		reference2Item := NewQStandardItemF(item.Instance2.Ref.String())
		var typeItem *gui.QStandardItem
		if item.IsTouch {
			typeItem = NewQStandardItemF("Start touch")
		} else {
			typeItem = NewQStandardItemF("End touch")
		}

		rootNode.AppendRow([]*gui.QStandardItem{name1Item, reference1Item, name2Item, reference2Item, typeItem})
	}

	subpacketList.SetModel(standardModel)
	subpacketList.SetSelectionMode(0)
	subpacketList.SetSortingEnabled(true)
	layerLayout.AddWidget(subpacketList, 0, 0)
}
