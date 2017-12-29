package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket90(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet90Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Schema version: %d", MainLayer.SchemaVersion), 0, 0)
	
	requestList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Flag name"})

	requestListRootNode := standardModel.InvisibleRootItem()
	for i := 0; i < len(MainLayer.RequestedFlags); i++ {
		requestListRootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("%s", MainLayer.RequestedFlags[i])})
	}

	requestList.SetModel(standardModel)
	requestList.SetSelectionMode(0)
	requestList.SetSortingEnabled(true)

	layerLayout.AddWidget(NewQLabelF("Requested flags:"), 0, 0)
	layerLayout.AddWidget(requestList, 0, 0)
}
