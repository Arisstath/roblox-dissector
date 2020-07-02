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
)

type PacketListViewer struct {
	title string

	treeView    *gtk.TreeView
	colRenderer *gtk.CellRendererText
	model       *gtk.TreeStore
	filterModel *gtk.TreeModelFilter
	sortModel   *gtk.TreeModelSort
}

func NewPacketListViewer(title string) (*PacketListViewer, error) {
	viewer := &PacketListViewer{}
	model, err := gtk.TreeStoreNew(
		glib.TYPE_INT64,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_INT64,
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
		treeView.AppendColumn(col)
	}

	viewer.title = title
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

func (viewer *PacketListViewer) NotifyOfflinePacket(layers *peer.PacketLayers) error {
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
	return nil
}

func (viewer *PacketListViewer) NotifyPacket(channel string, layers *peer.PacketLayers) {
	var err error
	if channel == "offline" {
		err = viewer.NotifyOfflinePacket(layers)
	} else if channel == "reliable" {
		//viewer.NotifyPartialPacket(layers)
	} else if channel == "full-reliable" {
		//viewer.NotifyFullPacket(layers)
	}
	if err != nil {
		println("Error while adding packet:", err.Error())
	}
}
