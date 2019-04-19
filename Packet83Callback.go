package main

import (
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"github.com/dustin/go-humanize"
)

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
	0x13: show83_13,
}
var Callbacks83_09 = map[uint8](func(peer.Packet83_09Subpacket) widgets.QWidget_ITF){
	0x0: show83_09_00,
	0x4: show83_09_04,
	0x5: show83_09_05,
	0x6: show83_09_06,
}

// Allow caller to manually pass Parent.
// This is because the Parent() stored in datamodel.Instance isn't necessarily
// the Parent we want.
func showReplicationInstance(this *datamodel.Instance, parent *datamodel.Instance) []*gui.QStandardItem {
	rootNameItem := NewStringItem(this.Name())
	typeItem := NewStringItem(this.ClassName)
	referenceItem := NewStringItem(this.Ref.String())
	var parentItem *gui.QStandardItem
	if parent != nil {
		parentItem = NewStringItem(parent.Ref.String())
	} else {
		parentItem = NewStringItem("DataModel/NULL")
	}
	pathItem := NewStringItem(this.GetFullName())

	return []*gui.QStandardItem{
		rootNameItem,
		typeItem,
		nil,
		referenceItem,
		parentItem,
		pathItem,
	}
}

// Remember to lock the properties mutex!
// TODO: remove numNils hack
func showProperties(properties map[string]rbxfile.Value, numNils int) *gui.QStandardItem {
	propertyRootItem := NewQStandardItemF("%d properties", len(properties))
	for name, property := range properties {
		nameItem := NewStringItem(name)
		if property != nil {
			typeItem := NewStringItem(datamodel.TypeString(property))
			var valueItem *gui.QStandardItem
			if property.Type() == rbxfile.TypeProtectedString {
				valueItem = NewQStandardItemF("... (len %d)", len(property.String()))
			} else {
				valueItem = NewStringItem(property.String())
			}

			baseRow := make([]*gui.QStandardItem, numNils)
			baseRow = append(baseRow, nameItem, typeItem, valueItem)
			propertyRootItem.AppendRow(baseRow)
		} else {
			baseRow := make([]*gui.QStandardItem, numNils)
			baseRow = append(baseRow, nameItem, NewStringItem("nil"))
			propertyRootItem.AppendRow(baseRow)
		}
	}
	return propertyRootItem
}

type Packet83Subpacket peer.Packet83Subpacket

func show83_0B(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_0B)
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Name", "Type", "Value", "Reference", "Parent", "Path"})

	rootNode := standardModel.InvisibleRootItem()
	if this != nil && this.Instances != nil { // if arraylen == 0, this is nil
		for i, instance := range this.Instances {
			indexItem := NewUintItem(i)
			row := []*gui.QStandardItem{indexItem}
			row = append(row, showReplicationInstance(instance.Instance, instance.Parent)...)
			indexItem.AppendRow([]*gui.QStandardItem{showProperties(instance.Properties, 1)})
			rootNode.AppendRow(row)
		}
	}
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)
	instanceList.SortByColumn(0, core.Qt__AscendingOrder)

	return instanceList
}
func show83_01(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_01)
	return NewQLabelF("Delete instance: %s, %s", this.Instance.Ref.String(), this.Instance.GetFullName())
}

