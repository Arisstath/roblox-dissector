package main
import "github.com/google/gopacket"
import "strconv"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"

func ShowPacket83(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet83Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	packetListLabel := NewQLabelF("Replication subpackets:")
	layerLayout.AddWidget(packetListLabel, 0, 0)

	packetList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type"})
	rootItem := standardModel.InvisibleRootItem()
	for index, subpacket := range(MainLayer.SubPackets) {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("%d", index),
			NewQStandardItemF(subpacket.TypeString()),
		})
	}
	packetList.SetSelectionMode(1)
	packetList.SetSortingEnabled(true)
	packetList.SetModel(standardModel)
	packetList.ConnectSelectionChanged(func (selected *core.QItemSelection, deselected *core.QItemSelection) {
		if len(selected.Indexes()) != 0 {
			index, _ := strconv.Atoi(standardModel.Item(selected.Indexes()[0].Row(), 0).Data(0).ToString())
			subpacket := MainLayer.SubPackets[index]

			subWindow := widgets.NewQWidget(packetList, core.Qt__Window)
			subWindowLayout := widgets.NewQVBoxLayout2(subWindow)

			isClient := context.PacketFromClient(packet)
			isServer := context.PacketFromServer(packet)

			var direction string
			if isClient {
				direction = "Direction: Client -> Server"
			} else if isServer {
				direction = "Direction: Server -> Client"
			} else {
				direction = "Direction: Unknown"
			}
			directionLabel := widgets.NewQLabel2(direction, nil, 0)
			subWindowLayout.AddWidget(directionLabel, 0, 0)
			subWindowLayout.AddWidget(subpacket.Show(), 0, 0)
			subWindow.SetWindowTitle("Replication Packet Window: " + subpacket.TypeString())
			subWindow.Show()
		}
	})
	layerLayout.AddWidget(packetList, 0, 0)
}
