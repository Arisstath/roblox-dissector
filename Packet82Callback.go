package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/widgets"
import "fmt"
import "strconv"

func ShowPacket82(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet82Layer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)

	labelForDescriptorView := NewQLabelF("Dictionaries:")
	layerLayout.AddWidget(labelForDescriptorView, 0, 0)

	descriptorView := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "IDx", "Unknown int"})

	dictionaryRootNode := standardModel.InvisibleRootItem()
	classDescriptorItem := gui.NewQStandardItem2(fmt.Sprintf("ClassDescriptor (%d entries)", len(MainLayer.ClassDescriptor)))

	for _, item := range MainLayer.ClassDescriptor {
		nameItem := gui.NewQStandardItem2(item.Name)
		idXItem := gui.NewQStandardItem2(strconv.Itoa(int(item.IDx)))
		unknownIntItem := gui.NewQStandardItem2(strconv.Itoa(int(item.OtherID)))
		classDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{classDescriptorItem})

	propertyDescriptorItem := gui.NewQStandardItem2(fmt.Sprintf("PropertyDescriptor (%d entries)", len(MainLayer.PropertyDescriptor)))

	for _, item := range MainLayer.PropertyDescriptor {
		nameItem := gui.NewQStandardItem2(item.Name)
		idXItem := gui.NewQStandardItem2(strconv.Itoa(int(item.IDx)))
		unknownIntItem := gui.NewQStandardItem2(strconv.Itoa(int(item.OtherID)))
		propertyDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{propertyDescriptorItem})

	eventDescriptorItem := gui.NewQStandardItem2(fmt.Sprintf("EventDescriptor (%d entries)", len(MainLayer.EventDescriptor)))

	for _, item := range MainLayer.EventDescriptor {
		nameItem := gui.NewQStandardItem2(item.Name)
		idXItem := gui.NewQStandardItem2(strconv.Itoa(int(item.IDx)))
		unknownIntItem := gui.NewQStandardItem2(strconv.Itoa(int(item.OtherID)))
		eventDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{eventDescriptorItem})

	typeDescriptorItem := gui.NewQStandardItem2(fmt.Sprintf("TypeDescriptor (%d entries)", len(MainLayer.TypeDescriptor)))

	for _, item := range MainLayer.TypeDescriptor {
		nameItem := gui.NewQStandardItem2(item.Name)
		idXItem := gui.NewQStandardItem2(strconv.Itoa(int(item.IDx)))
		unknownIntItem := gui.NewQStandardItem2(strconv.Itoa(int(item.OtherID)))
		typeDescriptorItem.AppendRow([]*gui.QStandardItem{nameItem, idXItem, unknownIntItem})
	}
	dictionaryRootNode.AppendRow([]*gui.QStandardItem{typeDescriptorItem})

	descriptorView.SetModel(standardModel)
	descriptorView.SetSelectionMode(0)
	descriptorView.SetSortingEnabled(true)

	layerLayout.AddWidget(descriptorView, 0, 0)
}
