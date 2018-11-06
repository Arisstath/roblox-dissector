package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/Gskartwii/roblox-dissector/peer"
import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/rbxfile/bin"
import "os"
import "os/exec"
import "fmt"
import "strconv"
import "sync/atomic"
import "sync"
import "net/http"
import "io/ioutil"
import "strings"
import "errors"

import "context"

var window *widgets.QMainWindow

type DefaultValues map[string](map[string]rbxfile.Value)

type PacketList map[uint32]([]*gui.QStandardItem)
type TwoWayPacketList struct {
	Server  PacketList
	Client  PacketList
	MServer *sync.Mutex
	MClient *sync.Mutex
	EServer *sync.Cond
	EClient *sync.Cond
}

type StudioSettings struct {
	Location string

	Flags   string
	Address string
	Port    string
	RBXL    string
}

type PlayerSettings struct {
	Location   string
	Flags      string
	GameID     string
	TrackerID  string
	AuthTicket string
}

type ServerSettings struct {
	Port                   string
	EnumSchemaLocation     string
	InstanceSchemaLocation string
	DictionaryLocation     string
}

type DefaultsSettings struct {
	Files []string
}

type PlayerProxySettings struct {
	Certfile string
	Keyfile  string
}

type SelectionHandlerList map[uint64](func())
type MyPacketListView struct {
	*widgets.QTreeView
	packetRowsByUniqueID    *TwoWayPacketList
	packetRowsBySplitPacket *TwoWayPacketList

	CurrentACKSelection []*gui.QStandardItem
	SelectionHandlers   SelectionHandlerList
	RootNode            *gui.QStandardItem
	PacketIndex         uint64
	StandardModel       *gui.QStandardItemModel

	MSelectionHandlers *sync.Mutex
	MGUI               *sync.Mutex

	IsCapturing       bool
	CaptureJobContext context.Context
	StopCaptureJob    context.CancelFunc
	InjectPacket      chan peer.RakNetPacket

	StudioVersion string
	PlayerVersion string
	DefaultValues DefaultValues

	StudioSettings      *StudioSettings
	PlayerSettings      *PlayerSettings
	ServerSettings      *ServerSettings
	DefaultsSettings    *DefaultsSettings
	PlayerProxySettings *PlayerProxySettings
	Context             *peer.CommunicationContext

	//FilterSettings FilterSettings
}

func NewTwoWayPacketList() *TwoWayPacketList {
	server := &sync.Mutex{}
	client := &sync.Mutex{}
	return &TwoWayPacketList{
		make(PacketList),
		make(PacketList),

		server,
		client,
		sync.NewCond(server),
		sync.NewCond(client),
	}
}

func NewMyPacketListView(parent widgets.QWidget_ITF) *MyPacketListView {
	captureContext, captureCancel := context.WithCancel(context.Background())
	new := &MyPacketListView{
		QTreeView:               widgets.NewQTreeView(parent),
		packetRowsByUniqueID:    NewTwoWayPacketList(),
		packetRowsBySplitPacket: NewTwoWayPacketList(),

		SelectionHandlers: make(SelectionHandlerList),

		MSelectionHandlers: &sync.Mutex{},
		MGUI:               &sync.Mutex{},

		CaptureJobContext: captureContext,
		StopCaptureJob:    captureCancel,

		StudioSettings:      &StudioSettings{},
		PlayerSettings:      &PlayerSettings{},
		ServerSettings:      &ServerSettings{},
		DefaultsSettings:    &DefaultsSettings{},
		PlayerProxySettings: &PlayerProxySettings{},
	}
	return new
}

func (m *MyPacketListView) Reset() {
	m.StandardModel.Clear()
	m.StandardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Direction", "Length in Bytes", "Datagram Numbers", "Ordered Splits", "Total Splits"})

	m.CurrentACKSelection = []*gui.QStandardItem{}
	m.packetRowsByUniqueID = NewTwoWayPacketList()
	m.packetRowsBySplitPacket = NewTwoWayPacketList()
	m.SelectionHandlers = make(SelectionHandlerList)
	m.RootNode = m.StandardModel.InvisibleRootItem()
	m.PacketIndex = 0
	m.CaptureJobContext, m.StopCaptureJob = context.WithCancel(context.Background())
}

