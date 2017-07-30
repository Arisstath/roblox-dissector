package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"

func ShowPacket81(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet81Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	for i := 0; i < 5; i++ {
		thisLabel := NewQLabelF("Unknown boolean %d: %v", i, MainLayer.Bools[i])
		layerLayout.AddWidget(thisLabel, 0, 0)
	}
	string1Label := NewQLabelF("Unknown string: %s", MainLayer.String1)
	layerLayout.AddWidget(string1Label, 0, 0)

	deletedList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Unknown int 1", "Unknown string", "Unknown int 2"})

	deletedListRootNode := standardModel.InvisibleRootItem()
	for i := 0; i < len(MainLayer.Items); i++ {
		unknownInt1Item := NewQStandardItemF("%d", MainLayer.Items[i].Int1)
		unknownStringItem := NewQStandardItemF("%s", MainLayer.Items[i].String1)
		unknownInt2Item := NewQStandardItemF("%d", MainLayer.Items[i].Int2)
		deletedListRootNode.AppendRow([]*gui.QStandardItem{unknownInt1Item, unknownStringItem, unknownInt2Item})
	}

	deletedList.SetModel(standardModel)
	deletedList.SetSelectionMode(0)
	deletedList.SetSortingEnabled(true)

	layerLayout.AddWidget(deletedList, 0, 0)
}
