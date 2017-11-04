package main
import "github.com/therecipe/qt/widgets"

func NewProxyCaptureWidget(parent widgets.QWidget_ITF, callback func(string, string)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Set up proxy...")
	layout := widgets.NewQVBoxLayout()

	srcLabel := NewQLabelF("Source port:")
	srcTextBox := widgets.NewQLineEdit2("53640", nil)
	layout.AddWidget(srcLabel, 0, 0)
	layout.AddWidget(srcTextBox, 0, 0)

	dstLabel := NewQLabelF("Destination port:")
	dstTextBox := widgets.NewQLineEdit2("53641", nil)
	layout.AddWidget(dstLabel, 0, 0)
	layout.AddWidget(dstTextBox, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Destroy(true, true)
		callback(srcTextBox.Text(), dstTextBox.Text())
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