func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewQStandardItemF(format string, args ...interface{}) *gui.QStandardItem {
	if format == "%d" {
		ret := gui.NewQStandardItem()
		i, _ := strconv.Atoi(fmt.Sprintf(format, args...)) // hack
		ret.SetData(core.NewQVariant7(i), 0)
		ret.SetEditable(false)
		return ret
	}
	ret := gui.NewQStandardItem2(fmt.Sprintf(format, args...))
	ret.SetEditable(false)
	return ret
}

func NewBasicPacketViewer(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) *widgets.QVBoxLayout {
	tabWidget := NewDefaultPacketViewer(packetType, packet, context, layers)

	layerWidget := widgets.NewQWidget(tabWidget, 0)
	layerLayout := widgets.NewQVBoxLayout()
	layerWidget.SetLayout(layerLayout)
	tabWidget.InsertTab(0, layerWidget, PacketNames[packetType])

	return layerLayout
}

func (m *TwoWayPacketList) Add(index uint32, row []*gui.QStandardItem, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	var mutex *sync.Mutex
	var list PacketList
	var cond *sync.Cond

	if isClient {
		mutex = m.MClient
		list = m.Client
		cond = m.EClient
	} else if isServer {
		mutex = m.MServer
		list = m.Server
		cond = m.EServer
	} else {
		panic(errors.New("add not on server or client: " + packet.Source.String() + " - " + context.Client + " - " + context.Server))
	}
	mutex.Lock()

	list[index] = row
	cond.Broadcast()

	mutex.Unlock()
}

func (m *TwoWayPacketList) Has(index uint32, isClient bool, isServer bool) bool {
	if isClient {
		m.MClient.Lock()
		defer m.MClient.Unlock()
		_, ok := m.Client[index]
		return ok
	} else if isServer {
		m.MServer.Lock()
		defer m.MServer.Unlock()
		_, ok := m.Server[index]
		return ok
	} else {
		panic(errors.New("get not on server or client"))
	}
	return false
}

func (m *TwoWayPacketList) Get(index uint32, isClient bool, isServer bool) []*gui.QStandardItem {
	var rows []*gui.QStandardItem
	var ok bool
	if isClient {
		m.MClient.Lock()
		rows, ok = m.Client[index]
		for !ok {
			m.EClient.Wait()
			rows, ok = m.Client[index]
		}

		m.MClient.Unlock()
	} else if isServer {
		m.MServer.Lock()
		rows, ok = m.Server[index]
		for !ok {
			m.EServer.Wait()
			rows, ok = m.Server[index]
		}

		m.MServer.Unlock()
	} else {
		panic(errors.New("get not on server or client"))
	}

	return rows
}

func paintItems(row []*gui.QStandardItem, color *gui.QColor) {
	for i := 0; i < len(row); i++ {
		row[i].SetBackground(gui.NewQBrush3(color, core.Qt__SolidPattern))
	}
}

func (m *MyPacketListView) highlightByACK(ack peer.ACKRange, isClient bool, isServer bool) {
	var i uint32

	var mutex *sync.Mutex
	var packetList PacketList
	if isClient {
		mutex = m.packetRowsByUniqueID.MClient
		packetList = m.packetRowsByUniqueID.Client
	} else if isServer {
		mutex = m.packetRowsByUniqueID.MServer
		packetList = m.packetRowsByUniqueID.Server
	} else {
		return
	}
	mutex.Lock()

	for i = ack.Min; i <= ack.Max; i++ {
		m.CurrentACKSelection = append(m.CurrentACKSelection, packetList[i]...)
	}
	paintItems(m.CurrentACKSelection, gui.NewQColor3(0, 0, 255, 127))
	mutex.Unlock()
}

func (m *MyPacketListView) clearACKSelection() {
	for _, item := range m.CurrentACKSelection {
		item.SetBackground(gui.NewQBrush2(core.Qt__NoBrush))
	}
}

func (m *MyPacketListView) handleNoneSelected() {
	m.clearACKSelection()
}

func (m *MyPacketListView) registerSplitPacketRow(row []*gui.QStandardItem, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	if layers.Reliability.HasSplitPacket {
		m.packetRowsBySplitPacket.Add(uint32(layers.Reliability.SplitPacketID), row, packet, context, layers)
	}

	m.packetRowsByUniqueID.Add(layers.Reliability.SplitBuffer.UniqueID, row, packet, context, layers)
}

