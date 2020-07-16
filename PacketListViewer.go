package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	COL_ID = iota
	COL_PACKET
	COL_DIRECTION
	COL_LEN_BYTES
	COL_COLOR
	COL_HAS_LENGTH
	COL_MAIN_PACKET_ID
	COL_SUBPACKET_ID
	COL_PACKET_KIND
)

const (
	KIND_MAIN = iota
	KIND_DATA_REPLIC
	KIND_DATA_JOIN_DATA_INSTANCE
	KIND_PHYSICS
	KIND_TOUCH
)

type PacketListViewer struct {
	title string

	mainWidget  *gtk.Paned
	treeView    *gtk.TreeView
	model       *gtk.TreeStore
	filterModel *gtk.TreeModelFilter
	sortModel   *gtk.TreeModelSort

	packetDetailsViewer *PacketDetailsViewer

	packetRows  map[uint64]*gtk.TreePath
	packetStore map[uint64]*peer.PacketLayers
}

func NewPacketListViewer(title string) (*PacketListViewer, error) {
	viewer := &PacketListViewer{
		packetRows:  make(map[uint64]*gtk.TreePath),
		packetStore: make(map[uint64]*peer.PacketLayers),
	}
	model, err := gtk.TreeStoreNew(
		glib.TYPE_INT64,   // COL_ID
		glib.TYPE_STRING,  // COL_PACKET
		glib.TYPE_STRING,  // COL_DIRECTION
		glib.TYPE_INT64,   // COL_LEN_BYTES
		glib.TYPE_STRING,  // COL_COLOR
		glib.TYPE_BOOLEAN, // COL_HAS_LENGTH
		glib.TYPE_INT64,   // COL_MAIN_PACKET_ID
		glib.TYPE_INT64,   // COL_SUBPACKET_ID
		glib.TYPE_INT64,   // COL_PACKET_KIND
	)
	if err != nil {
		return nil, err
	}
	filterModel, err := model.FilterNew(nil)
	if err != nil {
		return nil, err
	}
	filterModel.SetVisibleFunc(func(model *gtk.TreeModelFilter, iter *gtk.TreeIter, userData ...interface{}) bool {
		return viewer.FilterAcceptsPacket(model, iter, userData)
	})
	sortModel, err := gtk.TreeModelSortNew(filterModel)
	if err != nil {
		return nil, err
	}
	treeView, err := gtk.TreeViewNewWithModel(sortModel)
	if err != nil {
		return nil, err
	}

	for i, colName := range []string{"ID", "Packet", "Direction", "Length in Bytes"} {
		colRenderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		col, err := gtk.TreeViewColumnNewWithAttribute(
			colName,
			colRenderer,
			"text",
			i,
		)
		if err != nil {
			return nil, err
		}
		col.AddAttribute(colRenderer, "background", COL_COLOR)
		if i == COL_LEN_BYTES {
			col.AddAttribute(colRenderer, "visible", COL_HAS_LENGTH)
		}
		col.SetSortColumnID(i)
		treeView.AppendColumn(col)
	}

	scrolledList, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolledList.Add(treeView)

	packetDetailsViewer, err := NewPacketDetailsViewer()
	if err != nil {
		return nil, err
	}

	mainWidget, err := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	mainWidget.Add(scrolledList)
	mainWidget.Add(packetDetailsViewer.mainWidget)

	sel, err := treeView.GetSelection()
	if err != nil {
		return nil, err
	}
	sel.SetMode(gtk.SELECTION_SINGLE)
	sel.Connect("changed", func(selection *gtk.TreeSelection) {
		viewer.selectionChanged(selection)
	})

	viewer.title = title
	viewer.packetDetailsViewer = packetDetailsViewer
	viewer.mainWidget = mainWidget
	viewer.treeView = treeView
	viewer.model = model
	viewer.filterModel = filterModel
	viewer.sortModel = sortModel
	return viewer, nil
}

func (viewer *PacketListViewer) FilterAcceptsPacket(model *gtk.TreeModelFilter, iter *gtk.TreeIter, userData interface{}) bool {
	return true
}

func (viewer *PacketListViewer) NotifyOfflinePacket(layers *peer.PacketLayers) {
	id := layers.UniqueID
	model := viewer.model
	newRow := model.Append(nil)
	viewer.packetStore[id] = layers
	model.SetValue(newRow, COL_ID, int64(id))
	model.SetValue(newRow, COL_PACKET, layers.String())
	var direction string

	if layers.Root.FromClient {
		direction = "C->S"
	} else if layers.Root.FromServer {
		direction = "S->C"
	} else {
		direction = "???"
	}
	model.SetValue(newRow, COL_DIRECTION, direction)
	model.SetValue(newRow, COL_LEN_BYTES, int64(len(layers.OfflinePayload)))
	model.SetValue(newRow, COL_HAS_LENGTH, true)
	model.SetValue(newRow, COL_PACKET_KIND, int64(KIND_MAIN))
	if layers.Error != nil {
		model.SetValue(newRow, COL_COLOR, "rgba(255,0,0,.5)")
	}

	var err error
	viewer.packetRows[id], err = model.GetPath(newRow)
	if err != nil {
		println("failed to get path:", err.Error())
	}
}

