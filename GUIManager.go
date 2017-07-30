package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/google/gopacket"
import "os"
import "fmt"
import "strconv"
import "sync/atomic"
import "sync"

var window *widgets.QMainWindow

type PacketList map[uint32]([]*gui.QStandardItem)
type TwoWayPacketList struct {
	Server PacketList
	Client PacketList
	MServer *sync.Mutex
	MClient *sync.Mutex
}

type SelectionHandlerList map[uint64](func ())
type MyPacketListView struct {
	*widgets.QTreeView
	packetRowsByUniqueID *TwoWayPacketList
	packetRowsBySplitPacket *TwoWayPacketList
	
	CurrentACKSelection []*gui.QStandardItem
	SelectionHandlers SelectionHandlerList
	RootNode *gui.QStandardItem
	PacketIndex uint64

	MSelectionHandlers *sync.Mutex
	MGUI *sync.Mutex
}

func NewTwoWayPacketList() *TwoWayPacketList {
	return &TwoWayPacketList{
		make(PacketList),
		make(PacketList),

		&sync.Mutex{},
		&sync.Mutex{},
	}
}

func NewMyPacketListView(parent widgets.QWidget_ITF) *MyPacketListView {
	new := &MyPacketListView{
		widgets.NewQTreeView(parent),
		NewTwoWayPacketList(),
		NewTwoWayPacketList(),
		nil,
		make(SelectionHandlerList),
		nil,
		0,

		&sync.Mutex{},
		&sync.Mutex{},
	}
	return new
}

func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewQStandardItemF(format string, args ...interface{}) *gui.QStandardItem {
	if format == "%d" {

		ret := gui.NewQStandardItem()
		i, _ := strconv.Atoi(fmt.Sprintf(format, args...)) // hack
		ret.SetData(core.NewQVariant7(i), 0)
		return ret
	}
	return gui.NewQStandardItem2(fmt.Sprintf(format, args...))
}

func NewBasicPacketViewer(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) *widgets.QVBoxLayout {
	subWindow := widgets.NewQWidget(window, core.Qt__Window)
	subWindowLayout := widgets.NewQVBoxLayout2(subWindow)

	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	var direction string
	if isClient {
		direction = "Direction: Client -> Server"
	} else if isServer {
		direction = "Direction: Server -> Client"
	} else {
		direction = "Direction: Unknown"
	}
	directionLabel := widgets.NewQLabel2(direction, nil, 0)
	lengthLabel := NewQLabelF("Length: %d", len(packet.ApplicationLayer().Payload()))
	subWindowLayout.AddWidget(directionLabel, 0, 0)
	subWindowLayout.AddWidget(lengthLabel, 0, 0)

	var datagramNumberLabel *widgets.QLabel
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.AllRakNetLayers
		datagramNumberLabel = NewQLabelF("Datagrams: %d - %d", allRakNetLayers[0].DatagramNumber, allRakNetLayers[len(allRakNetLayers) - 1].DatagramNumber)
	} else {
		datagramNumberLabel = NewQLabelF("Datagram: %d", layers.RakNet.DatagramNumber)
	}

	subWindowLayout.AddWidget(datagramNumberLabel, 0, 0)

	tabWidget := widgets.NewQTabWidget(nil)
	subWindowLayout.AddWidget(tabWidget, 0, 0)
	//packetDataLayout := widgets.NewQVBoxLayout()
	//packetDataWidget := widgets.NewQWidget(nil, 0)
	//packetDataWidget.SetLayout(packetDataLayout)

	//monospaceFont := gui.NewQFont2("monospace", 10, 50, false)

	//labelForPacketContents := widgets.NewQLabel2("Packet inner layer contents:", nil, 0)
	//packetContents := widgets.NewQPlainTextEdit2(hex.Dump(data), nil)
	//packetDataLayout.AddWidget(labelForPacketContents, 0, 0)
	//packetDataLayout.AddWidget(packetContents, 0, 0)
	//packetContents.SetReadOnly(true)
	//packetContents.Document().SetDefaultFont(monospaceFont)
	//packetContents.SetLineWrapMode(0)

	//tabWidget.AddTab(packetDataWidget, "Raw data")

	subWindow.SetWindowTitle("Packet Window: " + PacketNames[packetType])
	subWindow.Show()

	layerWidget := widgets.NewQWidget(nil, 0)
	layerLayout := widgets.NewQVBoxLayout()
	layerWidget.SetLayout(layerLayout)

	tabWidget.AddTab(layerWidget, PacketNames[packetType])

	return layerLayout
}