func (m *MyPacketListView) AddSplitPacket(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	if !m.packetRowsByUniqueID.Has(layers.Reliability.SplitBuffer.UniqueID, isClient, isServer) {
		m.AddFullPacket(packetType, packet, context, layers, nil)
		m.BindDefaultCallback(packetType, packet, context, layers)
	} else {
		m.handleSplitPacket(packetType, packet, context, layers)
	}
}

func (m *MyPacketListView) BindCallback(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) {
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	row := m.packetRowsByUniqueID.Get(layers.Reliability.SplitBuffer.UniqueID, isClient, isServer)
	index, _ := strconv.Atoi(row[0].Data(0).ToString())

	m.MSelectionHandlers.Lock()
	m.SelectionHandlers[uint64(index)] = func() {
		m.clearACKSelection()
		if activationCallback != nil && layers.Main != nil {
			activationCallback(packetType, packet, context, layers)
		} else {
			NewDefaultPacketViewer(packetType, packet, context, layers)
		}
	}
	m.MSelectionHandlers.Unlock()
	packetName := PacketNames[packetType]
	if packetName == "" {
		packetName = fmt.Sprintf("0x%02X", packetType)
	}
	row[1].SetData(core.NewQVariant14(packetName), 0)

	for _, item := range row {
		if layers.Main != nil {
			item.SetBackground(gui.NewQBrush2(core.Qt__NoBrush))
		} else {
			paintItems(row, gui.NewQColor3(255, 0, 0, 127))
		}
	}

	if packetType == 0x83 && layers.Main != nil && layers.Reliability != nil && layers.Reliability.SplitBuffer.IsFinal {
		mainLayer := layers.Main.(*peer.Packet83Layer)
		for _, subpacket := range mainLayer.SubPackets {
			if subpacket.Type() == 0x7 && strings.Contains(subpacket.(*peer.Packet83_07).EventName, "Remote") { // highlight events
				paintItems(row, gui.NewQColor3(0, 0, 255, 127))
				break
			}
		}
	}
}

func NewDefaultPacketViewer(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) *widgets.QTabWidget {
	subWindow := widgets.NewQWidget(window, core.Qt__Window)
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

	var datagramNumberLabel *widgets.QLabel
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.SplitBuffer.RakNetPackets
		datagramNumberLabel = NewQLabelF("Datagrams: %d - %d", allRakNetLayers[0].DatagramNumber, allRakNetLayers[len(allRakNetLayers)-1].DatagramNumber)
	} else {
		datagramNumberLabel = NewQLabelF("Datagram: %d", layers.RakNet.DatagramNumber)
	}

	subWindowLayout.AddWidget(datagramNumberLabel, 0, 0)

	tabWidget := widgets.NewQTabWidget(subWindow)
	subWindowLayout.AddWidget(tabWidget, 0, 0)

	logWidget := widgets.NewQWidget(tabWidget, 0)
	logLayout := widgets.NewQVBoxLayout()

	logBox := widgets.NewQTextEdit(logWidget)
	logBox.SetReadOnly(true)
	if layers.Reliability == nil {
		logBox.SetPlainText(packet.GetLog())
	} else {
		logBox.SetPlainText(layers.Reliability.GetLog())
	}
	logLayout.AddWidget(logBox, 0, 0)

	logWidget.SetLayout(logLayout)
	tabWidget.AddTab(logWidget, "Parser log")

	subWindow.SetWindowTitle("Packet Window: " + PacketNames[packetType])
	subWindow.Show()

	if layers.Reliability != nil {
		splitBuffer := layers.Reliability.SplitBuffer
		rakNets := splitBuffer.RakNetPackets
		reliables := splitBuffer.ReliablePackets

		relWidget := widgets.NewQWidget(tabWidget, 0)
		relLayout := widgets.NewQVBoxLayout()

		datagramInfo := new(strings.Builder)
		for _, rakNetLayer := range rakNets {
			fmt.Fprintf(datagramInfo, "%d,", rakNetLayer.DatagramNumber)
		}
		relLayout.AddWidget(NewQLabelF("Datagrams: %s", datagramInfo.String()), 0, 0)

		relLayout.AddWidget(NewQLabelF("Reliability: %d", layers.Reliability.Reliability), 0, 0)
		if layers.Reliability.IsReliable() {
			rmnInfo := new(strings.Builder)
			for _, reliable := range reliables {
				if reliable != nil {
					fmt.Fprintf(rmnInfo, "%d,", reliable.ReliableMessageNumber)
				} else {
					rmnInfo.WriteString("nil,")
				}
			}
			relLayout.AddWidget(NewQLabelF("Reliable MNs: %s", rmnInfo.String()), 0, 0)
		}

		if layers.Reliability.IsOrdered() {
			ordInfo := new(strings.Builder)
			for _, reliable := range reliables {
				if reliable != nil {
					fmt.Fprintf(ordInfo, "%d,", reliable.OrderingIndex)
				} else {
					ordInfo.WriteString("nil,")
				}
			}
			relLayout.AddWidget(NewQLabelF("Ordering channel: %d, indices: %s", layers.Reliability.OrderingChannel, ordInfo.String()), 0, 0)
		}

		if layers.Reliability.IsSequenced() {
			seqInfo := new(strings.Builder)
			for _, reliable := range reliables {
				if reliable != nil {
					fmt.Fprintf(seqInfo, "%d,", reliable.SequencingIndex)
				} else {
					seqInfo.WriteString("nil,")
				}
			}
			relLayout.AddWidget(NewQLabelF("Sequencing indices: %s", layers.Reliability.OrderingChannel, seqInfo.String()), 0, 0)
		}

		relWidget.SetLayout(relLayout)
		tabWidget.AddTab(relWidget, "Reliability Layer Debug")
	}

	return tabWidget
}

