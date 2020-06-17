package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/olebedev/emitter"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"github.com/yuin/gopher-lua"
)

type PacketList map[uint32]([]*gui.QStandardItem)
type PacketLayerList map[uint64]*peer.PacketLayers

// PacketListViewer is a Widget that stores packets
// captured from one Roblox communication session
type PacketListViewer struct {
	*widgets.QWidget
	TreeView      *widgets.QTreeView
	StandardModel *gui.QStandardItemModel
	ProxyModel    *core.QSortFilterProxyModel
	RootNode      *gui.QStandardItem

	UpdateInterval    time.Duration
	PendingRows       [][]*gui.QStandardItem
	UpdateContext     context.Context
	UpdatePassthrough bool
	UpdatePaused      bool

	packetRowsByUniqueID PacketList
	Packets              PacketLayerList
	PacketIndex          uint64

	InjectPacket chan peer.RakNetPacket

	Context      *peer.CommunicationContext
	Conversation *Conversation

	DefaultPacketWindow *PacketDetailsViewer
	Filter *lua.FunctionProto
}

func NewPacketListViewer(updateContext context.Context, parent widgets.QWidget_ITF, flags core.Qt__WindowType) *PacketListViewer {
	listViewer := &PacketListViewer{
		QWidget:              widgets.NewQWidget(parent, flags),
		packetRowsByUniqueID: make(PacketList),

		Packets: make(PacketLayerList),

		InjectPacket: make(chan peer.RakNetPacket, 1),
		Context:      peer.NewCommunicationContext(),

		UpdateInterval: time.Duration(time.Second / 2),
		UpdateContext:  updateContext,
	}

	layout := NewTopAlignLayout()
	treeView := widgets.NewQTreeView(listViewer)
	listViewer.TreeView = treeView
	packetDetailsViewer := NewPacketDetailsViewer(listViewer, 0)

	mainSplitter := widgets.NewQSplitter(listViewer)
	mainSplitter.SetOrientation(core.Qt__Vertical)
	mainSplitter.AddWidget(treeView)
	mainSplitter.AddWidget(packetDetailsViewer)
	listViewer.DefaultPacketWindow = packetDetailsViewer
	layout.AddWidget(mainSplitter, 0, 0)

	standardModel, proxy := NewFilteringModel(treeView)
	proxy.ConnectFilterAcceptsRow(func(sourceRow int, sourceParent *core.QModelIndex) bool {
		return listViewer.FilterAcceptsRow(sourceRow, sourceParent)
	})
	listViewer.StandardModel = standardModel
	listViewer.ProxyModel = proxy
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Packet", "Direction", "Length in Bytes", "Datagram Numbers", "Ordered Splits", "Total Splits"})
	treeView.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
	treeView.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	treeView.SetSortingEnabled(true)
	treeView.SetUniformRowHeights(true)
	listViewer.RootNode = standardModel.InvisibleRootItem()

	treeView.SelectionModel().ConnectCurrentRowChanged(listViewer.SelectionChanged)
	treeView.SetContextMenuPolicy(core.Qt__CustomContextMenu)
	treeView.ConnectCustomContextMenuRequested(func(position *core.QPoint) {
		index := treeView.IndexAt(position)
		if index.IsValid() {
			realSelectedValue, _ := strconv.Atoi(standardModel.Item(proxy.MapToSource(index).Row(), 0).Data(0).ToString())
			if listViewer.Packets[uint64(realSelectedValue)] != nil {
				thisPacket := listViewer.Packets[uint64(realSelectedValue)]
				customPacketMenu := NewPacketViewerMenu(listViewer, listViewer.Context, thisPacket, ActivationCallbacks[thisPacket.PacketType])
				customPacketMenu.Exec2(treeView.Viewport().MapToGlobal(position), nil)
			}
		}
	})

	listViewer.SetLayout(layout)
	go listViewer.UpdateLoop()
	return listViewer
}

