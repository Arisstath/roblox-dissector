package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewClientConsole(parent widgets.QWidget_ITF, client *peer.CustomClient, flags core.Qt__WindowType, ctx context.Context, cancelFunc context.CancelFunc) *widgets.QWidget {
	var logBuffer strings.Builder
	client.Logger = log.New(&logBuffer, "", log.Ltime|log.Lmicroseconds)

	window := widgets.NewQWidget(parent, flags)
	layout := NewTopAlignLayout()

	logLabel := NewLabel("Client log:")
	log := widgets.NewQPlainTextEdit(window)
	log.SetReadOnly(true)
	layout.AddWidget(logLabel, 0, 0)
	layout.AddWidget(log, 0, 0)

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				MainThreadRunner.RunOnMain(func() {
					log.Clear()
					log.InsertPlainText(logBuffer.String())
					log.VerticalScrollBar().SetValue(log.VerticalScrollBar().Maximum())
				})
				<-MainThreadRunner.Wait
			case <-ctx.Done():
				return
			}
		}
	}()

	window.ConnectCloseEvent(func(_ *gui.QCloseEvent) {
		ticker.Stop()
		go client.Disconnect()
		cancelFunc()
	})

	chatWidget := widgets.NewQWidget(window, 0)
	formLayout := widgets.NewQFormLayout(chatWidget)

	message := widgets.NewQLineEdit(window)
	toPlayer := widgets.NewQLineEdit(window)
	channel := widgets.NewQLineEdit(window)
	formLayout.AddRow3("Message:", message)
	formLayout.AddRow3("(To player):", toPlayer)
	formLayout.AddRow3("(Channel):", channel)
	sendMessage := widgets.NewQPushButton2("Send", window)
	sendMessage.ConnectReleased(func() {
		client.SendChat(message.Text(), toPlayer.Text(), channel.Text())
	})
	formLayout.AddRow5(sendMessage)

	chatWidget.SetLayout(formLayout)
	layout.AddWidget(chatWidget, 0, 0)

	window.SetLayout(layout)

	return window
}