func (viewer *PacketListViewer) updatePacketInfo(iter *gtk.TreeIter, layers *peer.PacketLayers) {
	model := viewer.model
	model.SetValue(iter, COL_PACKET, layers.String())
	var direction string

	if layers.Root.FromClient {
		direction = "C->S"
	} else if layers.Root.FromServer {
		direction = "S->C"
	} else {
		direction = "???"
	}
	model.SetValue(iter, COL_DIRECTION, direction)
	model.SetValue(iter, COL_LEN_BYTES, int64(layers.SplitPacket.RealLength))
	viewer.packetStore[layers.UniqueID] = layers
}

func (viewer *PacketListViewer) appendPartialPacketRow(layers *peer.PacketLayers) {
	id := layers.UniqueID
	model := viewer.model
	newRow := model.Append(nil)
	model.SetValue(newRow, COL_ID, int64(id))
	model.SetValue(newRow, COL_COLOR, "rgba(255,255,0,.5)")
	model.SetValue(newRow, COL_HAS_LENGTH, true)
	model.SetValue(newRow, COL_PACKET_KIND, int64(KIND_MAIN))
	viewer.updatePacketInfo(newRow, layers)

	var err error
	viewer.packetRows[id], err = model.GetPath(newRow)
	if err != nil {
		println("failed to get path:", err.Error())
	}
	viewer.packetStore[id] = layers
}

func (viewer *PacketListViewer) addSubpackets(iter *gtk.TreeIter, layers *peer.PacketLayers) {
	model := viewer.model
	switch layers.PacketType {
	case 0x83:
		mainLayer := layers.Main.(*peer.Packet83Layer)
		for index, subpacket := range mainLayer.SubPackets {
			newRow := model.Append(iter)
			model.SetValue(newRow, COL_ID, int64(index))
			model.SetValue(newRow, COL_MAIN_PACKET_ID, int64(layers.UniqueID))
			model.SetValue(newRow, COL_PACKET, subpacket.String())
			model.SetValue(newRow, COL_HAS_LENGTH, false)
			model.SetValue(newRow, COL_PACKET_KIND, int64(KIND_DATA_REPLIC))

			if joinData, ok := subpacket.(*peer.Packet83_0B); ok {
				for instanceIndex, instance := range joinData.Instances {
					instanceRow := model.Append(newRow)
					model.SetValue(instanceRow, COL_ID, int64(instanceIndex))
					model.SetValue(instanceRow, COL_MAIN_PACKET_ID, int64(layers.UniqueID))
					model.SetValue(instanceRow, COL_SUBPACKET_ID, int64(index))
					model.SetValue(instanceRow, COL_PACKET, instance.Instance.Ref.String()+": "+instance.Instance.Name())
					model.SetValue(instanceRow, COL_HAS_LENGTH, false)
					model.SetValue(instanceRow, COL_PACKET_KIND, int64(KIND_DATA_JOIN_DATA_INSTANCE))
				}
			}
		}
	case 0x85:
		mainLayer := layers.Main.(*peer.Packet85Layer)
		for index, subpacket := range mainLayer.SubPackets {
			newRow := model.Append(iter)
			model.SetValue(newRow, COL_ID, int64(index))
			model.SetValue(newRow, COL_MAIN_PACKET_ID, int64(layers.UniqueID))
			model.SetValue(newRow, COL_PACKET, subpacket.String())
			model.SetValue(newRow, COL_HAS_LENGTH, false)
			model.SetValue(newRow, COL_PACKET_KIND, int64(KIND_PHYSICS))
		}
	case 0x86:
		mainLayer := layers.Main.(*peer.Packet86Layer)
		for index, subpacket := range mainLayer.SubPackets {
			newRow := model.Append(iter)
			model.SetValue(newRow, COL_ID, int64(index))
			model.SetValue(newRow, COL_MAIN_PACKET_ID, int64(layers.UniqueID))
			model.SetValue(newRow, COL_PACKET, subpacket.String())
			model.SetValue(newRow, COL_HAS_LENGTH, false)
			model.SetValue(newRow, COL_PACKET_KIND, int64(KIND_TOUCH))
		}
	}
}

func (viewer *PacketListViewer) NotifyPartialPacket(layers *peer.PacketLayers) {
	existingRow, ok := viewer.packetRows[layers.UniqueID]

	if !ok {
		viewer.appendPartialPacketRow(layers)
	} else {
		iter, err := viewer.model.GetIter(existingRow)
		if err != nil {
			println("failed to get iter:", err.Error())
			return
		}
		viewer.updatePacketInfo(iter, layers)
	}
}

