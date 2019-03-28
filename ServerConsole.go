package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/Gskartwii/roblox-dissector/peer"
import "time"

func NewServerConsole(parent widgets.QWidget_ITF, server *peer.CustomServer) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Server watch console")
	layout := NewTopAlignLayout()

	clientsLabel := NewLabel("Clients:")
	clients := widgets.NewQTreeView(window)
	standardModel := NewProperSortModel(clients)
	standardModel.SetHorizontalHeaderLabels([]string{"Address"})

	updateClients := func() {
		rootNode := standardModel.InvisibleRootItem()
		if rootNode.RowCount() != len(server.Clients) {
			standardModel.Clear()
			standardModel.SetHorizontalHeaderLabels([]string{"Address"})
			rootNode = standardModel.InvisibleRootItem()
			for _, client := range server.Clients {
				rootNode.AppendRow([]*gui.QStandardItem{NewStringItem(client.Address.String())})
			}
		}
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for true {
			<-ticker.C
			updateClients()
		}
	}()

	clients.SetModel(standardModel)
	clients.SetSelectionMode(1)
	layout.AddWidget(clientsLabel, 0, 0)
	layout.AddWidget(clients, 0, 0)

	stopButton := widgets.NewQPushButton2("Stop", window)
	layout.AddWidget(stopButton, 0, 0)
	stopButton.ConnectPressed(func() {
		window.Destroy(true, true)
		server.Stop()
	})

	window.SetLayout(layout)
	window.ConnectCloseEvent(func(_ *gui.QCloseEvent) {
		ticker.Stop()
		server.Stop()
	})
	window.Show()
}
