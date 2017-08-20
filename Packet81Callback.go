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
	standardModel.SetHorizontalHeaderLabels([]string{"Class name", "Referent"})

	deletedListRootNode := standardModel.InvisibleRootItem()
	for i := 0; i < len(MainLayer.Items); i++ {
		unknownInt1Item := NewQStandardItemF("%s", context.StaticSchema.Instances[MainLayer.Items[i].ClassID].Name)
		unknownStringItem := NewQStandardItemF("%s", MainLayer.Items[i].Object1.Show())
		deletedListRootNode.AppendRow([]*gui.QStandardItem{unknownInt1Item, unknownStringItem})
	}

	deletedList.SetModel(standardModel)
	deletedList.SetSelectionMode(0)
	deletedList.SetSortingEnabled(true)

	layerLayout.AddWidget(deletedList, 0, 0)
}
