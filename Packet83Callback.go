package main

import "strconv"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"
import "roblox-dissector/peer"
import "github.com/gskartwii/rbxfile"

var SubpacketCallbacks = map[uint8](func(peer.Packet83Subpacket) widgets.QWidget_ITF){
	0xB:  show83_0B,
	0x1:  show83_01,
	0x2:  show83_02,
	0x3:  show83_03,
	0x4:  show83_04,
	0x5:  show83_05,
	0x6:  show83_06,
	0x7:  show83_07,
	0x9:  show83_09,
	0xA:  show83_0A,
	0x10: show83_10,
	0x11: show83_11,
	0x12: show83_12,
}
var Callbacks83_09 = map[uint8](func(peer.Packet83_09Subpacket) widgets.QWidget_ITF){
	0x0: show83_09_00,
	0x1: show83_09_01,
	0x5: show83_09_05,
	0x6: show83_09_06,
	0x7: show83_09_07,
}

func showReplicationInstance(this *rbxfile.Instance) []*gui.QStandardItem {
	rootNameItem := NewQStandardItemF("Name: %s", this.Name())
	typeItem := NewQStandardItemF(this.ClassName)
	referentItem := NewQStandardItemF(this.Reference)
	var parentItem *gui.QStandardItem
	if this.Parent() != nil {
		parentItem = NewQStandardItemF(this.Parent().Reference)
	} else {
		parentItem = NewQStandardItemF("DataModel/NULL")
	}
	pathItem := NewQStandardItemF("%s", this.GetFullName())

	if len(this.Properties) > 0 {
		propertyRootItem := NewQStandardItemF("%d properties", len(this.Properties))
		for name, property := range this.Properties {
			nameItem := NewQStandardItemF(name)
			if property != nil {
				typeItem := NewQStandardItemF(property.Type().String())
				var valueItem *gui.QStandardItem
				if property.Type() == rbxfile.TypeProtectedString {
					valueItem = NewQStandardItemF("... (len %d)", len(property.String()))
				} else {
					valueItem = NewQStandardItemF(property.String())
				}

				propertyRootItem.AppendRow([]*gui.QStandardItem{
					nameItem,
					typeItem,
					valueItem,
					nil,
					nil,
					nil,
				})
			} else {
				propertyRootItem.AppendRow([]*gui.QStandardItem{
					nameItem,
					NewQStandardItemF("!!nil"),
					nil,
					nil,
					nil,
					nil,
				})
			}
		}
		rootNameItem.AppendRow([]*gui.QStandardItem{propertyRootItem, nil, nil, nil, nil, nil})
	}
	return []*gui.QStandardItem{
		rootNameItem,
		typeItem,
		nil,
		referentItem,
		parentItem,
		pathItem,
	}
}

type Packet83Subpacket peer.Packet83Subpacket

