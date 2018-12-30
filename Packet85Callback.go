package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/Gskartwii/roblox-dissector/peer"

var NetworkHumanoidStates = [...]string{
	"Falling down",
	"Ragdoll",
	"Getting up",
	"Jumping",
	"Swimming",
	"Freefall",
	"Flying",
	"Landed",
	"Running",
	"Unknown 9",
	"Running, no physics",
	"Strafing, no physics",
	"Climbing",
	"Seated",
	"Standing on a platform",
	"Dead",
	"Pure physics",
	"Unknown 17",
	"None",
}

func ShowPacket85(packetType byte, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet85Layer)

	layerLayout := NewBasicPacketViewer(packetType, context, layers)

	labelForSubpackets := NewQLabelF("Physics replication (%d entries):", len(MainLayer.SubPackets))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	subpacketList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Reference", "Humanoid state", "CFrame", "Linear velocity", "Rotational velocity", "Position", "Precision", "Interval"})

	rootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.SubPackets {
		if item.Data.Instance == nil {
			rootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("nil!!")})
			continue
		}
		nameItem := NewQStandardItemF(item.Data.Instance.GetFullName())
		referenceItem := NewQStandardItemF(item.Data.Instance.Ref.String())
		humanoidStateItem := NewQStandardItemF("%s", NetworkHumanoidStates[item.NetworkHumanoidState])
		cframeItem := NewQStandardItemF(item.Data.CFrame.String())
		linVelItem := NewQStandardItemF(item.Data.LinearVelocity.String())
		rotVelItem := NewQStandardItemF(item.Data.RotationalVelocity.String())

		if item.Data.PlatformChild != nil {
			nameItem.AppendRow([]*gui.QStandardItem{
				NewQStandardItemF(item.Data.PlatformChild.GetFullName()),
				NewQStandardItemF(item.Data.PlatformChild.Ref.String()),
			})
		}
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
					nil,
					nil,
					NewQStandardItemF("%f", history.Interval),
				})
				if history.PlatformChild != nil {
					historyItem.AppendRow([]*gui.QStandardItem{
						NewQStandardItemF(history.PlatformChild.GetFullName()),
						NewQStandardItemF(history.PlatformChild.Ref.String()),
					})
				}
			}
			nameItem.AppendRow([]*gui.QStandardItem{historyItem})
		}
		if len(item.Children) > 0 {
			childrenItem := NewQStandardItemF("Children (%d entries)", len(item.Children))
			for _, child := range item.Children {
				if child.Instance == nil {
					println("WARNING: can't display nonexistent child!")
					continue
				}
				childItem := NewQStandardItemF(child.Instance.Name())
				childrenItem.AppendRow([]*gui.QStandardItem{
					childItem,
					NewQStandardItemF(child.Instance.Ref.String()),
					nil,
					NewQStandardItemF(child.CFrame.String()),
					NewQStandardItemF(child.LinearVelocity.String()),
					NewQStandardItemF(child.RotationalVelocity.String()),
				})
				if child.PlatformChild != nil {
					childItem.AppendRow([]*gui.QStandardItem{
						NewQStandardItemF(child.PlatformChild.GetFullName()),
						NewQStandardItemF(child.PlatformChild.Ref.String()),
					})
				}
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
					childItem.AppendRow([]*gui.QStandardItem{motorsItem})
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
