package main

import (
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
)

func NewEditFilterWindow(oldFilter string, callback func(string)) error {
	box, err := boxWithMargin()
	if err != nil {
		return err
	}

	filterBuffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		return err
	}
	filterBuffer.SetText(oldFilter)
	filterInput, err := gtk.TextViewNewWithBuffer(filterBuffer)
	if err != nil {
		return err
	}
	filterInput.SetProperty("monospace", true)
	filterInput.SetVExpand(true)
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return err
	}
	scrolled.Add(filterInput)
	box.Add(scrolled)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}

	buttonRow, err := gtk.ButtonBoxNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return err
	}
	buttonRow.SetLayout(gtk.BUTTONBOX_END)
	buttonRow.SetSpacing(8)
	loadFromFile, err := gtk.ButtonNewWithLabel("Load from file")
	if err != nil {
		return err
	}
	loadFromFile.Connect("clicked", func() {
		chooser, err := gtk.FileChooserNativeDialogNew("Choose filter", win, gtk.FILE_CHOOSER_ACTION_OPEN, "Choose", "Cancel")
		if err != nil {
			ShowError(win, err, "Making chooser")
			return
		}
		filter, err := gtk.FileFilterNew()
		if err != nil {
			ShowError(win, err, "Creating filter")
			return
		}
		filter.AddPattern("*.lua")
		filter.SetName("Lua filter scripts (*.lua)")
		chooser.AddFilter(filter)
		resp := chooser.NativeDialog.Run()
		if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
			filename := chooser.GetFilename()
			contents, err := ioutil.ReadFile(filename)
			if err != nil {
				ShowError(win, err, "Reading filter")
				return
			}
			filterBuffer.SetText(string(contents))
		}
	})
	buttonRow.Add(loadFromFile)

	okButton, err := gtk.ButtonNewWithLabel("OK")
	if err != nil {
		return err
	}
	okButton.Connect("clicked", func() {
		text, err := filterBuffer.GetProperty("text")
		if err != nil {
			ShowError(win, err, "Getting filter")
			return
		}
		win.Destroy()
		callback(text.(string))
	})
	buttonRow.Add(okButton)

	box.Add(buttonRow)

	win.Add(box)
	win.SetTitle("Edit filter script")
	win.SetSizeRequest(800, 640)
	win.ShowAll()

	return nil
}
