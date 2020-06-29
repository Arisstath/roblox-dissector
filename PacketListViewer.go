package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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
	model, err := gtk.TreeStoreNew(
		glib.TYPE_INT64,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_INT64,
	)
	if err != nil {
		return nil, err
	}
	initPath, err := gtk.TreePathNewFirst()
	if err != nil {
		return nil, err
	}
	filterModel, err := model.FilterNew(initPath)
	if err != nil {
		return nil, err
	}
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

	return &PacketListViewer{
		title, treeView, colRenderer, model, filterModel, sortModel,
	}, nil
}

func (viewer *PacketListViewer) FilterAcceptsPacket() bool {
	return true
}

func (viewer *PacketListViewer) NotifyPacket(layers *peer.PacketLayers) {

}