func (m *TwoWayPacketList) Add(index uint32, row []*gui.QStandardItem, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	var mutex *sync.Mutex
	var list PacketList

	if isClient {
		mutex = m.MClient
		list = m.Client
	} else if isServer {
		mutex = m.MServer
		list = m.Server
	} else {
		return
	}
	mutex.Lock()

	list[index] = row

	mutex.Unlock()
}

func (m *TwoWayPacketList) Get(index uint32, isClient bool, isServer bool) []*gui.QStandardItem {
	var rows []*gui.QStandardItem
	if isClient {
		m.MClient.Lock()
		rows = m.Client[index]
		m.MClient.Unlock()
	} else if isServer {
		m.MServer.Lock()
		rows = m.Server[index]
		m.MServer.Unlock()
	}
	return rows
}

func (m *MyPacketListView) highlightByACK(ack ACKRange, isClient bool, isServer bool) {
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
		for j := 0; j < len(packetList[i]); j++ {
			packetList[i][j].SetBackground(gui.NewQBrush3(gui.NewQColor3(0, 0, 255, 127), core.Qt__SolidPattern))
			m.CurrentACKSelection = append(m.CurrentACKSelection, packetList[i][j])
		}
	}
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

func (m *MyPacketListView) registerSplitPacketRow(row []*gui.QStandardItem, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	if layers.Reliability.HasSplitPacket {
		m.packetRowsBySplitPacket.Add(uint32(layers.Reliability.SplitPacketID), row, packet, context, layers)
	}

	m.packetRowsByUniqueID.Add(layers.Reliability.UniqueID, row, packet, context, layers)
}

func (m *MyPacketListView) AddSplitPacket(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	if layers.Reliability.IsFirst {
		m.AddFullPacket(packetType, packet, context, layers, nil)
	} else {
		m.handleSplitPacket(packetType, packet, context, layers)
	}
}

func (m *MyPacketListView) BindCallback(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers, activationCallback ActivationCallback) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	row := m.packetRowsByUniqueID.Get(layers.Reliability.UniqueID, isClient, isServer)
	index, _ := strconv.Atoi(row[0].Data(0).ToString())

	m.MSelectionHandlers.Lock()
	m.SelectionHandlers[uint64(index)] = func () {
		m.clearACKSelection()
		if activationCallback != nil {
			activationCallback(packetType, packet, context, layers)
		}
	}
	m.MSelectionHandlers.Unlock()
}

func (m *MyPacketListView) handleSplitPacket(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	row := m.packetRowsBySplitPacket.Get(uint32(layers.Reliability.SplitPacketID), isClient, isServer)
	m.registerSplitPacketRow(row, packet, context, layers)

	reliablePacket := layers.Reliability
	if reliablePacket.HasPacketType {
		row[1].SetData(core.NewQVariant14(PacketNames[reliablePacket.PacketType]), 0)
	}

	row[3].SetData(core.NewQVariant7(int(layers.Reliability.RealLength)), 0)
	row[4].SetData(core.NewQVariant14(fmt.Sprintf("%d - %d", layers.Reliability.AllRakNetLayers[0].DatagramNumber, layers.RakNet.DatagramNumber)), 0)
	row[5].SetData(core.NewQVariant14(fmt.Sprintf("%d/%d", layers.Reliability.NumReceivedSplits, layers.Reliability.SplitPacketCount)), 0)
}

