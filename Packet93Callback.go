package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "roblox-dissector/peer"

func ShowPacket93(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet93Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	unknownBool1Label := NewQLabelF("Protocol schema sync: %v", MainLayer.ProtocolSchemaSync)
	unknownBool2Label := NewQLabelF("API dictionary compression: %v", MainLayer.ApiDictionaryCompression)
	layerLayout.AddWidget(unknownBool1Label, 0, 0)
	layerLayout.AddWidget(unknownBool2Label, 0, 0)

	labelForParamList := NewQLabelF("Network params:")
	layerLayout.AddWidget(labelForParamList, 0, 0)

	paramList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Value"})

	paramListRootNode := standardModel.InvisibleRootItem()
	for Name, Value := range MainLayer.Params {
		nameItem := NewQStandardItemF(Name)
		valueItem := NewQStandardItemF("%v", Value)
		paramListRootNode.AppendRow([]*gui.QStandardItem{nameItem, valueItem})
	}

	paramList.SetModel(standardModel)
	paramList.SetSelectionMode(0)

	layerLayout.AddWidget(paramList, 0, 0)
}
