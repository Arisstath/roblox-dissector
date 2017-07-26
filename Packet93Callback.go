package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "strconv"

func ShowPacket93(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet93Layer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)

	unknownBool1Label := NewQLabelF("Unknown bool 1: %v", MainLayer.UnknownBool1)
	unknownBool2Label := NewQLabelF("Unknown bool 2: %v", MainLayer.UnknownBool2)
	layerLayout.AddWidget(unknownBool1Label, 0, 0)
	layerLayout.AddWidget(unknownBool2Label, 0, 0)

	labelForParamList := NewQLabelF("Network params:")
	layerLayout.AddWidget(labelForParamList, 0, 0)

	paramList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Value"})

	paramListRootNode := standardModel.InvisibleRootItem()
	for Name, Value := range MainLayer.Params {
		nameItem := gui.NewQStandardItem2(Name)
		valueItem := gui.NewQStandardItem2(strconv.FormatBool(Value))
		nameItem.SetEditable(false)
		valueItem.SetEditable(false)
		paramListRootNode.AppendRow([]*gui.QStandardItem{nameItem, valueItem})
	}

	paramList.SetModel(standardModel)
	paramList.SetSelectionMode(0)

	layerLayout.AddWidget(paramList, 0, 0)
}