func (m *MyPacketListView) AddFullPacket(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers, activationCallback ActivationCallback) []*gui.QStandardItem {
	index := atomic.AddUint64(&m.PacketIndex, 1)
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	packetName := PacketNames[packetType]
	if packetName == "" {
		packetName = fmt.Sprintf("0x%02X", packetType)
	}
	indexItem := NewQStandardItemF("%d", index)
	packetTypeItem := NewQStandardItemF(packetName)
	indexItem.SetEditable(false)
	packetTypeItem.SetEditable(false)

	rootRow := []*gui.QStandardItem{indexItem, packetTypeItem}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewQStandardItemF("Client -> Server")
	} else if isServer {
		direction = NewQStandardItemF("Server -> Client")
	} else {
		direction = NewQStandardItemF("Unknown direction")
	}

	direction.SetEditable(false)
	rootRow = append(rootRow, direction)

	var length *gui.QStandardItem
	if layers.Reliability != nil {
		length = NewQStandardItemF("%d", layers.Reliability.LengthInBits / 8)
	} else {
		length = NewQStandardItemF("%d", len(packet.ApplicationLayer().Payload()))
	}
	length.SetEditable(false)
	rootRow = append(rootRow, length)
	var datagramNumber *gui.QStandardItem
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.AllRakNetLayers

		firstLayer := allRakNetLayers[0]
		lastLayer := allRakNetLayers[len(allRakNetLayers) - 1]
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
	datagramNumber.SetEditable(false)
	rootRow = append(rootRow, datagramNumber)

	if layers.Reliability != nil {
		receivedSplits := NewQStandardItemF("%d/%d", layers.Reliability.NumReceivedSplits, layers.Reliability.SplitPacketCount)
		receivedSplits.SetEditable(false)
		rootRow = append(rootRow, receivedSplits)
	} else {
		rootRow = append(rootRow, nil)
	}

	if layers.Reliability != nil {
		m.registerSplitPacketRow(rootRow, packet, context, layers)
	}

	if layers.Reliability == nil || layers.Reliability.IsFinal { // Only bind if we're done parsing the packet
		m.MSelectionHandlers.Lock()
		m.SelectionHandlers[index] = func () {
			m.clearACKSelection()
			if activationCallback != nil {
				activationCallback(packetType, packet, context, layers)
			}
		}
		m.MSelectionHandlers.Unlock()
	}

	m.MGUI.Lock()
	m.RootNode.AppendRow(rootRow)
	m.MGUI.Unlock()

	return rootRow
}

func (m *MyPacketListView) AddACK(ack ACKRange, packet gopacket.Packet, context *CommunicationContext, layer *RakNetLayer, activationCallback func()) {
	index := atomic.AddUint64(&m.PacketIndex, 1)
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	var packetName *gui.QStandardItem
	if ack.Min == ack.Max {
		packetName = NewQStandardItemF("ACK for packet %d", ack.Min)
	} else {
		packetName = NewQStandardItemF("ACK for packets %d - %d", ack.Min, ack.Max)
	}

	indexItem := NewQStandardItemF("%d", index)
	indexItem.SetEditable(false)
	packetName.SetEditable(false)

	rootRow := []*gui.QStandardItem{indexItem, packetName}

	var direction *gui.QStandardItem
	if isClient {
		direction = NewQStandardItemF("Client -> Server")
	} else if isServer {
		direction = NewQStandardItemF("Server -> Client")
	} else {
		direction = NewQStandardItemF("Unknown direction")
	}

	direction.SetEditable(false)
	rootRow = append(rootRow, direction)

	m.MGUI.Lock()
	m.RootNode.AppendRow(rootRow)
	m.MGUI.Unlock()

	m.MSelectionHandlers.Lock()
	m.SelectionHandlers[index] = func () {
		m.clearACKSelection()
		m.highlightByACK(ack, isServer, isClient) // intentionally the other way around
	}
	m.MSelectionHandlers.Unlock()

}

func GUIMain(done chan bool, viewerChan chan *MyPacketListView) {
	widgets.NewQApplication(len(os.Args), os.Args)
	window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Roblox PCAP Dissector")

	layout := widgets.NewQVBoxLayout()
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(layout)

	packetViewer := NewMyPacketListView(nil)
	layout.AddWidget(packetViewer, 0, 0)
	window.SetCentralWidget(widget)

	standardModel := NewProperSortModel(packetViewer)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Direction", "Length in Bytes", "Datagram Numbers", "Received Splits"})
	packetViewer.RootNode = standardModel.InvisibleRootItem()
	packetViewer.SetModel(standardModel)
	packetViewer.SetSelectionMode(1)
	packetViewer.SetSortingEnabled(true)
	packetViewer.ConnectSelectionChanged(func (selected *core.QItemSelection, deselected *core.QItemSelection) {
		if len(selected.Indexes()) == 0 {
			packetViewer.handleNoneSelected()
		}
		realSelectedValue, _ := strconv.Atoi(standardModel.Item(selected.Indexes()[0].Row(), 0).Data(0).ToString())
		if packetViewer.SelectionHandlers[uint64(realSelectedValue)] == nil {
			packetViewer.handleNoneSelected()
		} else {
			packetViewer.SelectionHandlers[uint64(realSelectedValue)]()
		}
	})

	window.Show()

	viewerChan <- packetViewer

	widgets.QApplication_Exec()
	done <- true
}