// TODO: Properties may be changed later by 83_03?
func show83_02(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_02)
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Reference", "Parent", "Path"})

	rootNode := standardModel.InvisibleRootItem()
	row := showReplicationInstance(this.Instance, this.Parent)
	row[0].AppendRow([]*gui.QStandardItem{showProperties(this.Properties, 0)})
	rootNode.AppendRow(row)
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}
func show83_03(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_03)
	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	layout.AddWidget(NewQLabelF("Reference: %s", this.Instance.Ref.String()), 0, 0)
	layout.AddWidget(NewQLabelF("Has version: %v", this.HasVersion), 0, 0)
	if this.HasVersion {
		layout.AddWidget(NewQLabelF("Version: %d", this.Version), 0, 0)
	}
	if this.Schema == nil {
		layout.AddWidget(NewLabel("Property name: Parent"), 0, 0)
		layout.AddWidget(NewLabel("Property type: Reference"), 0, 0)
	} else {
		layout.AddWidget(NewQLabelF("Property name: %s", this.Schema.Name), 0, 0)
		layout.AddWidget(NewQLabelF("Property type: %s", this.Schema.TypeString), 0, 0)
	}
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
	layout := NewTopAlignLayout()
	if this.Instance != nil {
		layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	} else {
		layout.AddWidget(NewLabel("Object: nil"), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Acked property name: %s", this.Schema.Name), 0, 0)
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
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Packet version: %d", this.PacketVersion), 0, 0)
	layout.AddWidget(NewQLabelF("Timestamp: %d", this.Timestamp), 0, 0)
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Fps: %f, %f, %f", this.Fps1, this.Fps2, this.Fps3), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 1: %d", this.SendStats), 0, 0)
	layout.AddWidget(NewQLabelF("Stats 2: %d", this.ExtraStats), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_06(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_06)
	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
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
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Object: %s", this.Instance.GetFullName()), 0, 0)
	layout.AddWidget(NewQLabelF("Reference: %s", this.Instance.Ref.String()), 0, 0)
	layout.AddWidget(NewQLabelF("Event name: %s", this.Schema.Name), 0, 0)
	layout.AddWidget(NewLabel("Arguments:"), 0, 0)

	argumentList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Value"})
	rootNode := standardModel.InvisibleRootItem()

	for i, argument := range this.Event.Arguments {
		if argument != nil {
			rootNode.AppendRow([]*gui.QStandardItem{
				NewUintItem(i),
				NewStringItem(datamodel.TypeString(argument)),
				NewStringItem(argument.String()),
			})
		} else {
			rootNode.AppendRow([]*gui.QStandardItem{
				NewUintItem(i),
				NewStringItem("nil"),
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
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	layout.AddWidget(NewQLabelF("Int 3: %d", this.Int3), 0, 0)
	layout.AddWidget(NewQLabelF("Int 4: %d", this.Int4), 0, 0)
	layout.AddWidget(NewQLabelF("Int 5: %d", this.Int5), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09_04(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_04)

	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09_05(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_05)
	return NewQLabelF("Id challenge: %d", this.Challenge)
}
func show83_09_06(t peer.Packet83_09Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09_06)

	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Id challenge: %d", this.Challenge), 0, 0)
	layout.AddWidget(NewQLabelF("Response: %d", this.Response), 0, 0)
	widget.SetLayout(layout)

	return widget
}
func show83_09(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_09)
	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Type: %d", this.SubpacketType), 0, 0)

	callback := Callbacks83_09[this.SubpacketType]
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
	layout := NewTopAlignLayout()

	n, pref := humanize.ComputeSI(this.MemoryStats.TotalServerMemory)
	layout.AddWidget(NewQLabelF("Total server memory: %G %sB", n, pref), 0, 0)
	layout.AddWidget(NewLabel("Memory by category:"), 0, 0)

	memoryStatsList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Used memory"})
	rootItem := standardModel.InvisibleRootItem()

	developerRoot := NewStringItem("Developer's categories")
	for _, stats := range this.MemoryStats.DeveloperTags {
		n, pref := humanize.ComputeSI(stats.Memory)
		developerRoot.AppendRow([]*gui.QStandardItem{
			NewStringItem(stats.Name),
			NewQStandardItemF("%G %sB", n, pref),
		})
	}
	rootItem.AppendRow([]*gui.QStandardItem{developerRoot})

	internalRoot := NewStringItem("Internal categories")
	for _, stats := range this.MemoryStats.InternalCategories {
		n, pref := humanize.ComputeSI(stats.Memory)
		internalRoot.AppendRow([]*gui.QStandardItem{
			NewStringItem(stats.Name),
			NewQStandardItemF("%G %sB", n, pref),
		})
	}
	memoryStatsList.SetSortingEnabled(true)
	memoryStatsList.SetModel(standardModel)
	layout.AddWidget(memoryStatsList, 0, 0)

	layout.AddWidget(NewQLabelF("DataStore enabled: %v", this.DataStoreStats.Enabled), 0, 0)
	if this.DataStoreStats.Enabled {
		layout.AddWidget(NewQLabelF("DataStore GetAsync: %G", this.DataStoreStats.GetAsync), 0, 0)
		layout.AddWidget(NewQLabelF("DataStore Set/IncrementAsync: %G", this.DataStoreStats.SetAndIncrementAsync), 0, 0)
		layout.AddWidget(NewQLabelF("DataStore UpdateAsync: %G", this.DataStoreStats.UpdateAsync), 0, 0)
		layout.AddWidget(NewQLabelF("DataStore GetSortedAsync: %G", this.DataStoreStats.GetSortedAsync), 0, 0)
		layout.AddWidget(NewQLabelF("DataStore SetIncrementSortedAsync: %G", this.DataStoreStats.SetIncrementSortedAsync), 0, 0)
		layout.AddWidget(NewQLabelF("DataStore OnUpdate: %G", this.DataStoreStats.OnUpdate), 0, 0)
	}

	layout.AddWidget(NewLabel("Job stats:"), 0, 0)
	jobStatsList := widgets.NewQTreeView(nil)
	standardModel = NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Duty cycle (%)", "Op/s (1/s)", "Time/op (ms)"})

	rootNode := standardModel.InvisibleRootItem()
	for _, job := range this.JobStats {
		rootNode.AppendRow([]*gui.QStandardItem{
			NewStringItem(job.Name),
			NewQStandardItemF("%G", job.Stat1),
			NewQStandardItemF("%G", job.Stat2),
			NewQStandardItemF("%G", job.Stat3),
		})
	}
	jobStatsList.SetModel(standardModel)
	jobStatsList.SetSelectionMode(0)
	jobStatsList.SetSortingEnabled(true)
	layout.AddWidget(jobStatsList, 0, 0)

	layout.AddWidget(NewLabel("Script stats:"), 0, 0)
	scriptStatsList := widgets.NewQTreeView(nil)
	standardModel = NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Activity (%)", "Rate (1/s)"})

	rootNode = standardModel.InvisibleRootItem()
	for _, script := range this.ScriptStats {
		rootNode.AppendRow([]*gui.QStandardItem{
			NewStringItem(script.Name),
			NewQStandardItemF("%G", script.Stat1),
			NewUintItem(script.Stat2),
		})
	}
	scriptStatsList.SetModel(standardModel)
	scriptStatsList.SetSelectionMode(0)
	scriptStatsList.SetSortingEnabled(true)
	layout.AddWidget(scriptStatsList, 0, 0)

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
	layerLayout := NewTopAlignLayout()
	hashListLabel := NewLabel("Hashes:")
	layerLayout.AddWidget(hashListLabel, 0, 0)

	hashList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Hash"})
	rootItem := standardModel.InvisibleRootItem()
	for index, hash := range this.HashList {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewUintItem(index),
			NewQStandardItemF("%08X", hash),
		})
	}
	for index, hash := range this.SecurityTokens {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("ST%d", index),
			NewQStandardItemF("%016X", hash),
		})
	}
	hashList.SetSortingEnabled(true)
	hashList.SetModel(standardModel)
	layerLayout.AddWidget(hashList, 0, 0)

	widget.SetLayout(layerLayout)

	return widget
}

