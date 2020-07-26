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
	KIND_DATA_STREAM_DATA_INSTANCE
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

	packetRows        map[uint64]*gtk.TreePath
	packetStore       map[uint64]*peer.PacketLayers
	packetTypeApplied map[uint64]bool
	lazyLoadFakeRows  map[uint64]*gtk.TreePath
}

func NewPacketListViewer(title string) (*PacketListViewer, error) {
	viewer := &PacketListViewer{
		packetRows:        make(map[uint64]*gtk.TreePath),
		packetStore:       make(map[uint64]*peer.PacketLayers),
		packetTypeApplied: make(map[uint64]bool),
		lazyLoadFakeRows:  make(map[uint64]*gtk.TreePath),
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

	for i, colName := range []string{"ID", "Packet", "Direction", "Length in bytes"} {
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
	treeView.Connect("row-expanded", func(_ *gtk.TreeView, iter *gtk.TreeIter, sortFilterPath *gtk.TreePath) {
		baseId, err := viewer.uint64FromIter(iter, COL_ID)
		if err != nil {
			println("failed to base id from selection")
			return
		}
		kind, err := viewer.uint64FromIter(iter, COL_PACKET_KIND)
		if err != nil {
			println("failed to get packet kind from selection")
			return
		}

		if kind != KIND_MAIN {
			// Packets that aren't MAIN will never have lazy-loading
			return
		}

		if fakeRow, ok := viewer.lazyLoadFakeRows[baseId]; ok {
			fakeIter, err := model.GetIter(fakeRow)
			if err != nil {
				println("failed to fake iter from selection")
				return
			}
			model.Remove(fakeIter)

			filterPath := sortModel.ConvertPathToChildPath(sortFilterPath)
			if filterPath == nil {
				println("failed to filterpath from selection")
				return
			}
			modelPath := filterModel.ConvertPathToChildPath(filterPath)
			if filterPath == nil {
				println("failed to modelpath from selection")
				return
			}
			iter, err = model.GetIter(modelPath)
			if err != nil {
				println("failed to real iter from selection")
				return
			}

			viewer.addSubpackets(iter, viewer.packetStore[baseId])
			delete(viewer.lazyLoadFakeRows, baseId)
			viewer.treeView.ExpandRow(sortFilterPath, false)
		}
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

func (viewer *PacketListViewer) NotifyACK(layers *peer.PacketLayers) {
	id := layers.UniqueID
	model := viewer.model
	viewer.packetStore[id] = layers
	var direction string

	if layers.Root.FromClient {
		direction = "C->S"
	} else if layers.Root.FromServer {
		direction = "S->C"
	} else {
		direction = "???"
	}
	var newRow gtk.TreeIter
	model.InsertWithValues(&newRow, nil, -1, []int{COL_ID, COL_PACKET, COL_DIRECTION, COL_HAS_LENGTH, COL_PACKET_KIND}, []interface{}{
		int64(id),
		layers.String(),
		direction,
		false,
		int64(KIND_MAIN),
	})

	var err error
	viewer.packetRows[id], err = model.GetPath(&newRow)
	if err != nil {
		println("failed to get path:", err.Error())
	}
}

func (viewer *PacketListViewer) updatePacketInfo(iter *gtk.TreeIter, layers *peer.PacketLayers) {
	model := viewer.model
	if !viewer.packetTypeApplied[layers.UniqueID] && layers.SplitPacket.HasPacketType {
		model.SetValue(iter, COL_PACKET, layers.String())
		viewer.packetTypeApplied[layers.UniqueID] = true
	}
	viewer.packetStore[layers.UniqueID] = layers
}

func (viewer *PacketListViewer) appendPartialPacketRow(layers *peer.PacketLayers) {
	id := layers.UniqueID
	model := viewer.model
	var direction string

	if layers.Root.FromClient {
		direction = "C->S"
	} else if layers.Root.FromServer {
		direction = "S->C"
	} else {
		direction = "???"
	}
	var newRow gtk.TreeIter
	model.InsertWithValues(&newRow, nil, -1, []int{COL_ID, COL_COLOR, COL_HAS_LENGTH, COL_PACKET_KIND, COL_DIRECTION}, []interface{}{
		int64(id),
		"rgba(255,255,0,.5)",
		true,
		int64(KIND_MAIN),
		direction,
	},
	)

	var err error
	viewer.packetRows[id], err = model.GetPath(&newRow)
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
			var newRow gtk.TreeIter
			err := model.InsertWithValues(&newRow, iter, -1, []int{
				COL_ID,
				COL_MAIN_PACKET_ID,
				COL_PACKET,
				COL_HAS_LENGTH,
				COL_PACKET_KIND,
			}, []interface{}{
				int64(index),
				int64(layers.UniqueID),
				subpacket.String(),
				false,
				int64(KIND_DATA_REPLIC),
			})
			if err != nil {
				println("failed to insert rows:", err.Error())
			}

			if joinData, ok := subpacket.(*peer.Packet83_0B); ok {
				for instanceIndex, instance := range joinData.Instances {
					model.InsertWithValues(nil, &newRow, -1, []int{
						COL_ID,
						COL_MAIN_PACKET_ID,
						COL_SUBPACKET_ID,
						COL_PACKET,
						COL_HAS_LENGTH,
						COL_PACKET_KIND,
					}, []interface{}{
						int64(instanceIndex),
						int64(layers.UniqueID),
						int64(index),
						instance.Instance.Ref.String() + ": " + instance.Instance.Name(),
						false,
						int64(KIND_DATA_JOIN_DATA_INSTANCE),
					})
				}
			} else if streamData, ok := subpacket.(*peer.Packet83_0D); ok {
				for instanceIndex, instance := range streamData.Instances {
					model.InsertWithValues(nil, &newRow, -1, []int{
						COL_ID,
						COL_MAIN_PACKET_ID,
						COL_SUBPACKET_ID,
						COL_PACKET,
						COL_HAS_LENGTH,
						COL_PACKET_KIND,
					}, []interface{}{
						int64(instanceIndex),
						int64(layers.UniqueID),
						int64(index),
						instance.Instance.Ref.String() + ": " + instance.Instance.Name(),
						false,
						int64(KIND_DATA_STREAM_DATA_INSTANCE),
					})
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

func (viewer *PacketListViewer) addLazySubpackets(iter *gtk.TreeIter, layers *peer.PacketLayers) {
	model := viewer.model

	var lazyIter *gtk.TreeIter
	switch layers.PacketType {
	case 0x83:
		mainLayer := layers.Main.(*peer.Packet83Layer)
		if len(mainLayer.SubPackets) > 0 {
			lazyIter = model.Append(iter)
		}
	case 0x85:
		mainLayer := layers.Main.(*peer.Packet85Layer)
		if len(mainLayer.SubPackets) > 0 {
			lazyIter = model.Append(iter)
		}
	case 0x86:
		mainLayer := layers.Main.(*peer.Packet86Layer)
		if len(mainLayer.SubPackets) > 0 {
			lazyIter = model.Append(iter)
		}
	}
	if lazyIter != nil {
		var err error
		viewer.lazyLoadFakeRows[layers.UniqueID], err = model.GetPath(lazyIter)
		if err != nil {
			println("failed to get lazy iter path:", err.Error())
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
		delete(viewer.packetTypeApplied, layers.UniqueID)

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
			viewer.addLazySubpackets(iter, layers)
		}
		viewer.model.SetValue(iter, COL_LEN_BYTES, int64(layers.SplitPacket.RealLength))
		viewer.model.SetValue(iter, COL_PACKET, layers.String())
	}
}

func (viewer *PacketListViewer) NotifyPacket(channel string, layers *peer.PacketLayers, forgetAcks bool) {
	if channel == "offline" {
		viewer.NotifyOfflinePacket(layers)
	} else if channel == "reliable" {
		viewer.NotifyPartialPacket(layers)
	} else if channel == "full-reliable" {
		viewer.NotifyFullPacket(layers)
	} else if channel == "ack" && !forgetAcks {
		viewer.NotifyACK(layers)
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
		layers := viewer.packetStore[baseId]
		viewer.packetDetailsViewer.ShowPacket(layers)
		if layers.Main != nil {
			packetViewer, err := viewerForMainPacket(layers.Main)
			if err != nil {
				println("failed to get packet viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		} else if layers.Error != nil {
			packetViewer, err := blanketViewer("Error while decoding: " + layers.Error.Error() + "\n\n" + layers.String())
			if err != nil {
				println("failed to get packet viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		} else if layers.RakNet.Flags.IsACK || layers.RakNet.Flags.IsNAK {
			packetViewer, err := ackViewer(layers.RakNet)
			if err != nil {
				println("failed to get packet viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		} else {
			packetViewer, err := blanketViewer("Incomplete packet")
			if err != nil {
				println("failed to get packet viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		}
	} else {
		mainPacketId, err := viewer.uint64FromIter(treeIter, COL_MAIN_PACKET_ID)
		if err != nil {
			println("failed to parent id from selection:", err.Error())
			return
		}
		viewer.packetDetailsViewer.ShowPacket(viewer.packetStore[mainPacketId])

		switch kind {
		case KIND_DATA_REPLIC:
			packetViewer, err := viewerForDataPacket(viewer.packetStore[mainPacketId].Main.(*peer.Packet83Layer).SubPackets[baseId])
			if err != nil {
				println("failed to get subpacket viewer:", err.Error())
				return
			}
			viewer.packetDetailsViewer.ShowMainLayer(packetViewer)
		case KIND_DATA_JOIN_DATA_INSTANCE, KIND_DATA_STREAM_DATA_INSTANCE:
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
			subpacket := viewer.packetStore[mainPacketId].Main.(*peer.Packet83Layer).SubPackets[joinDataSubpacket]
			var joinDataInstances []*peer.ReplicationInstance
			if kind == KIND_DATA_JOIN_DATA_INSTANCE {
				joinDataInstances = subpacket.(*peer.Packet83_0B).Instances
			} else {
				joinDataInstances = subpacket.(*peer.Packet83_0D).Instances
			}

			instViewer.ViewInstance(joinDataInstances[baseId])
			instViewer.mainWidget.ShowAll()

			viewer.packetDetailsViewer.ShowMainLayer(instViewer.mainWidget)
		case KIND_PHYSICS:
			physicsPacketViewer, err := NewPhysicsPacketViewer()
			if err != nil {
				println("failed to create physics packet viewer:", err.Error())
				return
			}
			physicsPacketViewer.ViewPacket(viewer.packetStore[mainPacketId].Main.(*peer.Packet85Layer).SubPackets[baseId])
			physicsPacketViewer.mainWidget.ShowAll()

			viewer.packetDetailsViewer.ShowMainLayer(physicsPacketViewer.mainWidget)
		case KIND_TOUCH:
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
