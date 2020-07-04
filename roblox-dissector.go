package main

import (
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	settings, err := gtk.SettingsGetDefault()
	if err != nil {
		println("Failed to get settings:", err.Error())
		return
	}
	settings.Set("gtk-application-prefer-dark-theme", true)

	win, err := NewDissectorWindow()
	if err != nil {
		println("Failed to create main window:", err.Error())
		return
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.ShowAll()

	gtk.Main()
}
