package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gskartwii/rbxfile/bin"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewClientConsole(parent widgets.QWidget_ITF, client *peer.CustomClient) {
	var logBuffer strings.Builder
	client.Logger = log.New(&logBuffer, "", log.Ltime|log.Lmicroseconds)

	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Client watch console")
	layout := widgets.NewQVBoxLayout()

	logLabel := NewQLabelF("Client log:")
	log := widgets.NewQPlainTextEdit(window)
	log.SetReadOnly(true)
	layout.AddWidget(logLabel, 0, 0)
	layout.AddWidget(log, 0, 0)

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for true {
			<-ticker.C
			log.Clear()
			log.InsertPlainText(logBuffer.String())
		}
	}()

	window.ConnectCloseEvent(func(_ *gui.QCloseEvent) {
		ticker.Stop()
	})

	labelForChat := NewQLabelF("Message, [to player, channel]:")
	message := widgets.NewQLineEdit(window)
	toPlayer := widgets.NewQLineEdit(window)
	channel := widgets.NewQLineEdit(window)
	layout.AddWidget(labelForChat, 0, 0)
	layout.AddWidget(message, 0, 0)
	layout.AddWidget(toPlayer, 0, 0)
	layout.AddWidget(channel, 0, 0)

	sendMessage := widgets.NewQPushButton2("Send!", window)
	sendMessage.ConnectPressed(func() {
		client.SendChat(message.Text(), toPlayer.Text(), channel.Text())
	})
	layout.AddWidget(sendMessage, 0, 0)

	dumpSchema := widgets.NewQPushButton2("Dump schema...", window)
	dumpSchema.ConnectPressed(func() {
		location := widgets.QFileDialog_GetSaveFileName(window, "Save schema...", "", "Text files (*.txt)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}
		location2 := widgets.QFileDialog_GetSaveFileName(window, "Save enums...", "", "Text files (*.txt)", "", 0)
		writer2, err := os.OpenFile(location2, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}
		client.Context.StaticSchema.Dump(writer, writer2)
	})

	dumpRbxl := widgets.NewQPushButton2("Dump DataModel...", window)
	dumpRbxl.ConnectPressed(func() {
		location := widgets.QFileDialog_GetSaveFileName(window, "Save as RBXL...", "", "Roblox place files (*.rbxl)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		writableClone := client.Context.DataModel.Copy()
		stripInvalidTypes(writableClone.Instances, nil, 0)

		err = bin.SerializePlace(writer, nil, client.Context.DataModel)
		if err != nil {
			println("while serializing place:", err.Error())
			return
		}
	})
	layout.AddWidget(dumpSchema, 0, 0)
	layout.AddWidget(dumpRbxl, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
