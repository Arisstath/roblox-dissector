package main

// Outdated!!
/*
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/widgets"
import "github.com/Gskartwii/roblox-dissector/peer"
import "os"
import "encoding/gob"

func ShowPacket82(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet82Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	labelForDescriptorView := NewQLabelF("Dictionaries:")
	layerLayout.AddWidget(labelForDescriptorView, 0, 0)

	descriptorView := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "IDx", "Unknown int"})

	dictionaryRootNode := standardModel.InvisibleRootItem()
	classDescriptorItem := NewQStandardItemF("ClassDescriptor (%d entries)", len(MainLayer.ClassDescriptor))

	for _, item := range MainLayer.ClassDescriptor {
		nameItem := NewQStandardItemF(item.Name)
		idXItem := NewQStandardItemF("%d", item.IDx)
		unknownIntItem := NewQStandardItemF("%d", item.OtherID)
		classDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{classDescriptorItem})

	propertyDescriptorItem := NewQStandardItemF("PropertyDescriptor (%d entries)", len(MainLayer.PropertyDescriptor))

	for _, item := range MainLayer.PropertyDescriptor {
		nameItem := NewQStandardItemF(item.Name)
		idXItem := NewQStandardItemF("%d", item.IDx)
		unknownIntItem := NewQStandardItemF("%d", item.OtherID)
		propertyDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{propertyDescriptorItem})

	eventDescriptorItem := NewQStandardItemF("EventDescriptor (%d entries)", len(MainLayer.EventDescriptor))

	for _, item := range MainLayer.EventDescriptor {
		nameItem := NewQStandardItemF(item.Name)
		idXItem := NewQStandardItemF("%d", item.IDx)
		unknownIntItem := NewQStandardItemF("%d", item.OtherID)
		eventDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{eventDescriptorItem})

	typeDescriptorItem := NewQStandardItemF("TypeDescriptor (%d entries)", len(MainLayer.TypeDescriptor))

	for _, item := range MainLayer.TypeDescriptor {
		nameItem := NewQStandardItemF(item.Name)
		idXItem := NewQStandardItemF("%d", item.IDx)
		unknownIntItem := NewQStandardItemF("%d", item.OtherID)
		typeDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{typeDescriptorItem})

	descriptorView.SetModel(standardModel)
	descriptorView.SetSelectionMode(0)
	descriptorView.SetSortingEnabled(true)

	dumpButton := widgets.NewQPushButton2("Dump...", nil)
	dumpButton.ConnectPressed(func() {
		location := widgets.QFileDialog_GetSaveFileName(dumpButton, "Save dictionaries...", "", "GOB files (*.gob)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		err = gob.NewEncoder(writer).Encode(MainLayer)
		if err != nil {
			println("while encoding:", err.Error())
		}
	})

	layerLayout.AddWidget(descriptorView, 0, 0)
	layerLayout.AddWidget(dumpButton, 0, 0)
}

*/