func show83_0B(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_0B)
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent", "Path"})

	rootNode := standardModel.InvisibleRootItem()
	if this != nil && this.Instances != nil { // if arraylen == 0, this is nil
		for _, instance := range this.Instances {
			rootNode.AppendRow(showReplicationInstance(instance))
		}
	}
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}
func show83_01(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_01)
	return NewQLabelF("Delete instance: %s, %s", this.Instance.Reference, this.Instance.GetFullName())
}
func show83_02(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_02)
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent", "Path"})

	rootNode := standardModel.InvisibleRootItem()
	rootNode.AppendRow(showReplicationInstance(this.Child))
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}
func show83_03(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_03)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	if this.Instance != nil {
		layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	} else {
		layout.AddWidget(NewQLabelF("Object: nil"), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown int: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Property name: %s", this.PropertyName), 0, 0)
	layout.AddWidget(NewQLabelF("Property type: %s", this.Value.Type().String()), 0, 0)
	if this.Value.Type() == rbxfile.TypeProtectedString {
		layout.AddWidget(NewQLabelF("Property value: ... (len %d)", len(this.Value.String())), 0, 0)
	} else {
		layout.AddWidget(NewQLabelF("Property value: %s", this.Value.String()), 0, 0)
	}
	widget.SetLayout(layout)

	return widget
}
func show83_0A(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_0A)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	if this.Instance != nil {
		layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	} else {
		layout.AddWidget(NewQLabelF("Object: nil"), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Acked property name: %s", this.PropertyName), 0, 0)
	layout.AddWidget(NewQLabelF("Property version 1: %d", this.Versions[0]), 0, 0) // TODO
	widget.SetLayout(layout)

	return widget
}
func show83_04(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_04)
	return NewQLabelF("Marker: %d", this.MarkerId)
}
func show83_05(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_05)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Is ping back: %v", this.IsPingBack), 0, 0)
	layout.AddWidget(NewQLabelF("Timestamp: %d", this.Timestamp), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 1: %d", this.SendStats), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 2: %d", this.ExtraStats), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_06(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_06)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Is ping back: %v", this.IsPingBack), 0, 0)
	layout.AddWidget(NewQLabelF("Timestamp: %d", this.Timestamp), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 1: %d", this.SendStats), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 2: %d", this.ExtraStats), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_07(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_07)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	if this.Instance != nil {
		layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	} else {
		layout.AddWidget(NewQLabelF("Object: nil"), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Event name: %s", this.EventName), 0, 0)
	layout.AddWidget(NewQLabelF("Arguments:"), 0, 0)

	argumentList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Type", "Value"})
	rootNode := standardModel.InvisibleRootItem()

	for _, argument := range this.Event.Arguments {
		if argument != nil {
			rootNode.AppendRow([]*gui.QStandardItem{
				NewQStandardItemF(argument.Type().String()),
				NewQStandardItemF("%s", argument.String()),
			})
		} else {
			rootNode.AppendRow([]*gui.QStandardItem{
				NewQStandardItemF("nil"),
				nil,
			})
		}
	}

	argumentList.SetModel(standardModel)
	argumentList.SetSelectionMode(0)
	argumentList.SetSortingEnabled(true)
	layout.AddWidget(argumentList, 0, 0)
	widget.SetLayout(layout)

	return widget
}

func show83_09_00(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_00)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()

	for i := 0; i < 5; i++ {
		layout.AddWidget(NewQLabelF("Int %d: %d", i, this.Values[i]), 0, 0)
	}
	widget.SetLayout(layout)

	return widget
}
func show83_09_01(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_01)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	layout.AddWidget(NewQLabelF("Int 3: %d", this.Int3), 0, 0)
	layout.AddWidget(NewQLabelF("Int 4: %d", this.Int4), 0, 0)
	layout.AddWidget(NewQLabelF("Int 5: %d", this.Int5), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09_05(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_05)
	return NewQLabelF("Int: %d", this.Int)
}
func show83_09_06(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_06)

	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09_07(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	return NewQLabelF("(no values)")
}
func show83_09(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Type: %d", this.Type), 0, 0)

	callback := Callbacks83_09[this.Type]
	if callback == nil {
		println("unsupported callback")
		widget.SetLayout(layout)
		return widget
	}

	layout.AddWidget(callback(this.Subpacket), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_10(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_10)
	return NewQLabelF("Replication tag: %d", this.TagId)
}
func show83_11(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_11)
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Skip stat set 1: %v", this.SkipStats1), 0, 0)
	if !this.SkipStats1 {
		layout.AddWidget(NewQLabelF("Stat 1/1: %s", this.Stats_1_1), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/2: %G", this.Stats_1_2), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/3: %G", this.Stats_1_3), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/4: %G", this.Stats_1_4), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/5: %v", this.Stats_1_5), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Skip stat set 2: %v", this.SkipStats2), 0, 0)
	if !this.SkipStats2 {
		layout.AddWidget(NewQLabelF("Stat 2/1: %s", this.Stats_2_1), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/2: %G", this.Stats_2_2), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/3: %d", this.Stats_2_3), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/4: %v", this.Stats_2_4), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Average ping ms: %G", this.AvgPingMs), 0, 0)
	layout.AddWidget(NewQLabelF("Average physics sender Pkt/s: %G", this.AvgPhysicsSenderPktPS), 0, 0)
	layout.AddWidget(NewQLabelF("Total data KB/s: %G", this.TotalDataKBPS), 0, 0)
	layout.AddWidget(NewQLabelF("Total physics KB/s: %G", this.TotalPhysicsKBPS), 0, 0)
	layout.AddWidget(NewQLabelF("Data throughput ratio: %G", this.DataThroughputRatio), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_12(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_12)
	widget := widgets.NewQWidget(nil, 0)
	layerLayout := widgets.NewQVBoxLayout()
	hashListLabel := NewQLabelF("Hashes:")
	layerLayout.AddWidget(hashListLabel, 0, 0)

	hashList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Hash"})
	rootItem := standardModel.InvisibleRootItem()
	for index, hash := range this.HashList {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("%d", index),
			NewQStandardItemF("%8X", hash),
		})
	}
	for index, hash := range this.SecurityTokens {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("ST%d", index),
			NewQStandardItemF("%16X", hash),
		})
	}
	hashList.SetSortingEnabled(true)
	hashList.SetModel(standardModel)
	layerLayout.AddWidget(hashList, 0, 0)

	return widget
}

func showPacket83Subpacket(this Packet83Subpacket) widgets.QWidget_ITF {
	return SubpacketCallbacks[peer.Packet83ToType(this)](this)
}

func ShowPacket83(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet83Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	packetListLabel := NewQLabelF("Replication subpackets:")
	layerLayout.AddWidget(packetListLabel, 0, 0)

	packetList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type"})
	rootItem := standardModel.InvisibleRootItem()
	for index, subpacket := range MainLayer.SubPackets {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("%d", index),
			NewQStandardItemF(peer.Packet83ToTypeString(subpacket)),
		})
	}
	packetList.SetSelectionMode(1)
	packetList.SetSortingEnabled(true)
	packetList.SetModel(standardModel)
	packetList.ConnectClicked(func(index *core.QModelIndex) {
		thisIndex, _ := strconv.Atoi(standardModel.Item(index.Row(), 0).Data(0).ToString())
		subpacket := MainLayer.SubPackets[thisIndex]

		subWindow := widgets.NewQWidget(packetList, core.Qt__Window)
		subWindowLayout := widgets.NewQVBoxLayout2(subWindow)

		isClient := context.IsClient(packet.Source)
		isServer := context.IsServer(packet.Source)

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

		showCallback, ok := SubpacketCallbacks[peer.Packet83ToType(subpacket)]
		if !ok {
			println("unsupported type:", peer.Packet83ToType(subpacket))
		} else {
			subWindowLayout.AddWidget(showCallback(subpacket), 0, 0)
		}

		subWindow.SetWindowTitle("Replication Packet Window: " + peer.Packet83ToTypeString(subpacket))
		subWindow.Show()
	})
	layerLayout.AddWidget(packetList, 0, 0)
}
