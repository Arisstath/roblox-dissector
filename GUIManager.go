package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/google/gopacket"
import "os"
import "strconv"
import "fmt"
import "encoding/hex"

var window *widgets.QMainWindow

type PacketList map[uint32]([]*gui.QStandardItem)
type SelectionHandlerList map[int](func ())
type MyPacketListView struct {
	*widgets.QTreeView
	ServerPackets PacketList
	ClientPackets PacketList
	CurrentACKSelection []*gui.QStandardItem
	SelectionHandlers SelectionHandlerList
	RootNode *gui.QStandardItem
	PacketIndex uint64
}

func NewMyPacketListView(parent widgets.QWidget_ITF) *MyPacketListView {
	new := &MyPacketListView{
		widgets.NewQTreeView(parent),
		make(PacketList),
		make(PacketList),
		nil,
		make(SelectionHandlerList),
		nil,
		0,
	}
	return new
}

func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewBasicPacketViewer(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) *widgets.QVBoxLayout {
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
	lengthLabel := NewQLabelF("Length: %d", len(data))
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
	packetDataLayout := widgets.NewQVBoxLayout()
	packetDataWidget := widgets.NewQWidget(nil, 0)
	packetDataWidget.SetLayout(packetDataLayout)

	monospaceFont := gui.NewQFont2("monospace", 10, 50, false)

	labelForPacketContents := widgets.NewQLabel2("Packet inner layer contents:", nil, 0)
	packetContents := widgets.NewQPlainTextEdit2(hex.Dump(data), nil)
	packetDataLayout.AddWidget(labelForPacketContents, 0, 0)
	packetDataLayout.AddWidget(packetContents, 0, 0)
	packetContents.SetReadOnly(true)
	packetContents.Document().SetDefaultFont(monospaceFont)
	packetContents.SetLineWrapMode(0)

	tabWidget.AddTab(packetDataWidget, "Raw data")

	subWindow.SetWindowTitle("Packet Window: " + PacketNames[data[0]])
	subWindow.Show()

	layerWidget := widgets.NewQWidget(nil, 0)
	layerLayout := widgets.NewQVBoxLayout()
	layerWidget.SetLayout(layerLayout)

	tabWidget.AddTab(layerWidget, PacketNames[data[0]])

	return layerLayout
}

func (m PacketList) Add(item *gui.QStandardItem, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	rakNetLayer := layers.RakNet
	reliabilityLayer := layers.Reliability

	if reliabilityLayer == nil || !reliabilityLayer.HasSplitPacket {
		m[rakNetLayer.DatagramNumber] = append(m[rakNetLayer.DatagramNumber], item)
	} else {
		for _, layer := range reliabilityLayer.AllRakNetLayers {
			m[layer.DatagramNumber] = append(m[layer.DatagramNumber], item)
		}
	}
}

func (m *MyPacketListView) AddToPacketList(item *gui.QStandardItem, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	if isClient {
		m.ClientPackets.Add(item, packet, context, layers)
	} else if isServer {
		m.ServerPackets.Add(item, packet, context, layers)
	}
}

func (m *MyPacketListView) HighlightByACK(ack ACKRange, isClient bool, isServer bool) {
	var i uint32

	var packetList PacketList
	if isClient {
		packetList = m.ClientPackets
	} else if isServer {
		packetList = m.ServerPackets
	} else {
		return
	}

	for i = ack.Min; i <= ack.Max; i++ {
		for j := 0; j < len(packetList[i]); j++ {
			packetList[i][j].SetBackground(gui.NewQBrush3(gui.NewQColor3(0, 0, 255, 127), core.Qt__SolidPattern))
			m.CurrentACKSelection = append(m.CurrentACKSelection, packetList[i][j])
		}
	}
}

func (m *MyPacketListView) ClearACKSelection() {
	for _, item := range m.CurrentACKSelection {
		item.SetBackground(gui.NewQBrush2(core.Qt__NoBrush))
	}
}

func (m *MyPacketListView) HandleNoneSelected() {
	m.ClearACKSelection()
}

