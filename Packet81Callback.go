package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

func ShowPacket81(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet81Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Distributed physics enabled: %v", MainLayer.DistributedPhysicsEnabled), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Stream job: %v", MainLayer.StreamJob), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Filtering enabled: %v", MainLayer.FilteringEnabled), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Allow third party sales: %v", MainLayer.AllowThirdPartySales), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Character auto spawn: %v", MainLayer.CharacterAutoSpawn), 0, 0)
	referentStringLabel := NewQLabelF("Top replication scope: %s", MainLayer.ReferentString)
	layerLayout.AddWidget(referentStringLabel, 0, 0)
	layerLayout.AddWidget(NewQLabelF("Int 1: %X", MainLayer.Int1), 0, 0)
	layerLayout.AddWidget(NewQLabelF("Int 2: %X", MainLayer.Int2), 0, 0)

	deletedList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Class name", "Referent"})

	deletedListRootNode := standardModel.InvisibleRootItem()
	for i := 0; i < len(MainLayer.Items); i++ {
		classNameItem := NewQStandardItemF("%s", context.StaticSchema.Instances[MainLayer.Items[i].ClassID].Name)
		referenceItem := NewQStandardItemF("%s", MainLayer.Items[i].Instance.Reference)
		deletedListRootNode.AppendRow([]*gui.QStandardItem{classNameItem, referenceItem})
	}

	deletedList.SetModel(standardModel)
	deletedList.SetSelectionMode(0)
	deletedList.SetSortingEnabled(true)

	layerLayout.AddWidget(deletedList, 0, 0)
}
