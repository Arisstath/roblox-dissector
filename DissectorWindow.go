package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/dreadl0ck/gopcap"
	"github.com/google/gopacket/pcap"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
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

	CurrentSession        *CaptureSession
	CurrentConversation   *Conversation
	StopAction            *widgets.QAction
	BrowseDataModelAction *widgets.QAction

	TabWidget *widgets.QTabWidget
	Sessions  []*CaptureSession

	StudioVersion string
	PlayerVersion string
	// TODO: Remove?
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

func (window *DissectorWindow) CaptureFromFile(file string, isIPv4 bool) {
	handle, err := pcap.OpenOffline(file)
	if err != nil {
		window.ShowCaptureError(err)
		return
	}

	nameBase := filepath.Base(file)
	session := NewCaptureSession(nameBase, window)
	window.Sessions = append(window.Sessions, session)
	index := window.TabWidget.AddTab(session.PacketListViewers[0], fmt.Sprintf("Conversation: %s#1", nameBase))
	window.TabWidget.SetCurrentIndex(index)

	countPackets, err := gopcap.Count(file)
	if err != nil {
		session.StopCapture()
		window.ShowCaptureError(err)
		return
	}
	progressChan := make(chan int, 8)

	progressDialog := widgets.NewQProgressDialog2("Reading packets...", "Cancel", 0, int(countPackets), window, 0)
	progressDialog.SetWindowTitle("PCAP parsing in progress")
	progressDialog.SetWindowModality(core.Qt__WindowModal)

	go func() {
		for newProgress := range progressChan {
			progressDialog.SetValue(newProgress)
			if progressDialog.WasCanceled() {
				session.StopCapture()
				break
			}
		}
		progressDialog.SetValue(int(countPackets))
	}()

	go func() {
		err := session.CaptureFromHandle(handle, isIPv4, progressChan)
		handle.Close()
		MainThreadRunner.RunOnMain(func() {
			close(progressChan)
			if err != nil {
				window.ShowCaptureError(err)
			}
			session.StopCapture()
			// This must be called for session in which
			// the data isn't displayed during capture
			session.UpdateModels()
		})
		<-MainThreadRunner.Wait
	}()
}

