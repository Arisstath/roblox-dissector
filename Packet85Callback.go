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
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Reference", "Unknown int", "CFrame", "Vector3 1", "Vector3 2", "Angle"})

	rootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.SubPackets {
		nameItem := NewQStandardItemF(item.Instance.GetFullName())
		referenceItem := NewQStandardItemF(item.Instance.Reference)
		unknownIntItem := NewQStandardItemF("%d", item.UnknownInt)
		cframeItem := NewQStandardItemF(item.CFrame.String())
		pos1Item := NewQStandardItemF(item.Pos1.String())
		pos2Item := NewQStandardItemF(item.Pos2.String())

		if len(item.Motors) > 0 {
			motorsItem := NewQStandardItemF("Motors (%d entries)", len(item.Motors))
			for _, motor := range item.Motors {
				motorsItem.AppendRow([]*gui.QStandardItem{
					nil,
					nil,
					nil,
					nil,
					NewQStandardItemF(motor.Coords1.String()),
					NewQStandardItemF(motor.Coords2.String()),
					NewQStandardItemF("%d", motor.Angle),
				})
			}
			nameItem.AppendRow([]*gui.QStandardItem{motorsItem})
		}
		if len(item.Children) > 0 {
			childrenItem := NewQStandardItemF("Children (%d entries)", len(item.Children))
			for _, child := range item.Children {
				childrenItem.AppendRow([]*gui.QStandardItem{
					NewQStandardItemF(child.Instance.Name()),
					NewQStandardItemF(child.Instance.Reference),
					NewQStandardItemF("%d", child.UnknownInt),
					NewQStandardItemF(child.CFrame.String()),
					NewQStandardItemF(child.Pos1.String()),
					NewQStandardItemF(child.Pos2.String()),
					nil,
				})
				if len(child.Motors) > 0 {
					motorsItem := NewQStandardItemF("Motors (%d entries)", len(item.Motors))
					for _, motor := range child.Motors {
						motorsItem.AppendRow([]*gui.QStandardItem{
							nil,
							nil,
							nil,
							nil,
							NewQStandardItemF(motor.Coords1.String()),
							NewQStandardItemF(motor.Coords2.String()),
							NewQStandardItemF("%d", motor.Angle),
						})
					}
					nameItem.AppendRow([]*gui.QStandardItem{motorsItem})
				}
			}

			nameItem.AppendRow([]*gui.QStandardItem{childrenItem})
		}

		rootNode.AppendRow([]*gui.QStandardItem{nameItem, referenceItem, unknownIntItem, cframeItem, pos1Item, pos2Item, nil})
	}

	subpacketList.SetModel(standardModel)
	subpacketList.SetSelectionMode(0)
	subpacketList.SetSortingEnabled(true)
	layerLayout.AddWidget(subpacketList, 0, 0)
}
