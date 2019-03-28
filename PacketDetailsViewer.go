package main

import (
	"fmt"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type PacketDetailsViewer struct {
	*widgets.QWidget
	TabLayout      *widgets.QTabWidget
	LogBox         *widgets.QTextEdit
	ReliabilityTab *widgets.QWidget
	MainTab        *widgets.QWidget
}

func NewPacketDetailsViewer(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *PacketDetailsViewer {
	basicWidget := widgets.NewQWidget(parent, flags)
	layout := widgets.NewQVBoxLayout2(basicWidget)
	layout.SetAlign(core.Qt__AlignTop)

	detailsViewer := &PacketDetailsViewer{QWidget: basicWidget}

	mainScrollArea := widgets.NewQScrollArea(detailsViewer)
	mainScrollArea.SetWidgetResizable(true)
	tabWidget := widgets.NewQTabWidget(basicWidget)
	detailsViewer.TabLayout = tabWidget

	logWidget := widgets.NewQWidget(tabWidget, 0)
	logLayout := NewTopAlignLayout()
	logBox := widgets.NewQTextEdit(logWidget)
	logBox.SetReadOnly(true)
	logLayout.AddWidget(logBox, 0, 0)
	logWidget.SetLayout(logLayout)
	tabWidget.AddTab(logWidget, "Parser log")
	detailsViewer.LogBox = logBox

	relWidget := widgets.NewQWidget(tabWidget, 0)
	relLayout := NewTopAlignLayout()
	relLayout.AddWidget(NewLabel("No ReliabilityLayer selected!"), 0, 0)
	relWidget.SetLayout(relLayout)
	tabWidget.AddTab(relWidget, "Reliability Layer")
	detailsViewer.ReliabilityTab = relWidget

	mainWidget := widgets.NewQWidget(tabWidget, 0)
	mainLayout := NewTopAlignLayout()
	mainLayout.AddWidget(NewLabel("No packets selected!"), 0, 0)
	mainWidget.SetLayout(mainLayout)
	mainTabIndex := tabWidget.AddTab(mainWidget, "Main Layer")
	tabWidget.SetCurrentIndex(mainTabIndex)
	detailsViewer.MainTab = mainWidget

	mainScrollArea.SetWidget(tabWidget)
	layout.AddWidget(mainScrollArea, 0, 0)

	return detailsViewer
}

func (viewer *PacketDetailsViewer) Update(context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) {
	if layers.Reliability != nil {
		viewer.LogBox.SetPlainText(layers.Reliability.GetLog())
	}
	if layers.Error != nil {
		viewer.LogBox.SetPlainText(viewer.LogBox.ToPlainText() + "\nError: " + layers.Error.Error())
	}

	originalIndex := viewer.TabLayout.CurrentIndex()

	// TODO: improve this layout
	// We must destroy the entire widget here, because
	// AddWidget() on the layout will parent child widgets
	// to the QWidget
	reliabilityIndex := viewer.TabLayout.IndexOf(viewer.ReliabilityTab)
	viewer.ReliabilityTab.DestroyQWidget()
	viewer.ReliabilityTab = widgets.NewQWidget(viewer, 0)
	if layers.Reliability != nil {
		relLayout := NewTopAlignLayout()
		splitBuffer := layers.Reliability.SplitBuffer
		rakNets := splitBuffer.RakNetPackets
		reliables := splitBuffer.ReliablePackets

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

		viewer.ReliabilityTab.SetLayout(relLayout)
	} else {
		newRelLayout := NewTopAlignLayout()
		newRelLayout.AddWidget(NewLabel("No ReliabilityLayer selected!"), 0, 0)
		viewer.ReliabilityTab.SetLayout(newRelLayout)
	}
	viewer.TabLayout.InsertTab(reliabilityIndex, viewer.ReliabilityTab, "Reliability Layer")

	mainIndex := viewer.TabLayout.IndexOf(viewer.MainTab)
	viewer.MainTab.DestroyQWidget()
	viewer.MainTab = widgets.NewQWidget(viewer, 0)
	if activationCallback != nil && layers.Main != nil {
		newMainLayout := NewTopAlignLayout()
		activationCallback(newMainLayout, context, layers)
		viewer.MainTab.SetLayout(newMainLayout)
	} else {
		newMainLayout := NewTopAlignLayout()
		newMainLayout.AddWidget(NewLabel("No main layer selected!"), 0, 0)
		viewer.MainTab.SetLayout(newMainLayout)
	}
	viewer.TabLayout.InsertTab(mainIndex, viewer.MainTab, "Main Layer")

	viewer.TabLayout.SetCurrentIndex(originalIndex)
}

func NewPacketViewerMenu(parent widgets.QWidget_ITF, context *peer.CommunicationContext, layers *peer.PacketLayers, activationCallback ActivationCallback) *widgets.QMenu {
	menu := widgets.NewQMenu(parent)
	showPacketAction := menu.AddAction("View in new window")
	showPacketAction.ConnectTriggered(func(_ bool) {
		window := NewPacketDetailsViewer(parent, core.Qt__Window)
		window.Update(context, layers, activationCallback)
		window.SetWindowTitle(fmt.Sprintf("Packet window: %s", layers.String()))
		window.Show()
	})

	return menu
}
