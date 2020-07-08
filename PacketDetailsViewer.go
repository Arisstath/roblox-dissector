package main

import (
	"encoding/hex"
	"io/ioutil"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type PacketDetailsViewer struct {
	mainWidget *gtk.Notebook
	logBox     *gtk.TextView
	hexBox     *gtk.TextView

	hexDumpData []byte
}

func NewPacketDetailsViewer() (*PacketDetailsViewer, error) {
	viewer := &PacketDetailsViewer{}

	builder, err := gtk.BuilderNewFromFile("packetdetailsviewer.ui")
	if err != nil {
		return nil, err
	}
	notebook_, err := builder.GetObject("packetdetailsnotebook")
	if err != nil {
		return nil, err
	}

	notebook, ok := notebook_.(*gtk.Notebook)
	if !ok {
		return nil, invalidUi("notebook")
	}

	notebookContainer_, err := builder.GetObject("packetdetailsviewerwindow")
	if err != nil {
		return nil, err
	}
	notebookContainer, ok := notebookContainer_.(*gtk.Window)
	if !ok {
		return nil, invalidUi("notebook container")
	}
	notebookContainer.Remove(notebook)

	logBox_, err := builder.GetObject("logbox")
	if err != nil {
		return nil, err
	}
	logBox, ok := logBox_.(*gtk.TextView)
	if !ok {
		return nil, invalidUi("logbox")
	}
	hexBox_, err := builder.GetObject("hexbox")
	if err != nil {
		return nil, err
	}
	hexBox, ok := hexBox_.(*gtk.TextView)
	if !ok {
		return nil, invalidUi("hexbox")
	}
	copyHex_, err := builder.GetObject("copyashexstreambutton")
	if err != nil {
		return nil, err
	}
	copyHex, ok := copyHex_.(*gtk.Button)
	if !ok {
		return nil, invalidUi("copyhex")
	}
	copyHex.Connect("clicked", func() {
		clipboard, err := gtk.ClipboardGet(gdk.GdkAtomIntern("CLIPBOARD", true))
		if err != nil {
			println("Failed to get clipboard:", err.Error())
			return
		}
		clipboard.SetText(hex.EncodeToString(viewer.hexDumpData))
	})
	saveHex_, err := builder.GetObject("savehexbutton")
	if err != nil {
		return nil, err
	}
	saveHex, ok := saveHex_.(*gtk.Button)
	if !ok {
		return nil, invalidUi("savehex")
	}
	saveHex.Connect("clicked", func() {
		parentWindow, err := viewer.mainWidget.GetToplevel()
		if err != nil {
			println("Failed to get parent window:", err.Error())
			return
		}
		dialog, err := gtk.FileChooserNativeDialogNew(
			"Save packet as",
			parentWindow.(gtk.IWindow),
			gtk.FILE_CHOOSER_ACTION_SAVE,
			"Save",
			"Cancel",
		)
		if err != nil {
			println("Failed to make dialog:", err.Error())
			return
		}
		filter, err := gtk.FileFilterNew()
		if err != nil {
			println("Failed to make filter:", err.Error())
			return
		}
		filter.AddPattern("*.bin")
		dialog.AddFilter(filter)
		resp := dialog.Run()
		if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
			filename := dialog.GetFilename()
			err := ioutil.WriteFile(filename, viewer.hexDumpData, 0644)
			if err != nil {
				ShowError(viewer.mainWidget, err, "Saving packet to file")
			}
		}
	})

	viewer.mainWidget = notebook
	viewer.logBox = logBox
	viewer.hexBox = hexBox

	return viewer, nil
}

func (viewer *PacketDetailsViewer) ShowPacket(layers *peer.PacketLayers) error {
	if layers.OfflinePayload != nil {
		viewer.hexDumpData = layers.OfflinePayload
	} else {
		viewer.hexDumpData = layers.Reliability.SplitBuffer.Data
	}
	textBuffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		return err
	}
	textBuffer.SetText(hex.Dump(viewer.hexDumpData))
	viewer.hexBox.SetBuffer(textBuffer)

	return nil
}
