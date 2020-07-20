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

	tabs *gtk.Notebook
}

func ShowError(wdg gtk.IWidget, err error, extrainfo string) {
	widget := wdg.ToWidget()
	parentWindow, err := widget.GetToplevel()
	if err != nil {
		println("failed to find parent window:", err.Error())
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

	countPackets := 0.0
	frac := 0.0

	lastUpdate := time.Now()
	session.ProgressCallback = func(progress int) {
		if progress == -1 {
			progressDialog.Close()
			return
		}

		now := time.Now()
		if now.Sub(lastUpdate) < 100*time.Millisecond {
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

func NewDissectorWindow() (*gtk.Window, error) {
	winBuilder, err := gtk.BuilderNewFromFile("dissectorwindow.ui")
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
	fromFileMenuItem.Connect("activate", func() {
		chooser, err := gtk.FileChooserNativeDialogNew("Choose PCAP file", wind, gtk.FILE_CHOOSER_ACTION_OPEN, "Choose", "Cancel")
		if err != nil {
			dwin.ShowCaptureError(err, "Making chooser")
			return
		}
		filter, err := gtk.FileFilterNew()
		if err != nil {
			dwin.ShowCaptureError(err, "Creating filter")
			return
		}
		filter.AddPattern("*.pcap")
		filter.SetName("PCAP network capture files (*.pcap)")
		chooser.AddFilter(filter)
		resp := chooser.NativeDialog.Run()
		if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
			filename := chooser.GetFilename()
			dwin.CaptureFromFile(filename)
		}
	})

	return wind, nil
}