func (m *PacketListViewer) SetFilter(filterScript string) {
	compiled, err := CompileFilter(filterScript)
	if err != nil {
    	widgets.QMessageBox_Critical(m, "Filter Error", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
    	return
	}
	m.Filter = compiled
	m.ProxyModel.InvalidateFilter()
}

func (m *PacketListViewer) FilterAcceptsRow(sourceRow int, sourceParent *core.QModelIndex) bool {
    if m.Filter == nil {
        return true
    }
	realSelectedValue, _ := strconv.Atoi(m.StandardModel.Item(sourceRow, 0).Data(0).ToString())
	if packet, ok := m.Packets[uint64(realSelectedValue)]; ok {
    	if packet.Main == nil {
        	return true
    	}
		acc, err := FilterAcceptsPacket(NewLuaFilterState(), m.Filter, packet.Main, realSelectedValue)
		if err != nil {
        	widgets.QMessageBox_Critical(m, fmt.Sprintf("Filter Error on packet %d", realSelectedValue), fmt.Sprintf("Error: %s\n\nThe filter will be reset.", err.Error()), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
        	m.Filter = nil
			return true
		}
		return acc
	}
	println("Warning: Packet to be filtered not found")
	return true
}

func (m *PacketListViewer) registerSplitPacketRow(row []*gui.QStandardItem, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID] = row
}

func (m *PacketListViewer) AddSplitPacket(context *peer.CommunicationContext, layers *peer.PacketLayers) {
	if _, ok := m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID]; !ok {
		m.AddFullPacket(context, layers, nil)
	} else {
		m.handleSplitPacket(context, layers)
	}
}

func (m *PacketListViewer) BindCallback(context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) {
	row := m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID]
	index, _ := strconv.Atoi(row[0].Data(0).ToString())

	m.Packets[uint64(index)] = layers
	row[1].SetData(core.NewQVariant14(layers.String()), 0)

	for _, item := range row {
		if layers.Main != nil {
			item.SetBackground(gui.NewQBrush2(core.Qt__NoBrush))
		} else {
			paintItems(row, gui.NewQColor3(255, 0, 0, 127))
		}
	}
}

func (m *PacketListViewer) handleSplitPacket(context *peer.CommunicationContext, layers *peer.PacketLayers) {
	row := m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID]
	m.registerSplitPacketRow(row, context, layers)

	row[3].SetData(core.NewQVariant7(int(layers.Reliability.SplitBuffer.RealLength)), 0)
	if layers.Reliability.SplitBuffer.RakNetPackets[0] == nil {
		panic(errors.New("encountered nil first raknet"))
	}
	row[4].SetData(core.NewQVariant14(fmt.Sprintf("%d - %d", layers.Reliability.SplitBuffer.RakNetPackets[0].DatagramNumber, layers.RakNet.DatagramNumber)), 0)
	row[5].SetData(core.NewQVariant14(fmt.Sprintf("%d/%d", layers.Reliability.SplitBuffer.NumReceivedSplits, layers.Reliability.SplitPacketCount)), 0)
	row[6].SetData(core.NewQVariant7(len(layers.Reliability.SplitBuffer.RakNetPackets)), 0)
}

func (m *PacketListViewer) AddFullPacket(context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) []*gui.QStandardItem {
	index := m.PacketIndex
	m.PacketIndex++
	isClient := layers.Root.FromClient
	isServer := layers.Root.FromServer

	indexItem := NewUintItem(index)
	packetTypeItem := NewStringItem(layers.String())

	rootRow := []*gui.QStandardItem{indexItem, packetTypeItem}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewStringItem("C->S")
	} else if isServer {
		direction = NewStringItem("S->C")
	} else {
		direction = NewStringItem("???")
	}

	rootRow = append(rootRow, direction)

	var length *gui.QStandardItem
	if layers.Reliability != nil {
		length = NewUintItem(layers.Reliability.LengthInBits / 8)
	} else if layers.OfflinePayload != nil {
		length = NewUintItem(len(layers.OfflinePayload))
	} else {
		length = nil
	}
	rootRow = append(rootRow, length)
	var datagramNumber *gui.QStandardItem
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.SplitBuffer.RakNetPackets

		firstLayer := allRakNetLayers[0]
		lastLayer := allRakNetLayers[len(allRakNetLayers)-1]
		var firstLayerNumber, lastLayerNumber int32
		if firstLayer == nil {
			println("Encountered nil first raknet")
			firstLayerNumber = -1
		} else {
			firstLayerNumber = int32(firstLayer.DatagramNumber)
		}
		if lastLayer == nil {
			println("Encountered nil last raknet")
			lastLayerNumber = -1
		} else {
			lastLayerNumber = int32(lastLayer.DatagramNumber)
		}

		datagramNumber = NewQStandardItemF("%d - %d", firstLayerNumber, lastLayerNumber)
	} else if layers.RakNet != nil {
		datagramNumber = NewUintItem(layers.RakNet.DatagramNumber)
	} else {
		datagramNumber = nil
	}
	rootRow = append(rootRow, datagramNumber)

	if layers.Reliability != nil {
		receivedSplits := NewQStandardItemF("%d/%d", layers.Reliability.SplitBuffer.NumReceivedSplits, layers.Reliability.SplitPacketCount)
		rootRow = append(rootRow, receivedSplits)
	} else {
		rootRow = append(rootRow, nil)
	}

	if layers.Reliability != nil {
		rootRow = append(rootRow, NewStringItem("1"))
		m.registerSplitPacketRow(rootRow, context, layers)
	} else {
		rootRow = append(rootRow, nil)
	}

	m.Packets[index] = layers
	if layers.Reliability != nil {
		paintItems(rootRow, gui.NewQColor3(255, 255, 0, 127))
	}
	if m.UpdatePassthrough {
		m.RootNode.InsertRow(m.RootNode.RowCount(), rootRow)
	} else {
		m.PendingRows = append(m.PendingRows, rootRow)
	}

	return rootRow
}

