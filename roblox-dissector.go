package main

import (
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		println("Failed to create window: ", err.Error())
		return
	}
	win.SetTitle("Sala")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	label, err := gtk.LabelNew("Hello world!")
	if err != nil {
		println("Failed to create label: ", err.Error())
		return
	}
	win.Add(label)
	win.ShowAll()

	gtk.Main()
}
