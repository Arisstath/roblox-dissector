package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket86(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet86Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	labelForSubpackets := NewQLabelF("Touch replication (%d entries):", len(MainLayer.SubPackets))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	subpacketList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name 1", "Reference 1", "Name 2", "Reference 2", "Touch type"})

	rootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.SubPackets {
		name1Item := NewQStandardItemF(item.Instance1.Name())
		reference1Item := NewQStandardItemF(item.Instance1.Reference)
		name2Item := NewQStandardItemF(item.Instance2.Name())
		reference2Item := NewQStandardItemF(item.Instance2.Reference)
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
