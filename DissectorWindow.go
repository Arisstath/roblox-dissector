package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dreadl0ck/gopcap"
	"github.com/google/gopacket/pcap"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func invalidUi(name string) error {
	return errors.New("invalid ui (" + name + ")")
}

type DissectorWindow struct {
	*gtk.Window

	tabs           *gtk.Notebook
	forgetAcksItem *gtk.CheckMenuItem
}

func ShowError(wdg gtk.IWidget, err error, extrainfo string) {
	widget := wdg.ToWidget()
	parentWindow, topLevelErr := widget.GetToplevel()
	if topLevelErr != nil {
		println("failed to find parent window:", topLevelErr.Error())
		return
	}
	dialog := gtk.MessageDialogNew(
		parentWindow.(gtk.IWindow),
		gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_OK,
		"%s: %s",
		extrainfo,
		err.Error(),
	)
	dialog.SetTitle("Error")
	dialog.ShowAll()
	dialog.Show()
	dialog.Run()
}

func (win *DissectorWindow) ShowCaptureError(err error, extrainfo string) {
	ShowError(win, err, extrainfo)
}

func (win *DissectorWindow) CaptureFromPcapDevice(name string) {
	handle, err := pcap.OpenLive(name, 2000, false, 1*time.Second)
	if err != nil {
		println("error starting capture", err.Error())
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	context, cancelFunc := context.WithCancel(context.TODO())
	session, err := NewCaptureSession(name, cancelFunc, func(listViewer *PacketListViewer, err error) {
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		titleLabel, err := gtk.LabelNew(listViewer.title)
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		win.tabs.AppendPage(listViewer.mainWidget, titleLabel)
		listViewer.mainWidget.ShowAll()
		titleLabel.ShowAll()

		windowHeight := win.GetAllocatedHeight()
		paneHeight := int(0.6 * float64(windowHeight))
		listViewer.mainWidget.SetPosition(paneHeight)
		listViewer.mainWidget.SetWideHandle(true)
	})
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	session.ForgetAcks = win.forgetAcksItem.GetActive()

	go func() {
		err = CaptureFromHandle(context, session, handle)
		if err != nil {
			win.ShowCaptureError(err, "Starting capture")
			return
		}
	}()
}

func (win *DissectorWindow) CaptureFromFile(filename string) {
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	context, cancelFunc := context.WithCancel(context.TODO())

	progressDialog, err := gtk.DialogNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressDialog.SetTitle("File capture in progress")
	contentArea, err := progressDialog.GetContentArea()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressBar, err := gtk.ProgressBarNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressBar.Pulse()
	progressBar.SetShowText(true)
	progressBar.SetText("Scanning for packets...")
	contentArea.Add(progressBar)
	progressDialog.SetDeletable(false) // Hide "close" button
	progressDialog.SetModal(true)
	progressDialog.SetSizeRequest(200, 32)
	progressDialog.ShowAll()

	session, err := NewCaptureSession(filename, cancelFunc, func(listViewer *PacketListViewer, err error) {
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		titleLabel, err := gtk.LabelNew(listViewer.title)
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		win.tabs.AppendPage(listViewer.mainWidget, titleLabel)
		listViewer.mainWidget.ShowAll()
		titleLabel.ShowAll()

		windowHeight := win.GetAllocatedHeight()
		paneHeight := int(0.6 * float64(windowHeight))
		listViewer.mainWidget.SetPosition(paneHeight)
		listViewer.mainWidget.SetWideHandle(true)
	})
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	session.ForgetAcks = win.forgetAcksItem.GetActive()

	countPackets := 0.0
	frac := 0.0

	lastUpdate := time.Now()
	session.ProgressCallback = func(progress int) {
		if progress == -1 {
			progressDialog.Close()
			return
		}

		now := time.Now()
		if now.Sub(lastUpdate) < 500*time.Millisecond {
			return
		}
		lastUpdate = now

		frac = float64(progress) / countPackets
		progressBar.SetFraction(frac)
		progressBar.SetText(fmt.Sprintf("Reading packets: %.1f %%", frac*100))

		// Force redraw of this progress bar
		// We only force this every 100 ms, so it shouldn't be too bad
		glib.MainContextDefault().Iteration(false)
	}

	go func() {
		count, err := gopcap.Count(filename)
		if err != nil {
			glib.IdleAdd(func() bool {
				progressDialog.Close()
				return false
			})
		} else {
			countPackets = float64(count)
		}

		err = CaptureFromHandle(context, session, handle)
		session.ReportDone()
		if err != nil {
			win.ShowCaptureError(err, "Starting capture")
			return
		}
	}()
}

func (win *DissectorWindow) PromptCaptureFromFile() {
	chooser, err := gtk.FileChooserNativeDialogNew("Choose PCAP file", win, gtk.FILE_CHOOSER_ACTION_OPEN, "Choose", "Cancel")
	if err != nil {
		win.ShowCaptureError(err, "Making chooser")
		return
	}
	filter, err := gtk.FileFilterNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating filter")
		return
	}
	filter.AddPattern("*.pcap")
	filter.SetName("PCAP network capture files (*.pcap)")
	chooser.AddFilter(filter)
	resp := chooser.NativeDialog.Run()
	if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
		filename := chooser.GetFilename()
		win.CaptureFromFile(filename)
	}
}

