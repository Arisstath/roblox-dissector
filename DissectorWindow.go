package main

import (
	"context"
	"errors"
	"path"
	"time"

	"github.com/dreadl0ck/gopcap"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

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
	RBXLLocation           string
}

type DefaultsSettings struct {
	Files []string
}

type PlayerProxySettings struct {
	Certfile string
	Keyfile  string
}

type DissectorWindow struct {
	*widgets.QMainWindow

	TabWidget         *widgets.QTabWidget
	PacketListViewers []*PacketListViewer

	StudioVersion string
	PlayerVersion string
	DefaultValues DefaultValues

	StudioSettings      *StudioSettings
	PlayerSettings      *PlayerSettings
	ServerSettings      *ServerSettings
	DefaultsSettings    *DefaultsSettings
	PlayerProxySettings *PlayerProxySettings
}

func (window *DissectorWindow) ShowCaptureError(err error) {
	widgets.QMessageBox_Critical(window, "Capture Error", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
}

// TODO: Clean up the emitter connections
func (window *DissectorWindow) AddConversation(name string, conv *Conversation) *PacketListViewer {
	var viewer *PacketListViewer
	MainThreadRunner.RunOnMain(func() {
		viewer = NewPacketListViewer(window, 0)

		if viewer == nil {
			panic("viewer is nil")
		}
		viewer.BindToConversation(conv)

		window.PacketListViewers = append(window.PacketListViewers, viewer)
		window.TabWidget.AddTab(viewer, "Conversation: "+name)
	})
	<-MainThreadRunner.Wait
	if viewer == nil {
		panic("viewer is nil")
	}
	return viewer
}

func (window *DissectorWindow) CaptureFromHandle(handle *pcap.Handle, name string, isIPv4 bool, progressChan chan int) {
	err := handle.SetBPFFilter("udp")
	if err != nil {
		window.ShowCaptureError(err)
		return
	}

	var packetSource *gopacket.PacketSource
	if isIPv4 {
		packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
	} else {
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	}

	go captureJob(context.TODO(), name, packetSource, window, progressChan)
}

func (window *DissectorWindow) CaptureFromFile(file string, isIPv4 bool) {
	// TODO: Close this handle somewhere?
	handle, err := pcap.OpenOffline(file)
	if err != nil {
		window.ShowCaptureError(err)
		return
	}

	countPackets, err := gopcap.Count(file)
	if err != nil {
		window.ShowCaptureError(err)
		return
	}
	progressChan := make(chan int, 8)

	progressDialog := widgets.NewQProgressDialog2("Reading packets...", "Cancel", 0, int(countPackets), window, 0)
	progressDialog.SetWindowTitle("PCAP parsing in progress")

	go func() {
		for newProgress := range progressChan {
			progressDialog.SetValue(newProgress)
		}
		progressDialog.SetValue(int(countPackets))
	}()

	window.CaptureFromHandle(handle, path.Base(file), isIPv4, progressChan)
}

func (window *DissectorWindow) CaptureFromLive(itfName string, promisc bool) {
	handle, err := pcap.OpenLive(itfName, 2000, promisc, 10*time.Second)
	if err != nil {
		window.ShowCaptureError(err)
		return
	}

	window.CaptureFromHandle(handle, "<LIVE>", false, nil)
}

func (window *DissectorWindow) CaptureFromInjectionProxy(src string, dst string) {
	/*go func() {
		err := captureFromInjectionProxy(context.TODO(), src, dst, window)
		if err != nil {
			window.ShowCaptureError(err)
		}
	}()*/
	window.ShowCaptureError(errors.New("injection proxy is disabled (FIXME)"))
}
func (window *DissectorWindow) CaptureFromDivertProxy(_ *PlayerProxySettings) {
	window.ShowCaptureError(errors.New("divert proxy is disabled (FIXME)"))
}
func (window *DissectorWindow) CaptureFromPlayerProxy(_ *PlayerProxySettings) {
	window.ShowCaptureError(errors.New("divert proxy is disabled (FIXME)"))
}

func NewDissectorWindow(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *DissectorWindow {
	window := &DissectorWindow{
		QMainWindow: widgets.NewQMainWindow(parent, flags),

		StudioSettings:      &StudioSettings{},
		PlayerSettings:      &PlayerSettings{},
		ServerSettings:      &ServerSettings{},
		DefaultsSettings:    &DefaultsSettings{},
		PlayerProxySettings: &PlayerProxySettings{},
	}

	window.SetWindowTitle("Roblox Dissector")

	captureBar := window.MenuBar().AddMenu2("&Capture")
	captureFileAction := captureBar.AddAction("From &file...")
	capture4FileAction := captureBar.AddAction("From &RawCap file...")
	captureLiveAction := captureBar.AddAction("From &live interface...")
	captureInjectAction := captureBar.AddAction("From &injection proxy...")
	captureDivertAction := captureBar.AddAction("From &WinDivert proxy...")
	// FIXME: No capture stop action
	//captureStopAction := captureBar.AddAction("&Stop capture")
	captureFromPlayerProxyAction := captureBar.AddAction("From pl&ayer proxy")

	captureFileAction.ConnectTriggered(func(checked bool) {
		file := widgets.QFileDialog_GetOpenFileName(window, "Capture from file", "", "PCAP files (*.pcap)", "", 0)
		window.CaptureFromFile(file, false)
	})

	capture4FileAction.ConnectTriggered(func(checked bool) {
		file := widgets.QFileDialog_GetOpenFileName(window, "Capture from RawCap file", "", "PCAP files (*.pcap)", "", 0)
		window.CaptureFromFile(file, true)
	})
	captureLiveAction.ConnectTriggered(func(checked bool) {
		NewSelectInterfaceWidget(window, window.CaptureFromLive)
	})
	captureInjectAction.ConnectTriggered(func(checked bool) {
		NewProxyCaptureWidget(window, window.CaptureFromInjectionProxy)
	})
	captureDivertAction.ConnectTriggered(func(checked bool) {
		NewPlayerProxyWidget(window, window.PlayerProxySettings, window.CaptureFromDivertProxy)
	})
	captureFromPlayerProxyAction.ConnectTriggered(func(checked bool) {
		NewPlayerProxyWidget(window, window.PlayerProxySettings, window.CaptureFromPlayerProxy)
	})

	tabWidget := widgets.NewQTabWidget(window)
	tabWidget.SetTabsClosable(true)
	window.TabWidget = tabWidget
	window.SetCentralWidget(tabWidget)

	return window
}
