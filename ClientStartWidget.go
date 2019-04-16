package main

import "github.com/therecipe/qt/widgets"

func NewClientStartWidget(parent widgets.QWidget_ITF, callback func(string)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start self client...")
	layout := NewTopAlignLayout()

	uriLabel := NewLabel("roblox-dissector URI:")
	uri := widgets.NewQLineEdit(window)
	layout.AddWidget(uriLabel, 0, 0)
	layout.AddWidget(uri, 0, 0)

	startButton := widgets.NewQPushButton2("Start", window)
	startButton.ConnectReleased(func() {
		window.Destroy(true, true)
		callback(uri.Text())
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