func (window *DissectorWindow) CaptureFromLive(itfName string, promisc bool) {
	handle, err := pcap.OpenLive(itfName, 2000, promisc, 1*time.Second)
	if err != nil {
		window.ShowCaptureError(err)
		return
	}

	nameBase := "<LIVE>:" + itfName
	session := NewCaptureSession(nameBase, window)
	// This must be set for live captures
	session.SetModel = true
	window.Sessions = append(window.Sessions, session)
	index := window.TabWidget.AddTab(session.PacketListViewers[0], fmt.Sprintf("Conversation: %s#1", nameBase))
	window.TabWidget.SetCurrentIndex(index)

	go func() {
		err := session.CaptureFromHandle(handle, false, nil)
		handle.Close()
		MainThreadRunner.RunOnMain(func() {
			session.StopCapture()
			if err != nil {
				window.ShowCaptureError(err)
			}
		})
		<-MainThreadRunner.Wait
	}()
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

func (window *DissectorWindow) OpenFileHandler(_ bool) {
	file := widgets.QFileDialog_GetOpenFileName(window, "Capture from file", "", "PCAP files (*.pcap)", "", 0)
	if file != "" {
		window.CaptureFromFile(file, false)
	}
}

func (window *DissectorWindow) OpenLiveInterfaceHandler(_ bool) {
	NewSelectInterfaceWidget(window, window.CaptureFromLive)
}

func (window *DissectorWindow) StopActionHandler(_ bool) {
	window.StopAction.SetEnabled(false)
	if window.CurrentSession == nil {
		return
	}
	if window.CurrentSession.IsCapturing {
		window.CurrentSession.StopCapture()
	}
}

func (window *DissectorWindow) BrowseDataModelHandler(_ bool) {
	if window.CurrentConversation == nil {
		return
	}
	ctx := window.CurrentConversation.Context
	// TODO: What to do with default values?
	NewDataModelBrowser(ctx, ctx.DataModel)
}

func (window *DissectorWindow) SessionSelected(session *CaptureSession, conv *Conversation) {
	if conv == nil {
		window.BrowseDataModelAction.SetEnabled(false)
	} else {
		window.BrowseDataModelAction.SetEnabled(conv.Context != nil && conv.Context.DataModel != nil)
	}

	if session == nil {
		window.StopAction.SetEnabled(false)
	} else {
		window.StopAction.SetEnabled(session.IsCapturing)
	}
}

// Handy for updating the state when something happens
func (window *DissectorWindow) UpdateButtons() {
	window.TabSelected(window.TabWidget.CurrentIndex())
}

func (window *DissectorWindow) TabSelected(index int) {
	window.CurrentSession = nil
	window.CurrentConversation = nil
	if index == -1 {
		window.SessionSelected(nil, nil)
		return
	}
	widget := window.TabWidget.Widget(index)
	for _, session := range window.Sessions {
		found := session.FindViewer(widget)
		if found != nil {
			window.CurrentSession = session
			window.CurrentConversation = found.Conversation
		}
	}
	window.SessionSelected(window.CurrentSession, window.CurrentConversation)
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
	captureFromPlayerProxyAction := captureBar.AddAction("From pl&ayer proxy")

	captureFileAction.ConnectTriggered(window.OpenFileHandler)

	capture4FileAction.ConnectTriggered(func(checked bool) {
		file := widgets.QFileDialog_GetOpenFileName(window, "Capture from RawCap file", "", "PCAP files (*.pcap)", "", 0)
		if file != "" {
			window.CaptureFromFile(file, true)
		}
	})
	captureLiveAction.ConnectTriggered(window.OpenLiveInterfaceHandler)
	captureInjectAction.ConnectTriggered(func(checked bool) {
		NewProxyCaptureWidget(window, window.CaptureFromInjectionProxy)
	})
	captureDivertAction.ConnectTriggered(func(checked bool) {
		NewPlayerProxyWidget(window, window.PlayerProxySettings, window.CaptureFromDivertProxy)
	})
	captureFromPlayerProxyAction.ConnectTriggered(func(checked bool) {
		NewPlayerProxyWidget(window, window.PlayerProxySettings, window.CaptureFromPlayerProxy)
	})

	toolBar := widgets.NewQToolBar("Basic functions", window)

	openAction := toolBar.AddAction2(gui.NewQIcon5(":/qml/folder-open-line.svg"), "Open PCAP file... (Ctrl+O)")
	openAction.SetShortcut(gui.NewQKeySequence2("Ctrl+O", gui.QKeySequence__PortableText))
	openAction.ConnectTriggered(window.OpenFileHandler)
	liveAction := toolBar.AddAction2(gui.NewQIcon5(":/qml/cloud-network-line.svg"), "Open live interface... (Ctrl+L)")
	liveAction.SetShortcut(gui.NewQKeySequence2("Ctrl+L", gui.QKeySequence__PortableText))
	liveAction.ConnectTriggered(window.OpenLiveInterfaceHandler)

	stopAction := toolBar.AddAction2(gui.NewQIcon5(":/qml/stop-line.svg"), "Stop session (Ctrl+T)")
	stopAction.SetShortcut(gui.NewQKeySequence2("Ctrl+T", gui.QKeySequence__PortableText))
	stopAction.ConnectTriggered(window.StopActionHandler)
	stopAction.SetEnabled(false)
	window.StopAction = stopAction

	browseDataModelAction := toolBar.AddAction2(gui.NewQIcon5(":/qml/tree-view-line.svg"), "Browse DataModel... (Ctrl+D)")
	browseDataModelAction.SetShortcut(gui.NewQKeySequence2("Ctrl+D", gui.QKeySequence__PortableText))
	browseDataModelAction.ConnectTriggered(window.BrowseDataModelHandler)
	browseDataModelAction.SetEnabled(false)
	window.BrowseDataModelAction = browseDataModelAction

	window.AddToolBar(core.Qt__TopToolBarArea, toolBar)

	tabWidget := widgets.NewQTabWidget(window)
	tabWidget.SetTabsClosable(true)
	window.TabWidget = tabWidget
	window.SetCentralWidget(tabWidget)

	tabWidget.ConnectCurrentChanged(window.TabSelected)
	tabWidget.ConnectTabCloseRequested(func(index int) {
		widget := tabWidget.Widget(index)

		var thisSession *CaptureSession
		var sessionIndex int
		for i, session := range window.Sessions {
			if session.FindViewer(widget) != nil {
				thisSession = session
				sessionIndex = i
			}
		}
		isEmpty := thisSession.RemoveViewer(widget)
		if isEmpty {
			thisSession.Destroy()
			copy(window.Sessions[sessionIndex:], window.Sessions[sessionIndex+1:])
			window.Sessions[len(window.Sessions)-1] = nil
			window.Sessions = window.Sessions[:len(window.Sessions)-1]
		}
		tabWidget.RemoveTab(index)
	})

	return window
}
