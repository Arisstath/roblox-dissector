package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

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
		nameItem := NewQStandardItemF(item.Instance.GetFullName())
		referenceItem := NewQStandardItemF(item.Instance.Reference)
		humanoidStateItem := NewQStandardItemF("%d", item.NetworkHumanoidState)
		cframeItem := NewQStandardItemF(item.CFrame.String())
		linVelItem := NewQStandardItemF(item.LinearVelocity.String())
		rotVelItem := NewQStandardItemF(item.RotationalVelocity.String())

		if len(item.Motors) > 0 {
			motorsItem := NewQStandardItemF("Motors (%d entries)", len(item.Motors))
			for _, motor := range item.Motors {
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
		if len(item.HistoryWaypoints) > 0 {
			waypointsItem := NewQStandardItemF("Waypoints (%d entries)", len(item.HistoryWaypoints))
			for _, waypoint := range item.HistoryWaypoints {
				waypointsItem.AppendRow([]*gui.QStandardItem{
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					NewQStandardItemF(waypoint.Position.String()),
					NewQStandardItemF("%d", waypoint.PrecisionLevel),
					NewQStandardItemF("%d", waypoint.Interval),
				})
			}
			nameItem.AppendRow([]*gui.QStandardItem{waypointsItem})
		}
		if len(item.Children) > 0 {
			childrenItem := NewQStandardItemF("Children (%d entries)", len(item.Children))
			for _, child := range item.Children {
				childrenItem.AppendRow([]*gui.QStandardItem{
					NewQStandardItemF(child.Instance.Name()),
					NewQStandardItemF(child.Instance.Reference),
					NewQStandardItemF("%d", child.NetworkHumanoidState),
					NewQStandardItemF(child.CFrame.String()),
					NewQStandardItemF(child.LinearVelocity.String()),
					NewQStandardItemF(child.RotationalVelocity.String()),
				})
				if len(child.Motors) > 0 {
					motorsItem := NewQStandardItemF("Motors (%d entries)", len(item.Motors))
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
