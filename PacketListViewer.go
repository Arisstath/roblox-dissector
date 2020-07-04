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
)

type PacketListViewer struct {
	title string

	mainWidget  *gtk.Paned
	treeView    *gtk.TreeView
	colRenderer *gtk.CellRendererText
	model       *gtk.TreeStore
	filterModel *gtk.TreeModelFilter
	sortModel   *gtk.TreeModelSort

	packetDetailsViewer *PacketDetailsViewer

	packetRows map[uint64]*gtk.TreePath
}

func NewPacketListViewer(title string) (*PacketListViewer, error) {
	viewer := &PacketListViewer{
		packetRows: make(map[uint64]*gtk.TreePath),
	}
	model, err := gtk.TreeStoreNew(
		glib.TYPE_INT64,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_INT64,
		glib.TYPE_STRING,
	)
	if err != nil {
		return nil, err
	}
	filterModel, err := model.FilterNew(nil)
	if err != nil {
		return nil, err
	}
	filterModel.SetVisibleFunc(func(model *gtk.TreeModelFilter, iter *gtk.TreeIter, userData interface{}) bool {
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
	colRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}

	for i, colName := range []string{"ID", "Packet", "Direction", "Length in Bytes"} {
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
		treeView.AppendColumn(col)
	}

	scrolledList, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	treeView.SetVExpand(true)
	scrolledList.Add(treeView)

	packetDetailsViewer, err := NewPacketDetailsViewer()
	if err != nil {
		return nil, err
	}
	packetDetailsViewer.mainWidget.SetVExpand(true)
	packetDetailsViewer.mainWidget.SetVAlign(gtk.ALIGN_FILL)
	scrolledList.SetVExpand(true)
	scrolledList.SetVAlign(gtk.ALIGN_FILL)

	mainWidget, err := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	mainWidget.Add(scrolledList)
	mainWidget.Add(packetDetailsViewer.mainWidget)

	viewer.title = title
	viewer.packetDetailsViewer = packetDetailsViewer
	viewer.mainWidget = mainWidget
	viewer.treeView = treeView
	viewer.colRenderer = colRenderer
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
	model.SetValue(iter, COL_COLOR, "rgba(255,255,0,.5)")
}

func (viewer *PacketListViewer) appendPartialPacketRow(layers *peer.PacketLayers) {
	id := layers.UniqueID
	model := viewer.model
	newRow := model.Append(nil)
	model.SetValue(newRow, COL_ID, int64(id))
	viewer.updatePacketInfo(newRow, layers)

	var err error
	viewer.packetRows[id], err = model.GetPath(newRow)
	if err != nil {
		println("failed to get path:", err.Error())
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
    	iter, err := viewer.model.GetIter(existingRow)
		if err != nil {
			println("failed to get iter:", err.Error())
			return
		}
		viewer.model.SetValue(iter, COL_COLOR, "rgba(255,255,255,1)") // finished with this packet
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
