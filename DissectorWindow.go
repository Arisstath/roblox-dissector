package main

import (
	"context"
	"errors"
	"github.com/gotk3/gotk3/gtk"
	"github.com/google/gopacket/pcap"
)

func invalidUi(name string) error {
	return errors.New("invalid ui (" + name + ")")
}

type DissectorWindow struct {
	*gtk.Window

	tabs *gtk.Notebook
}

func (win *DissectorWindow) ShowCaptureError(err error, extrainfo string) {
	dialog := gtk.MessageDialogNew(
		win,
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

func (win *DissectorWindow) CaptureFromFile(filename string) {
	println("Capture from", filename)
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
    	win.ShowCaptureError(err, "Starting capture")
    	return
	}
	context, cancelFunc := context.WithCancel(context.TODO())
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
		win.tabs.AppendPage(listViewer.treeView, titleLabel)
		listViewer.treeView.ShowAll()
		titleLabel.ShowAll()
	})
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	go func() {
    	err := CaptureFromHandle(context, session, handle, nil)
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
		chooser.AddFilter(filter)
		resp := chooser.NativeDialog.Run()
		if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
			filename := chooser.GetFilename()
			dwin.CaptureFromFile(filename)
		}
	})

	return wind, nil
}
