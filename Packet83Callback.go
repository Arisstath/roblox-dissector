package main
import "github.com/google/gopacket"
import "strconv"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

var SubpacketCallbacks = map[uint8](func() widgets.QWidget_ITF){
	0xB: show83_0B,
	0x1: show83_01,
	0x2: show83_02,
	0x3: show83_03,
	0x4: show83_04,
	0x5: show83_05,
	0x7: show83_07,
	0x9: show83_09,
	0x10: show83_10,
	0x11: show83_11,
}
var Callbacks83_09 = map[uint8](func() widgets.QWidget_ITF){
	0x1: show83_09_01,
	0x5: show83_09_05,
	0x7: show83_09_07,
	0x9: show83_09_09,
}

func showReplicationInstance(this *peer.ReplicationInstance) []*gui.QStandardItem {
	rootNameItem := NewQStandardItemF(this.findName())
	typeItem := NewQStandardItemF(this.ClassName)
	referentItem := NewQStandardItemF(this.Object1.Show())
	unknownBoolItem := NewQStandardItemF("%v", this.Bool1)
	parentItem := NewQStandardItemF(this.Object2.Show())

	for _, property := range this.Properties {
		nameItem := NewQStandardItemF(property.Name)
		typeItem := NewQStandardItemF(property.Type)
		valueItem := NewQStandardItemF(property.Show())

		rootNameItem.AppendRow([]*gui.QStandardItem{
			nameItem,
			typeItem,
			valueItem,
			nil,
			nil,
			nil,
		})
	}

	return []*gui.QStandardItem{
		rootNameItem,
		typeItem,
		nil,
		referentItem,
		unknownBoolItem,
		parentItem,
	}
}


type Packet83Subpacket peer.Packet83Subpacket
func show83_0B() widgets.QWidget_ITF {
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Unknown bool", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	for _, instance := range(this.Instances) {
		rootNode.AppendRow(showReplicationInstance(instance))
	}
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}
func show83_01() widgets.QWidget_ITF {
	return NewQLabelF("Init referent: %s", this.Object1.Show())
}
func show83_02() widgets.QWidget_ITF {
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Unknown bool", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	rootNode.AppendRow(showReplicationInstance(this.Child))
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}
func show83_03() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Referent: %s", this.Object1.Show()), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Property name: %s", this.PropertyName), 0, 0)
	layout.AddWidget(NewQLabelF("Property type: %s", this.Value.Type), 0, 0)
	layout.AddWidget(NewQLabelF("Property value: %s", this.Value.Show()), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_04() widgets.QWidget_ITF {
	return NewQLabelF("Marker: %d", this.MarkerId)
}
func show83_05() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	layout.AddWidget(NewQLabelF("Int 3: %d", this.Int3), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func packet83_07() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Referent: %s", this.Object1.Show()), 0, 0)
	layout.AddWidget(NewQLabelF("Event name: %s", this.EventName), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown int: %d", this.Event.UnknownInt), 0, 0)
	layout.AddWidget(NewQLabelF("Arguments:"), 0, 0)

	argumentList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Type", "Value"})
	rootNode := standardModel.InvisibleRootItem()

	for _, argument := range this.Event.Arguments {
		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF(argument.Type),
			NewQStandardItemF("%s", argument.Value.Show()),
		})
	}

	argumentList.SetModel(standardModel)
	argumentList.SetSelectionMode(0)
	argumentList.SetSortingEnabled(true)
	layout.AddWidget(argumentList, 0, 0)
	widget.SetLayout(layout)
	
	return widget
}

func show83_09_01() widgets.QWidget_ITF {
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
func show83_09_05() widgets.QWidget_ITF {
	return NewQLabelF("Int: %d", this.Int)
}
func show83_09_07() widgets.QWidget_ITF {
	return NewQLabelF("(no values)")
}
func show83_09_09() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Type: %d", this.Type), 0, 0)
	layout.AddWidget(Callbacks83_09[this.Type](), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_10() widgets.QWidget_ITF {
	return NewQLabelF("Replication tag: %d", this.TagId)
}
func show83_11() widgets.QWidget_ITF {
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

func (this Packet83Subpacket) Show() widgets.QWidget_ITF {
	return SubpacketCallbacks[this.Type()](this)
}


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
	packetList.ConnectClicked(func (index *core.QModelIndex) {
		thisIndex, _ := strconv.Atoi(standardModel.Item(index.Row(), 0).Data(0).ToString())
		subpacket := MainLayer.SubPackets[thisIndex]

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
	})
	layerLayout.AddWidget(packetList, 0, 0)
}
