package main

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	"github.com/therecipe/qt/widgets"
)

func NewServerStartWidget(parent widgets.QWidget_ITF, settings *ServerSettings, callback func(*ServerSettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start server...")
	layout := widgets.NewQFormLayout(window)

	rbxlLayout := NewFileBrowseLayout(window, false, settings.RBXLLocation, "Find place file...", "RBXLX files (*.rbxlx)")
	layout.AddRow4("RBXLX location:", rbxlLayout)

	schemaLayout := NewFileBrowseLayout(window, false, settings.SchemaLocation, "Find schema...", "Text files (*.txt)")
	layout.AddRow4("Schema location:", schemaLayout)

	// HACK: convenience
	if settings.Port == "" {
		settings.Port = "53640"
	}
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddRow3("Port number:", port)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Close()
		settings.Port = port.Text()
		settings.SchemaLocation = schemaLayout.FileName()
		settings.RBXLLocation = rbxlLayout.FileName()
		callback(settings)
	})
	layout.AddRow5(startButton)

	window.SetLayout(layout)
	window.Show()
}
