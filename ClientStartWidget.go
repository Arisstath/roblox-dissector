package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/core"
import "github.com/Gskartwii/roblox-dissector/peer"
import "strconv"

func NewClientStartWidget(parent widgets.QWidget_ITF, settings *peer.CustomClient, callback func(uint32, bool, string)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start self client...")
	layout := widgets.NewQVBoxLayout()

	placeIdLabel := NewQLabelF("Place id:")
	placeId := widgets.NewQLineEdit2("0", window)
	layout.AddWidget(placeIdLabel, 0, 0)
	layout.AddWidget(placeId, 0, 0)

	isGuest := widgets.NewQCheckBox2("Connect as guest?", window)
	layout.AddWidget(isGuest, 0, 0)

	ticketLabel := NewQLabelF("Auth ticket:")
	ticket := widgets.NewQLineEdit(window)
	layout.AddWidget(ticketLabel, 0, 0)
	layout.AddWidget(ticket, 0, 0)

	startButton := widgets.NewQPushButton2("Start", window)
	startButton.ConnectClicked(func(_ bool) {
		window.Destroy(true, true)
		placeIdVal, _ := strconv.Atoi(placeId.Text())
		callback(uint32(placeIdVal), isGuest.CheckState() == core.Qt__Checked, ticket.Text())
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