func (viewer *PacketListViewer) BindToConversation(conv *Conversation) {
	viewer.Conversation = conv
	context := conv.Context
	clientPacketEmitter := conv.ClientReader.Layers()
	clientErrorEmitter := conv.ClientReader.Errors()
	serverPacketEmitter := conv.ServerReader.Layers()
	serverErrorEmitter := conv.ServerReader.Errors()

	offlineHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		MainThreadRunner.RunOnMain(func() {
			viewer.AddFullPacket(context, layers, ActivationCallbacks[layers.PacketType])
		})
		<-MainThreadRunner.Wait
	}
	reliableHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		MainThreadRunner.RunOnMain(func() {
			viewer.AddSplitPacket(context, layers)
		})
		<-MainThreadRunner.Wait
	}
	fullReliableHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		MainThreadRunner.RunOnMain(func() {
			viewer.BindCallback(context, layers, ActivationCallbacks[layers.PacketType])
		})
		<-MainThreadRunner.Wait
	}

	clientPacketEmitter.On("offline", offlineHandler, emitter.Void)
	clientPacketEmitter.On("reliable", reliableHandler, emitter.Void)
	clientPacketEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	clientErrorEmitter.On("offline", offlineHandler, emitter.Void)
	clientErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	clientErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)

	serverPacketEmitter.On("offline", offlineHandler, emitter.Void)
	serverPacketEmitter.On("reliable", reliableHandler, emitter.Void)
	serverPacketEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	serverErrorEmitter.On("offline", offlineHandler, emitter.Void)
	serverErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	serverErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
}

func (m *PacketListViewer) SelectionChanged(index, _ *core.QModelIndex) {
	proxy := m.ProxyModel
	standardModel := m.StandardModel
	listViewer := m
	realSelectedValue, _ := strconv.Atoi(standardModel.Item(proxy.MapToSource(index).Row(), 0).Data(0).ToString())
	if listViewer.Packets[uint64(realSelectedValue)] != nil {
		thisPacket := listViewer.Packets[uint64(realSelectedValue)]
		listViewer.DefaultPacketWindow.Update(listViewer.Context, thisPacket, ActivationCallbacks[thisPacket.PacketType])
	}
}

func (m *PacketListViewer) UpdateModel() {
	m.TreeView.SetModel(m.ProxyModel)
	m.TreeView.SelectionModel().ConnectCurrentRowChanged(m.SelectionChanged)
	m.TreeView.SortByColumn(0, core.Qt__AscendingOrder)
}

func (m *PacketListViewer) UpdateLoop() {
	ticker := time.NewTicker(m.UpdateInterval)
	for {
		select {
		case <-m.UpdateContext.Done():
			MainThreadRunner.RunOnMain(func() {
				m.UpdatePassthrough = true
				for _, row := range m.PendingRows {
					m.RootNode.InsertRow(m.RootNode.RowCount(), row)
				}
				m.PendingRows = nil
			})
			<-MainThreadRunner.Wait
			return
		case <-ticker.C:
			MainThreadRunner.RunOnMain(func() {
				if !m.UpdatePaused {
					for _, row := range m.PendingRows {
						m.RootNode.InsertRow(m.RootNode.RowCount(), row)
					}
					m.PendingRows = nil
				}
			})
			<-MainThreadRunner.Wait
			// Clear
		}
	}
}
