package main

import (
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	winBuilder, err := gtk.BuilderNewFromFile("dissectorwindow.ui")
	if err != nil {
		println("Failed to create main window: ", err.Error())
		return
	}
	win, err := winBuilder.GetObject("dissectorwindow")
	if err != nil {
		println("Failed to create main window: ", err.Error())
		return
	}
	win.(*gtk.Window).ShowAll()

	gtk.Main()
}
