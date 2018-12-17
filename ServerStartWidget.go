package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gskartwii/rbxfile"
	"github.com/therecipe/qt/widgets"
)

// normalizeReferences changes the references of instances to a normalized form
// peer expects all instances to be of the form scope_id
func normalizeReferences(children []*rbxfile.Instance, dictionary *peer.InstanceDictionary) {
	for _, instance := range children {
		instance.Reference = dictionary.NewReference()
		normalizeReferences(instance.Children, dictionary)
	}
}

func NewServerStartWidget(parent widgets.QWidget_ITF, settings *ServerSettings, callback func(*ServerSettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start server...")
	layout := widgets.NewQVBoxLayout()

	rbxlLabel := NewQLabelF("RBXL location:")
	rbxlTextBox := widgets.NewQLineEdit2(settings.RBXLLocation, nil)
	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		rbxlTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find place...", "", "RBXL files (*.rbxl)", "", 0))
	})
	layout.AddWidget(rbxlLabel, 0, 0)
	layout.AddWidget(rbxlTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	enumLabel := NewQLabelF("Enum schema location:")
	enumTextBox := widgets.NewQLineEdit2(settings.EnumSchemaLocation, nil)
	browseButton = widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		enumTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find enum schema...", "", "Text files (*.txt)", "", 0))
	})
	layout.AddWidget(enumLabel, 0, 0)
	layout.AddWidget(enumTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	instanceLabel := NewQLabelF("Instance schema location:")
	instanceTextBox := widgets.NewQLineEdit2(settings.InstanceSchemaLocation, nil)
	browseButton = widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		instanceTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find instance schema...", "", "Text files (*.txt)", "", 0))
	})
	layout.AddWidget(instanceLabel, 0, 0)
	layout.AddWidget(instanceTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	portLabel := NewQLabelF("Port number:")
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddWidget(portLabel, 0, 0)
	layout.AddWidget(port, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectPressed(func() {
		window.Destroy(true, true)
		settings.Port = port.Text()
		settings.EnumSchemaLocation = enumTextBox.Text()
		settings.InstanceSchemaLocation = instanceTextBox.Text()
		settings.RBXLLocation = rbxlTextBox.Text()
		callback(settings)
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
