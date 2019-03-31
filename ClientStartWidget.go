package main

import "github.com/therecipe/qt/widgets"

import "github.com/Gskartwii/roblox-dissector/peer"
import "strconv"

func NewClientStartWidget(parent widgets.QWidget_ITF, settings *peer.CustomClient, callback func(uint32, string, string)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start self client...")
	layout := NewTopAlignLayout()

	placeIdLabel := NewLabel("Place id:")
	placeId := widgets.NewQLineEdit2("12109643", window)
	layout.AddWidget(placeIdLabel, 0, 0)
	layout.AddWidget(placeId, 0, 0)

	usernameLabel := NewLabel("Username:")
	passwordLabel := NewLabel("Password:")
	username := widgets.NewQLineEdit(window)
	password := widgets.NewQLineEdit(window)
	password.SetEchoMode(widgets.QLineEdit__Password)
	layout.AddWidget(usernameLabel, 0, 0)
	layout.AddWidget(username, 0, 0)
	layout.AddWidget(passwordLabel, 0, 0)
	layout.AddWidget(password, 0, 0)

	startButton := widgets.NewQPushButton2("Start", window)
	startButton.ConnectReleased(func() {
		window.Destroy(true, true)
		placeIdVal, _ := strconv.Atoi(placeId.Text())
		callback(uint32(placeIdVal), username.Text(), password.Text())
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