func (win *DissectorWindow) PromptCaptureLive() {
	err := PromptInterfaceName(win.CaptureFromPcapDevice)
	if err != nil {
		win.ShowCaptureError(err, "Making interface chooser")
	}
}

func NewDissectorWindow() (*gtk.Window, error) {
	winBuilder, err := gtk.BuilderNewFromFile("res/dissectorwindow.ui")
	if err != nil {
		return nil, err
	}
	win, err := winBuilder.GetObject("dissectorwindow")
	if err != nil {
		return nil, err
	}

	wind, ok := win.(*gtk.Window)
	if !ok {
		return nil, invalidUi("mainwindow")
	}
	wind.SetTitle("Sala")
	wind.SetDefaultSize(800, 640)

	dwin := &DissectorWindow{
		Window: wind,
	}

	tabs, err := winBuilder.GetObject("conversationtabs")
	if err != nil {
		return nil, err
	}

	tabsNotebook, ok := tabs.(*gtk.Notebook)
	if !ok {
		return nil, invalidUi("convtabs")
	}
	dwin.tabs = tabsNotebook

	fromFileItem, err := winBuilder.GetObject("fromfileitem")
	if err != nil {
		return nil, err
	}
	fromFileMenuItem, ok := fromFileItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("fromfileitem")
	}
	fromFileMenuItem.Connect("activate", dwin.PromptCaptureFromFile)
	fromFileButton_, err := winBuilder.GetObject("fromfilebutton")
	if err != nil {
		return nil, err
	}
	fromFileButton, ok := fromFileButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("fromfilebutton")
	}
	fromFileButton.Connect("clicked", dwin.PromptCaptureFromFile)

	fromLiveItem, err := winBuilder.GetObject("frominterfaceitem")
	if err != nil {
		return nil, err
	}
	fromLiveMenuItem, ok := fromLiveItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("frominterfaceitem")
	}
	fromLiveMenuItem.Connect("activate", dwin.PromptCaptureLive)
	fromLiveButton_, err := winBuilder.GetObject("frominterfacebutton")
	if err != nil {
		return nil, err
	}
	fromLiveButton, ok := fromLiveButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("frominterfacebutton")
	}
	fromLiveButton.Connect("clicked", dwin.PromptCaptureLive)

	forgetAcksItem_, err := winBuilder.GetObject("forgetacksitem")
	if err != nil {
		return nil, err
	}
	forgetAcksItem, ok := forgetAcksItem_.(*gtk.CheckMenuItem)
	if !ok {
		return nil, invalidUi("forgetacksitem")
	}
	dwin.forgetAcksItem = forgetAcksItem

	return wind, nil
}
