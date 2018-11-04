package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "roblox-dissector/peer"

func ShowPacket85(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet85Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	labelForSubpackets := NewQLabelF("Physics replication (%d entries):", len(MainLayer.SubPackets))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	subpacketList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Reference", "Humanoid state", "CFrame", "Linear velocity", "Rotational velocity", "Position", "Precision", "Interval"})

	rootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.SubPackets {
		nameItem := NewQStandardItemF(item.Data.Instance.GetFullName())
		referenceItem := NewQStandardItemF(item.Data.Instance.Reference)
		humanoidStateItem := NewQStandardItemF("%d", item.NetworkHumanoidState)
		cframeItem := NewQStandardItemF(item.Data.CFrame.String())
		linVelItem := NewQStandardItemF(item.Data.LinearVelocity.String())
		rotVelItem := NewQStandardItemF(item.Data.RotationalVelocity.String())

		if len(item.Data.Motors) > 0 {
			motorsItem := NewQStandardItemF("Motors (%d entries)", len(item.Data.Motors))
			for _, motor := range item.Data.Motors {
				motorsItem.AppendRow([]*gui.QStandardItem{
					nil,
					nil,
					nil,
					NewQStandardItemF(motor.String()),
					nil,
					nil,
					nil,
					nil,
					nil,
				})
			}
			nameItem.AppendRow([]*gui.QStandardItem{motorsItem})
		}
		if len(item.History) > 0 {
			historyItem := NewQStandardItemF("History (%d entries)", len(item.History))
			for _, history := range item.History {
				historyItem.AppendRow([]*gui.QStandardItem{
					nil,
					nil,
					nil,
					NewQStandardItemF(history.CFrame.String()),
					NewQStandardItemF(history.LinearVelocity.String()),
					NewQStandardItemF(history.RotationalVelocity.String()),
				})
			}
			nameItem.AppendRow([]*gui.QStandardItem{historyItem})
		}
		if len(item.Children) > 0 {
			childrenItem := NewQStandardItemF("Children (%d entries)", len(item.Children))
			for _, child := range item.Children {
				childrenItem.AppendRow([]*gui.QStandardItem{
					NewQStandardItemF(child.Instance.Name()),
					NewQStandardItemF(child.Instance.Reference),
					nil,
					NewQStandardItemF(child.CFrame.String()),
					NewQStandardItemF(child.LinearVelocity.String()),
					NewQStandardItemF(child.RotationalVelocity.String()),
				})
				if len(child.Motors) > 0 {
					motorsItem := NewQStandardItemF("Motors (%d entries)", len(child.Motors))
					for _, motor := range child.Motors {
						motorsItem.AppendRow([]*gui.QStandardItem{
							nil,
							nil,
							nil,
							NewQStandardItemF(motor.String()),
							nil,
							nil,
							nil,
							nil,
							nil,
						})
					}
					nameItem.AppendRow([]*gui.QStandardItem{motorsItem})
				}
			}

			nameItem.AppendRow([]*gui.QStandardItem{childrenItem})
		}

		rootNode.AppendRow([]*gui.QStandardItem{nameItem, referenceItem, humanoidStateItem, cframeItem, linVelItem, rotVelItem})
	}

	subpacketList.SetModel(standardModel)
	subpacketList.SetSelectionMode(0)
	subpacketList.SetSortingEnabled(true)
	layerLayout.AddWidget(subpacketList, 0, 0)
}