func (viewer *PacketListViewer) NotifyFullPacket(layers *peer.PacketLayers) {
	existingRow, ok := viewer.packetRows[layers.UniqueID]
	if !ok {
		println("haven't seen this full packet yet:", layers.UniqueID)
		return
	} else {
		delete(viewer.packetRows, layers.UniqueID)
		iter, err := viewer.model.GetIter(existingRow)
		if err != nil {
			println("failed to get iter:", err.Error())
			return
		}
		if layers.Error != nil {
			viewer.model.SetValue(iter, COL_COLOR, "rgba(255,0,0,.5)")
		} else {
			// Why doesn't GTK have `transparent` for colors, like CSS?
			viewer.model.SetValue(iter, COL_COLOR, "rgba(0,0,0,0)") // finished with this packet
		}
		viewer.updatePacketInfo(iter, layers)

		viewer.addSubpackets(iter, layers)
	}
}

func (viewer *PacketListViewer) NotifyPacket(channel string, layers *peer.PacketLayers) {
	if channel == "offline" {
		viewer.NotifyOfflinePacket(layers)
	} else if channel == "reliable" {
		viewer.NotifyPartialPacket(layers)
	} else if channel == "full-reliable" {
		viewer.NotifyFullPacket(layers)
	}
}

func (viewer *PacketListViewer) getGoValue(treeIter *gtk.TreeIter, col int) (interface{}, error) {
	id, err := viewer.sortModel.GetValue(treeIter, col)
	if err != nil {
		return nil, err
	}
	return id.GoValue()
}

func (viewer *PacketListViewer) goValueFromPath(path *gtk.TreePath, col int) (interface{}, error) {
	treeIter, err := viewer.sortModel.GetIter(path)
	if err != nil {
		return nil, err
	}
	return viewer.getGoValue(treeIter, col)
}

func (viewer *PacketListViewer) uint64FromIter(treeIter *gtk.TreeIter, col int) (uint64, error) {
	v, err := viewer.getGoValue(treeIter, col)
	if err != nil {
		return 0, err
	}
	return uint64(v.(int64)), nil
}

func (viewer *PacketListViewer) selectionChanged(selection *gtk.TreeSelection) {
	_, treeIter, ok := selection.GetSelected()
	if !ok {
		println("nothing selected")
		return
	}
	baseId, err := viewer.uint64FromIter(treeIter, COL_ID)
	if err != nil {
		println("failed to base id from selection")
		return
	}
	kind, err := viewer.uint64FromIter(treeIter, COL_PACKET_KIND)
	if err != nil {
		println("failed to get packet kind from selection")
		return
	}

	if kind == KIND_MAIN {
		viewer.packetDetailsViewer.ShowPacket(viewer.packetStore[baseId])
	} else {
		mainPacketId, err := viewer.uint64FromIter(treeIter, COL_MAIN_PACKET_ID)
		if err != nil {
			println("failed to parent id from selection:", err.Error())
			return
		}
		viewer.packetDetailsViewer.ShowPacket(viewer.packetStore[mainPacketId])

		if kind == KIND_DATA_REPLIC {
			packetViewer, err := viewerForDataPacket(viewer.packetStore[mainPacketId].Main.(*peer.Packet83Layer).SubPackets[baseId])
			if err != nil {
				println("failed to get subpacket viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		} else if kind == KIND_DATA_JOIN_DATA_INSTANCE {
			joinDataSubpacket, err := viewer.uint64FromIter(treeIter, COL_SUBPACKET_ID)
			if err != nil {
				println("failed to subpacket id from selection:", err.Error())
				return
			}

			instViewer, err := NewInstanceViewer()
			if err != nil {
				println("failed to make make subpacket window:", err.Error())
				return
			}
			instViewer.ViewInstance(viewer.packetStore[mainPacketId].Main.(*peer.Packet83Layer).SubPackets[joinDataSubpacket].(*peer.Packet83_0B).Instances[baseId])
			instViewer.mainWidget.ShowAll()

			viewer.packetDetailsViewer.ShowMainLayer(instViewer.mainWidget)
		} else if kind == KIND_PHYSICS {
			physicsPacketViewer, err := NewPhysicsPacketViewer()
			if err != nil {
				println("failed to create physics packet viewer:", err.Error())
				return
			}
			physicsPacketViewer.ViewPacket(viewer.packetStore[mainPacketId].Main.(*peer.Packet85Layer).SubPackets[baseId])
			physicsPacketViewer.mainWidget.ShowAll()

			viewer.packetDetailsViewer.ShowMainLayer(physicsPacketViewer.mainWidget)
		} else if kind == KIND_TOUCH {
			packet := viewer.packetStore[mainPacketId].Main.(*peer.Packet86Layer).SubPackets[baseId]
			packetViewer, err := blanketViewer(packet.String())
			if err != nil {
				println("failed to get subpacket viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		}
	}
}