func show83_13(t peer.Packet83Subpacket) widgets.QWidget_ITF {
	this := t.(*peer.Packet83_13)
	widget := widgets.NewQWidget(nil, 0)
	layout := NewTopAlignLayout()
	layout.AddWidget(NewQLabelF("Instance: %s", this.Instance.GetFullName()), 0, 0)
	layout.AddWidget(NewQLabelF("Reference: %s", this.Instance.Ref.String()), 0, 0)
	if this.Parent != nil {
		layout.AddWidget(NewQLabelF("Parent: %s", this.Parent.GetFullName()), 0, 0)
		layout.AddWidget(NewQLabelF("Parent ref: %s", this.Parent.Ref.String()), 0, 0)
	} else {
		layout.AddWidget(NewLabel("Parent: nil"), 0, 0)
	}

	widget.SetLayout(layout)

	return widget
}

func showPacket83Subpacket(this Packet83Subpacket) widgets.QWidget_ITF {
	return SubpacketCallbacks[this.Type()](this)
}

func ShowPacket83(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet83Layer)

	mainSplitter := widgets.NewQSplitter(nil)
	mainSplitter.SetOrientation(core.Qt__Horizontal)

	subWindow := widgets.NewQWidget(nil, 0)
	subWindowLayout := NewTopAlignLayout()
	subWindowLayout.AddWidget(NewLabel("No replication subpacket selected!"), 0, 0)

	packetList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Packet"})
	rootItem := standardModel.InvisibleRootItem()
	for index, subpacket := range MainLayer.SubPackets {
		rootItem.AppendRow([]*gui.QStandardItem{
			NewUintItem(index),
			NewStringItem(subpacket.String()),
		})
	}
	packetList.SetSelectionMode(1)
	packetList.SetSortingEnabled(true)
	packetList.SortByColumn(0, core.Qt__AscendingOrder)
	packetList.SetModel(standardModel)
	packetList.ConnectClicked(func(index *core.QModelIndex) {
		// We must preserve this sizes because they will be reset
		// when the subWindow is removed
		oldSizes := mainSplitter.Sizes()

		thisIndex, _ := strconv.Atoi(standardModel.Item(index.Row(), 0).Data(0).ToString())
		subpacket := MainLayer.SubPackets[thisIndex]
		subWindow.DestroyQWidget()
		subWindow = widgets.NewQWidget(nil, 0)
		subWindowLayout := NewTopAlignLayout()
		subWindowLayout.SetAlign(core.Qt__AlignTop)

		showCallback, ok := SubpacketCallbacks[subpacket.Type()]
		if !ok {
			subWindowLayout.AddWidget(NewQLabelF("Unsupported packet type %d", subpacket.Type()), 0, 0)
		} else {
			subWindowLayout.AddWidget(showCallback(subpacket), 0, 0)
		}

		subWindow.SetLayout(subWindowLayout)
		mainSplitter.AddWidget(subWindow)

		mainSplitter.SetSizes(oldSizes)
	})
	mainSplitter.AddWidget(packetList)
	mainSplitter.AddWidget(subWindow)
	subWindow.SetLayout(subWindowLayout)

	layerLayout.AddWidget(mainSplitter, 0, 0)
}