func (m *MyPacketListView) BindDefaultCallback(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	row := m.packetRowsByUniqueID.Get(layers.Reliability.SplitBuffer.UniqueID, isClient, isServer)
	index, _ := strconv.Atoi(row[0].Data(0).ToString())

	m.MSelectionHandlers.Lock()
	m.SelectionHandlers[uint64(index)] = func() {
		m.clearACKSelection()
		NewDefaultPacketViewer(packetType, packet, context, layers)
	}
	m.MSelectionHandlers.Unlock()
}

func (m *MyPacketListView) handleSplitPacket(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	row := m.packetRowsBySplitPacket.Get(uint32(layers.Reliability.SplitPacketID), isClient, isServer)
	m.registerSplitPacketRow(row, packet, context, layers)

	reliablePacket := layers.Reliability
	if reliablePacket.SplitBuffer.HasPacketType {
		packetType := reliablePacket.SplitBuffer.PacketType
		packetName := PacketNames[packetType]
		if packetName == "" {
			packetName = fmt.Sprintf("0x%02X", packetType)
		}
		row[1].SetData(core.NewQVariant14(packetName), 0)
	}

	row[3].SetData(core.NewQVariant7(int(layers.Reliability.SplitBuffer.RealLength)), 0)
	if layers.Reliability.SplitBuffer.RakNetPackets[0] == nil {
		panic(errors.New("encountered nil first raknet!!"))
	}
	row[4].SetData(core.NewQVariant14(fmt.Sprintf("%d - %d", layers.Reliability.SplitBuffer.RakNetPackets[0].DatagramNumber, layers.RakNet.DatagramNumber)), 0)
	row[5].SetData(core.NewQVariant14(fmt.Sprintf("%d/%d", layers.Reliability.SplitBuffer.NumReceivedSplits, layers.Reliability.SplitPacketCount)), 0)
	row[6].SetData(core.NewQVariant7(len(layers.Reliability.SplitBuffer.RakNetPackets)), 0)
}

