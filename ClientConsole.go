package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "roblox-dissector/peer"
import "time"
import "strings"
import "log"

func NewClientConsole(parent widgets.QWidget_ITF, client *peer.CustomClient) {
	var logBuffer strings.Builder
	client.Logger = log.New(&logBuffer, "", log.Ltime | log.Lmicroseconds)

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
			<- ticker.C
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

	window.SetLayout(layout)
	window.Show()
}
