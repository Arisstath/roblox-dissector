package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/olebedev/emitter"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
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

	packetRowsByUniqueID PacketList
	Packets              PacketLayerList
	PacketIndex          uint64

	IsCapturing       bool
	CaptureJobContext context.Context
	StopCaptureJob    context.CancelFunc
	InjectPacket      chan peer.RakNetPacket

	Context *peer.CommunicationContext

	DefaultPacketWindow *PacketDetailsViewer
}

func NewPacketListViewer(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *PacketListViewer {
	captureContext, captureCancel := context.WithCancel(context.Background())
	listViewer := &PacketListViewer{
		QWidget:              widgets.NewQWidget(parent, flags),
		packetRowsByUniqueID: make(PacketList),

		Packets: make(PacketLayerList),

		CaptureJobContext: captureContext,
		StopCaptureJob:    captureCancel,

		InjectPacket: make(chan peer.RakNetPacket, 1),
		Context:      peer.NewCommunicationContext(),
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
		// TODO
		return true
	})
	listViewer.StandardModel = standardModel
	listViewer.ProxyModel = proxy
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Direction", "Length in Bytes", "Datagram Numbers", "Ordered Splits", "Total Splits"})
	treeView.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
	treeView.SetSortingEnabled(true)
	listViewer.RootNode = standardModel.InvisibleRootItem()

	treeView.ConnectClicked(func(index *core.QModelIndex) {
		realSelectedValue, _ := strconv.Atoi(standardModel.Item(proxy.MapToSource(index).Row(), 0).Data(0).ToString())
		if listViewer.Packets[uint64(realSelectedValue)] != nil {
			thisPacket := listViewer.Packets[uint64(realSelectedValue)]
			listViewer.DefaultPacketWindow.Update(listViewer.Context, thisPacket, ActivationCallbacks[thisPacket.PacketType])
		}
	})
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

	return listViewer
}

func (m *PacketListViewer) registerSplitPacketRow(row []*gui.QStandardItem, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID] = row
}

func (m *PacketListViewer) AddSplitPacket(context *peer.CommunicationContext, layers *peer.PacketLayers) {
	if _, ok := m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID]; !ok {
		m.AddFullPacket(context, layers, nil)
		m.BindDefaultCallback(context, layers)
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

// TODO: Is this still needed?
func (m *PacketListViewer) BindDefaultCallback(context *peer.CommunicationContext, layers *peer.PacketLayers) {
	row := m.packetRowsByUniqueID[layers.Reliability.SplitBuffer.UniqueID]
	index, _ := strconv.Atoi(row[0].Data(0).ToString())

	m.Packets[uint64(index)] = layers
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
	m.PacketIndex++
	index := m.PacketIndex
	isClient := layers.Root.FromClient
	isServer := layers.Root.FromServer

	indexItem := NewQStandardItemF("%d", index)
	packetTypeItem := NewQStandardItemF(layers.String())

	rootRow := []*gui.QStandardItem{indexItem, packetTypeItem}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewQStandardItemF("C->S")
	} else if isServer {
		direction = NewQStandardItemF("S->C")
	} else {
		direction = NewQStandardItemF("???")
	}

	rootRow = append(rootRow, direction)

	var length *gui.QStandardItem
	if layers.Reliability != nil {
		length = NewQStandardItemF("%d", layers.Reliability.LengthInBits/8)
	} else {
		length = NewQStandardItemF("???")
	}
	rootRow = append(rootRow, length)
	var datagramNumber *gui.QStandardItem
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.SplitBuffer.RakNetPackets

		firstLayer := allRakNetLayers[0]
		lastLayer := allRakNetLayers[len(allRakNetLayers)-1]
		var firstLayerNumber, lastLayerNumber int32
		if firstLayer == nil {
			fmt.Printf("Encountered nil first raknet")
			firstLayerNumber = -1
		} else {
			firstLayerNumber = int32(firstLayer.DatagramNumber)
		}
		if lastLayer == nil {
			fmt.Printf("Encountered nil last raknet")
			lastLayerNumber = -1
		} else {
			lastLayerNumber = int32(lastLayer.DatagramNumber)
		}

		datagramNumber = NewQStandardItemF("%d - %d", firstLayerNumber, lastLayerNumber)
	} else {
		datagramNumber = NewQStandardItemF("%d", layers.RakNet.DatagramNumber)
	}
	rootRow = append(rootRow, datagramNumber)

	if layers.Reliability != nil {
		receivedSplits := NewQStandardItemF("%d/%d", layers.Reliability.SplitBuffer.NumReceivedSplits, layers.Reliability.SplitPacketCount)
		rootRow = append(rootRow, receivedSplits)
	} else {
		rootRow = append(rootRow, nil)
	}
	rootRow = append(rootRow, NewQStandardItemF("???"))

	if layers.Reliability != nil {
		m.registerSplitPacketRow(rootRow, context, layers)
	}

	if layers.Reliability == nil { // Only bind if we're done parsing the packet
		m.Packets[index] = layers
	} else {
		paintItems(rootRow, gui.NewQColor3(255, 255, 0, 127))
	}
	m.RootNode.AppendRow(rootRow)

	return rootRow
}

func (viewer *PacketListViewer) BindToConversation(conv *Conversation) {
	context := conv.Context
	clientPacketReader := conv.ClientReader
	serverPacketReader := conv.ServerReader

	simpleHandler := func(e *emitter.Event) {
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
	// ACK and ReliabilityLayer are nops

	clientPacketReader.BindDataModelHandlers()
	serverPacketReader.BindDataModelHandlers()

	clientPacketReader.LayerEmitter.On("simple", simpleHandler, emitter.Void)
	clientPacketReader.LayerEmitter.On("reliable", reliableHandler, emitter.Void)
	clientPacketReader.LayerEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("simple", simpleHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)

	serverPacketReader.LayerEmitter.On("simple", simpleHandler, emitter.Void)
	serverPacketReader.LayerEmitter.On("reliable", reliableHandler, emitter.Void)
	serverPacketReader.LayerEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("simple", simpleHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
}

func (m *PacketListViewer) UpdateModel() {
	MainThreadRunner.RunOnMain(func() {
		m.TreeView.SetModel(m.ProxyModel)
	})
	<-MainThreadRunner.Wait
}