func (m *MyPacketListView) AddFullPacket(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) []*gui.QStandardItem {
	index := atomic.AddUint64(&m.PacketIndex, 1)
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	packetName := PacketNames[packetType]
	if packetName == "" {
		packetName = fmt.Sprintf("0x%02X", packetType)
	}
	indexItem := NewQStandardItemF("%d", index)
	packetTypeItem := NewQStandardItemF(packetName)

	rootRow := []*gui.QStandardItem{indexItem, packetTypeItem}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewQStandardItemF("Client -> Server")
	} else if isServer {
		direction = NewQStandardItemF("Server -> Client")
	} else {
		direction = NewQStandardItemF("Unknown direction")
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
			fmt.Printf("Encountered nil first raknet with %02X\n", packetType)
			firstLayerNumber = -1
		} else {
			firstLayerNumber = int32(firstLayer.DatagramNumber)
		}
		if lastLayer == nil {
			fmt.Printf("Encountered nil last raknet with %02X\n", packetType)
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
		m.registerSplitPacketRow(rootRow, packet, context, layers)
	}

	if layers.Reliability == nil { // Only bind if we're done parsing the packet
		m.MSelectionHandlers.Lock()
		m.SelectionHandlers[index] = func() {
			m.clearACKSelection()
			if activationCallback != nil && layers.Main != nil {
				activationCallback(packetType, packet, context, layers)
			} else {
				NewDefaultPacketViewer(packetType, packet, context, layers)
			}
		}
		m.MSelectionHandlers.Unlock()
	} else {
		paintItems(rootRow, gui.NewQColor3(255, 255, 0, 127))
	}

	m.MGUI.Lock()
	m.RootNode.AppendRow(rootRow)
	m.MGUI.Unlock()

	return rootRow
}

func (m *MyPacketListView) AddACK(ack peer.ACKRange, packet *peer.UDPPacket, context *peer.CommunicationContext, layer *peer.RakNetLayer, activationCallback func()) {
	index := atomic.AddUint64(&m.PacketIndex, 1)
	isClient := context.IsClient(packet.Source)
	isServer := context.IsServer(packet.Source)

	var packetName *gui.QStandardItem
	if ack.Min == ack.Max {
		packetName = NewQStandardItemF("ACK for packet %d", ack.Min)
	} else {
		packetName = NewQStandardItemF("ACK for packets %d - %d", ack.Min, ack.Max)
	}

	indexItem := NewQStandardItemF("%d", index)

	rootRow := []*gui.QStandardItem{indexItem, packetName}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewQStandardItemF("Client -> Server")
	} else if isServer {
		direction = NewQStandardItemF("Server -> Client")
	} else {
		direction = NewQStandardItemF("Unknown direction")
	}

	rootRow = append(rootRow, direction)

	m.MGUI.Lock()
	m.RootNode.AppendRow(rootRow)
	m.MGUI.Unlock()

	m.MSelectionHandlers.Lock()
	m.SelectionHandlers[index] = func() {
		m.clearACKSelection()
		m.highlightByACK(ack, isServer, isClient) // intentionally the other way around
	}
	m.MSelectionHandlers.Unlock()
}