func (m *MyPacketListView) Add(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers, activationCallback ActivationCallback) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	packetName := PacketNames[data[0]]
	if packetName == "" {
		packetName = fmt.Sprintf("0x%02X", data[0])
	}
	indexItem := gui.NewQStandardItem2(strconv.Itoa(int(m.PacketIndex)))
	packetType := gui.NewQStandardItem2(packetName)
	indexItem.SetEditable(false)
	packetType.SetEditable(false)

	m.PacketIndex++

	rootRow := []*gui.QStandardItem{indexItem, packetType}

	var direction *gui.QStandardItem
	if isClient {
		direction = gui.NewQStandardItem2("Client -> Server")
	} else if isServer {
		direction = gui.NewQStandardItem2("Server -> Client")
	} else {
		direction = gui.NewQStandardItem2("Unknown direction")
	}

	direction.SetEditable(false)
	rootRow = append(rootRow, direction)

	length := gui.NewQStandardItem2(strconv.Itoa(len(data)))
	length.SetEditable(false)
	rootRow = append(rootRow, length)
	datagramNumberString := strconv.Itoa(int(layers.RakNet.DatagramNumber))
	if layers.Reliability != nil && layers.Reliability.HasSplitPacket {
		allRakNetLayers := layers.Reliability.AllRakNetLayers
		datagramNumberString = strconv.Itoa(int(allRakNetLayers[0].DatagramNumber)) + " - " + strconv.Itoa(int(allRakNetLayers[len(allRakNetLayers) - 1].DatagramNumber))
	}
	datagramNumber := gui.NewQStandardItem2(datagramNumberString)
	datagramNumber.SetEditable(false)
	rootRow = append(rootRow, datagramNumber)

	if layers.Reliability != nil {
		m.AddToPacketList(packetType, packet, context, layers)
	}

	m.RootNode.AppendRow(rootRow)

	m.SelectionHandlers[packetType.Row()] = func () {
		m.ClearACKSelection()
		if activationCallback != nil {
			activationCallback(data, packet, context, layers)
		}
	}
}

func (m *MyPacketListView) AddACK(ack ACKRange, packet gopacket.Packet, context *CommunicationContext, layer *RakNetLayer, activationCallback func()) {
	isClient := SourceInterfaceFromPacket(packet) == context.GetClient()
	isServer := SourceInterfaceFromPacket(packet) == context.GetServer()

	var rangeString string
	if ack.Min == ack.Max {
		rangeString = strconv.Itoa(int(ack.Min))
	} else {
		rangeString = strconv.Itoa(int(ack.Min)) + "-" + strconv.Itoa(int(ack.Max))
	}

	indexItem := gui.NewQStandardItem2(strconv.Itoa(int(m.PacketIndex)))
	packetName := gui.NewQStandardItem2("ACK for packets " + rangeString)
	indexItem.SetEditable(false)
	packetName.SetEditable(false)

	m.PacketIndex++

	rootRow := []*gui.QStandardItem{indexItem, packetName}

	var direction *gui.QStandardItem
	if isClient {
		direction = gui.NewQStandardItem2("Client -> Server")
	} else if isServer {
		direction = gui.NewQStandardItem2("Server -> Client")
	} else {
		direction = gui.NewQStandardItem2("Unknown direction")
	}

	direction.SetEditable(false)
	rootRow = append(rootRow, direction)

	m.RootNode.AppendRow(rootRow)

	m.SelectionHandlers[packetName.Row()] = func () {
		m.ClearACKSelection()
		m.HighlightByACK(ack, isServer, isClient) // intentionally the other way around
	}
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

	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Type", "Direction", "Length in Bytes", "Datagram Numbers"})
	packetViewer.RootNode = standardModel.InvisibleRootItem()
	packetViewer.SetModel(standardModel)
	packetViewer.SetSelectionMode(1)
	packetViewer.SetSortingEnabled(true)
	packetViewer.ConnectSelectionChanged(func (selected *core.QItemSelection, deselected *core.QItemSelection) {
		if len(selected.Indexes()) == 0 {
			packetViewer.HandleNoneSelected()
		}
		realSelectedValue := selected.Indexes()[0].Row()
		if packetViewer.SelectionHandlers[realSelectedValue] == nil {
			packetViewer.HandleNoneSelected()
		} else {
			packetViewer.SelectionHandlers[realSelectedValue]()
		}
	})

	window.Show()

	viewerChan <- packetViewer

	widgets.QApplication_Exec()
	done <- true
}
