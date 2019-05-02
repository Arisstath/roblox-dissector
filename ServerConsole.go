package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/olebedev/emitter"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewServerConsole(parent widgets.QWidget_ITF, server *peer.CustomServer, session *CaptureSession) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Server watch console")
	layout := NewTopAlignLayout()

	clientsLabel := NewLabel("Clients:")
	clients := widgets.NewQTreeView(window)
	standardModel := NewProperSortModel(clients)
	standardModel.SetHorizontalHeaderLabels([]string{"Address"})

	updateClients := func() {
		rootNode := standardModel.InvisibleRootItem()
		println("after disc have", rootNode.RowCount(), len(server.Clients))
		if rootNode.RowCount() != len(server.Clients) {
			standardModel.Clear()
			standardModel.SetHorizontalHeaderLabels([]string{"Address"})
			rootNode = standardModel.InvisibleRootItem()
			for _, client := range server.Clients {
				rootNode.AppendRow([]*gui.QStandardItem{NewStringItem(client.Address.String())})
			}
		}
	}

	server.ClientEmitter.On("client", func(e *emitter.Event) {
		updateClients()
		client := e.Args[0].(*peer.ServerClient)
		client.GenericEvents.On("disconnected", func(e *emitter.Event) {
			println("received client disconnection")
			updateClients()
		}, emitter.Void)
	}, emitter.Void)
	clients.SetModel(standardModel)
	clients.SetSelectionMode(1)
	layout.AddWidget(clientsLabel, 0, 0)
	layout.AddWidget(clients, 0, 0)

	stopButton := widgets.NewQPushButton2("Stop", window)
	layout.AddWidget(stopButton, 0, 0)
	// don't close the window upon stopping capture, that's unintuitive
	stopButton.ConnectReleased(session.StopCapture)

	window.SetLayout(layout)
	window.ConnectCloseEvent(func(_ *gui.QCloseEvent) {
		session.StopCapture()
	})
	window.Show()
}