func GUIMain() {
	widgets.NewQApplication(len(os.Args), os.Args)
	window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Roblox PCAP Dissector")

	layout := widgets.NewQVBoxLayout()
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(layout)

	packetViewer := NewMyPacketListView(nil)
	layout.AddWidget(packetViewer, 0, 0)
	window.SetCentralWidget(widget)

	standardModel, proxy := NewFilteringModel(packetViewer)
	proxy.ConnectFilterAcceptsRow(func(sourceRow int, sourceParent *core.QModelIndex) bool {
		return true
	})

	packetViewer.StandardModel = standardModel
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Direction", "Length in Bytes", "Datagram Numbers", "Ordered Splits", "Total Splits"})
	packetViewer.RootNode = standardModel.InvisibleRootItem()
	packetViewer.SetModel(proxy)
	packetViewer.SetSelectionMode(1)
	packetViewer.SetSortingEnabled(true)
	packetViewer.ConnectClicked(func(index *core.QModelIndex) {
		realSelectedValue, _ := strconv.Atoi(standardModel.Item(index.Row(), 0).Data(0).ToString())
		if packetViewer.SelectionHandlers[uint64(realSelectedValue)] == nil {
			packetViewer.handleNoneSelected()
		} else {
			packetViewer.SelectionHandlers[uint64(realSelectedValue)]()
		}
	})

	captureBar := window.MenuBar().AddMenu2("&Capture")
	captureFileAction := captureBar.AddAction("From &file...")
	capture4FileAction := captureBar.AddAction("From &RawCap file...")
	captureLiveAction := captureBar.AddAction("From &live interface...")
	captureInjectAction := captureBar.AddAction("From &injection proxy...")
	captureDivertAction := captureBar.AddAction("From &WinDivert proxy...")
	captureStopAction := captureBar.AddAction("&Stop capture")
	captureFromPlayerProxyAction := captureBar.AddAction("From pl&ayer proxy")
	captureFromPlayerProxyAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}
		NewPlayerProxyWidget(packetViewer, packetViewer.PlayerProxySettings, func(settings *PlayerProxySettings) {
			packetViewer.Reset()

			packetViewer.IsCapturing = true
			context := peer.NewCommunicationContext()
			packetViewer.Context = context

			captureFromPlayerProxy(settings, packetViewer.CaptureJobContext, packetViewer.InjectPacket, packetViewer, packetViewer.Context)
		})
	})

	captureStopAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}
	})

	captureFileAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}
		file := widgets.QFileDialog_GetOpenFileName(window, "Capture from file", "", "PCAP files (*.pcap)", "", 0)
		packetViewer.IsCapturing = true

		context := peer.NewCommunicationContext()
		packetViewer.Context = context

		packetViewer.Reset()

		go func() {
			captureFromFile(file, false, packetViewer.CaptureJobContext, packetViewer, context)
			packetViewer.IsCapturing = false
		}()
	})
	capture4FileAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}
		file := widgets.QFileDialog_GetOpenFileName(window, "Capture from RawCap file", "", "PCAP files (*.pcap)", "", 0)
		packetViewer.IsCapturing = true

		context := peer.NewCommunicationContext()
		packetViewer.Context = context

		packetViewer.Reset()

		go func() {
			captureFromFile(file, true, packetViewer.CaptureJobContext, packetViewer, context)
			packetViewer.IsCapturing = false
		}()
	})
	captureLiveAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}

		NewSelectInterfaceWidget(packetViewer, func(thisItf string, usePromisc bool) {
			packetViewer.IsCapturing = true

			context := peer.NewCommunicationContext()
			packetViewer.Context = context

			packetViewer.Reset()

			go func() {
				captureFromLive(thisItf, false, usePromisc, packetViewer.CaptureJobContext, packetViewer, context)
				packetViewer.IsCapturing = false
			}()
		})
	})
	captureInjectAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}

		NewProxyCaptureWidget(packetViewer, func(src string, dst string) {
			packetViewer.IsCapturing = true

			context := peer.NewCommunicationContext()
			packetViewer.Context = context

			packetViewer.Reset()

			go func() {
				captureFromInjectionProxy(src, dst, packetViewer.CaptureJobContext, packetViewer.InjectPacket, packetViewer, context)
				packetViewer.IsCapturing = false
			}()
		})
	})
	captureDivertAction.ConnectTriggered(func(checked bool) {
		if packetViewer.IsCapturing {
			packetViewer.StopCaptureJob()
		}

		NewPlayerProxyWidget(packetViewer, packetViewer.PlayerProxySettings, func(settings *PlayerProxySettings) {
			packetViewer.Reset()

			packetViewer.IsCapturing = true
			context := peer.NewCommunicationContext()
			packetViewer.Context = context

			autoDetectWinDivertProxy(settings, packetViewer.CaptureJobContext, packetViewer.InjectPacket, packetViewer, packetViewer.Context)
		})
	})

	resp, err := http.Get("http://setup.roblox.com/versionQTStudio")
	if err != nil {
		println("trying to get studio version: " + err.Error())
	} else {
		studioVersion, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			println("trying to read studio version: " + err.Error())
		} else {
			packetViewer.StudioVersion = string(studioVersion)
			potentialLocation := os.Getenv("LOCALAPPDATA") + `/Roblox/Versions/` + packetViewer.StudioVersion + `/RobloxStudioBeta.exe`

			if _, err := os.Stat(potentialLocation); !os.IsNotExist(err) {
				packetViewer.StudioSettings.Location = potentialLocation
			}
		}
		resp.Body.Close()
	}

	resp, err = http.Get("http://setup.roblox.com/version")
	if err != nil {
		println("trying to get player version: " + err.Error())
	} else {
		playerVersion, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("trying to read player version: " + err.Error())
		} else {
			packetViewer.PlayerVersion = string(playerVersion)
			potentialLocation := os.Getenv("LOCALAPPDATA") + `/Roblox/Versions/` + packetViewer.PlayerVersion + `/RobloxPlayerBeta.exe`
			if _, err := os.Stat(potentialLocation); !os.IsNotExist(err) {
				packetViewer.PlayerSettings.Location = potentialLocation
			}
		}
		resp.Body.Close()
	}

	packetViewer.StudioSettings.Flags = ``
	packetViewer.StudioSettings.Port = "53640"
	packetViewer.PlayerSettings.Flags = `--play -a https://www.roblox.com/Login/Negotiate.ashx --launchtime=1503226579241`
	packetViewer.PlayerSettings.AuthTicket = `Guest%3A-306579839`
	packetViewer.PlayerSettings.TrackerID = "11076148732"

	manageRobloxBar := window.MenuBar().AddMenu2("Start &Roblox")
	startServerAction := manageRobloxBar.AddAction("Start &local server...")
	startClientAction := manageRobloxBar.AddAction("Start local &client...")
	startPlayerAction := manageRobloxBar.AddAction("Start Roblox &Player...")
	startServerAction.ConnectTriggered(func(checked bool) {
		NewStudioChooser(packetViewer, packetViewer.StudioSettings, func(settings *StudioSettings) {
			packetViewer.StudioSettings = settings

			flags := []string{}
			flags = append(flags, "-task", "StartServer")
			flags = append(flags, "-port", settings.Port, "-creatorId", "0", "-creatorType", "0", "-placeVersion")
			err = exec.Command(settings.Location, flags...).Start()
			if err != nil {
				println("while starting process:", err.Error())
			}
		})
	})
	startClientAction.ConnectTriggered(func(checked bool) {
		NewStudioChooser(packetViewer, packetViewer.StudioSettings, func(settings *StudioSettings) {
			packetViewer.StudioSettings = settings

			flags := []string{}
			flags = append(flags, "-task", "StartClient")
			flags = append(flags, "-port", settings.Port)
			err = exec.Command(settings.Location, flags...).Start()
			if err != nil {
				println("while starting process:", err.Error())
			}
		})
	})
	startPlayerAction.ConnectTriggered(func(checked bool) {
		NewPlayerChooser(packetViewer, packetViewer.PlayerSettings, func(settings *PlayerSettings) {
			packetViewer.PlayerSettings = settings
			placeID, err := strconv.Atoi(settings.GameID)
			if err != nil {
				println("while converting place id:", err.Error())
				return
			}

			flags := []string{}
			joinScript := fmt.Sprintf(`https://assetgame.roblox.com/game/PlaceLauncher.ashx?request=RequestGame&browserTrackerId=%s&placeId=%d&isPartyLeader=false&genderId=2`, settings.TrackerID, placeID)
			flags = append(flags, strings.Split(settings.Flags, " ")...)
			flags = append(flags, "-j", joinScript)
			flags = append(flags, "-b", settings.TrackerID)
			flags = append(flags, "-t", settings.AuthTicket)
			err = exec.Command(settings.Location, flags...).Start()
			if err != nil {
				println("while starting process:", err.Error())
			}
		})
	})

	toolsBar := window.MenuBar().AddMenu2("&Tools")

	scriptDumperAction := toolsBar.AddAction("Dump &scripts")
	scriptDumperAction.ConnectTriggered(func(checked bool) {
		dumpScripts(packetViewer.Context.DataModel.Instances, 0)
		scriptData, err := os.OpenFile("dumps/scriptKeys", os.O_RDWR|os.O_CREATE, 0666)
		defer scriptData.Close()
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}

		_, err = fmt.Fprintf(scriptData, "Int 1: %d\nInt 2: %d", packetViewer.Context.Int1, packetViewer.Context.Int2)
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}
	})
	dumperAction := toolsBar.AddAction("&DataModel dumper lite...")
	dumperAction.ConnectTriggered(func(checked bool) {
		location := widgets.QFileDialog_GetSaveFileName(packetViewer, "Save as RBXL...", "", "Roblox place files (*.rbxl)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		stripInvalidTypes(packetViewer.Context.DataModel.Instances, packetViewer.DefaultValues, 0)

		err = bin.SerializePlace(writer, nil, packetViewer.Context.DataModel)
		if err != nil {
			println("while serializing place:", err.Error())
			return
		}
	})
	browseAction := toolsBar.AddAction("&Browse DataModel...")
	browseAction.ConnectTriggered(func(checked bool) {
		if packetViewer.Context != nil {
			NewDataModelBrowser(packetViewer.Context, packetViewer.Context.DataModel, packetViewer.DefaultValues)
		}
	})

	readDefaults := toolsBar.AddAction("Parse &default values...")
	readDefaults.ConnectTriggered(func(checked bool) {
		NewFindDefaultsWidget(window, packetViewer.DefaultsSettings, func(settings *DefaultsSettings) {
			packetViewer.DefaultValues = ParseDefaultValues(settings.Files)
		})
	})

	viewCache := toolsBar.AddAction("&View string cache...")
	viewCache.ConnectTriggered(func(checked bool) {
		NewViewCacheWidget(packetViewer, packetViewer.Context)
	})

	injectChat := toolsBar.AddAction("Inject &chat message...")
	injectChat.ConnectTriggered(func(checked bool) {
		if packetViewer.Context == nil {
			println("context is nil!")
			return
		} else if packetViewer.Context.DataModel == nil {
			println("datamodel instances is nil!")
			return
		}

		dataModel := packetViewer.Context.DataModel.Instances
		var players, replicatedStorage *rbxfile.Instance
		for i := 0; i < len(dataModel); i++ {
			if dataModel[i].ClassName == "Players" {
				players = dataModel[i]
			} else if dataModel[i].ClassName == "ReplicatedStorage" {
				replicatedStorage = dataModel[i]
			}
		}
		player := players.Children[0]
		println("chose player", player.Name())
		chatEvent := replicatedStorage.FindFirstChild("DefaultChatSystemChatEvents", false).FindFirstChild("SayMessageRequest", false)
		subpacket := &peer.Packet83_07{
			Instance:  chatEvent,
			EventName: "OnServerEvent",
			Event: &peer.ReplicationEvent{
				Arguments: []rbxfile.Value{
					rbxfile.ValueReference{Instance: player},
					rbxfile.ValueTuple{
						rbxfile.ValueString("Hello, this is a hacked message"),
						rbxfile.ValueString("All"),
					},
				},
			},
		}

		packetViewer.InjectPacket <- &peer.Packet83Layer{
			SubPackets: []peer.Packet83Subpacket{subpacket},
		}
	})

	peersBar := window.MenuBar().AddMenu2("&Peers...")
	//startSelfServer := peersBar.AddAction("Start self &server...")
	startSelfClient := peersBar.AddAction("Start self &client...")
	/*startSelfServer.ConnectTriggered(func(checked bool)() {
	        NewServerStartWidget(window, packetViewer.ServerSettings, func(settings *ServerSettings) {
	            port, _ := strconv.Atoi(settings.Port)
	            enums, err := os.Open(settings.EnumSchemaLocation)
	            if err != nil {
	                println("while parsing schema:", err.Error())
	                return
	            }
	            instances, err := os.Open(settings.InstanceSchemaLocation)
	            if err != nil {
	                println("while parsing schema:", err.Error())
	                return
	            }
	            schema, err := peer.ParseSchema(instances, enums)
	            if err != nil {
	                println("while parsing schema:", err.Error())
	                return
	            }
	            dictfile, err := os.Open(settings.DictionaryLocation)
	            if err != nil {
	                println("while parsing dict:", err.Error())
	                return
	            }
	            var dictionaries peer.Packet82Layer
	            err = gob.NewDecoder(dictfile).Decode(&dictionaries)
	            if err != nil {
	                println("while parsing dict:", err.Error())
	                return
	            }

	            go peer.StartServer(uint16(port), &dictionaries, &schema)
	        })
		})*/
	startSelfClient.ConnectTriggered(func(checked bool) {
		customClient := peer.NewCustomClient()
		NewClientStartWidget(window, customClient, func(placeId uint32, isGuest bool, ticket string) {
			NewClientConsole(window, customClient)
			customClient.SecuritySettings.InitWin10()
			if isGuest {
				go customClient.ConnectGuest(placeId, 2)
			} else {
				go customClient.ConnectWithAuthTicket(placeId, ticket)
			}
		})
	})

	window.Show()

	widgets.QApplication_Exec()
}
