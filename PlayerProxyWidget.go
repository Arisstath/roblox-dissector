package main

import "github.com/therecipe/qt/widgets"

func NewPlayerProxyWidget(parent widgets.QWidget_ITF, settings *PlayerProxySettings, callback func(*PlayerProxySettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose HTTPS server settings...")
	layout := widgets.NewQFormLayout(parent)

	certfileLayout := NewFileBrowseLayout(window, false, settings.Certfile, "Find certfile...", "Certfile (*.crt)")
	certfileWidget := widgets.NewQWidget(nil, 0)
	certfileWidget.SetLayout(certfileLayout)
	layout.AddRow3("Certfile location:", certfileWidget)

	keyfileLayout := NewFileBrowseLayout(window, false, settings.Certfile, "Find keyfile...", "Keyfile (*.pem)")
	keyfileWidget := widgets.NewQWidget(nil, 0)
	keyfileWidget.SetLayout(keyfileLayout)
	layout.AddRow3("Keyfile location:", keyfileWidget)

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
