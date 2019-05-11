package main

import "github.com/therecipe/qt/widgets"

func NewPlayerProxyWidget(parent widgets.QWidget_ITF, settings *PlayerProxySettings, callback func(*PlayerProxySettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose HTTPS server settings...")
	layout := widgets.NewQFormLayout(parent)

	certfileLayout := NewFileBrowseLayout(window, false, settings.Certfile, "Find certfile...", "Certfile (*.crt)")
	layout.AddRow4("Certfile location:", certfileLayout)

	keyfileLayout := NewFileBrowseLayout(window, false, settings.Certfile, "Find keyfile...", "Keyfile (*.pem)")
	layout.AddRow4("Keyfile location:", keyfileLayout)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Close()
		settings.Certfile = certfileLayout.FileName()
		settings.Keyfile = keyfileLayout.FileName()
		callback(settings)
	})
	layout.AddRow5(startButton)

	window.SetLayout(layout)
	window.Show()
}
