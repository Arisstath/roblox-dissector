package main

import (
	"github.com/gotk3/gotk3/gtk"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-profile" {
		go func() {
			println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
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
