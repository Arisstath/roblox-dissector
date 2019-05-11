package main

import "github.com/therecipe/qt/widgets"

func NewClientStartWidget(parent widgets.QWidget_ITF, callback func(string)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start self client...")
	layout := widgets.NewQFormLayout(window)

	uri := widgets.NewQLineEdit(window)
	layout.AddRow3("roblox-dissector URI:", uri)

	startButton := widgets.NewQPushButton2("Start", window)
	startButton.ConnectReleased(func() {
		window.Close()
		callback(uri.Text())
	})
	layout.AddRow5(startButton)

	window.SetLayout(layout)
	window.Show()
}
