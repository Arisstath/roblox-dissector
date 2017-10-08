package main
import "github.com/therecipe/qt/widgets"
import "strconv"

func NewProxyCaptureWidget(parent widgets.QWidget_ITF, callback func(uint16, uint16)()) {
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
	startButton.ConnectPressed(func() {
		window.Destroy(true, true)
		src, err := strconv.Atoi(srcTextBox.Text())
		if err != nil {
			println("fail src conv")
			return
		}
		dst, err := strconv.Atoi(dstTextBox.Text())
		if err != nil {
			println("fail dst conv")
			return
		}
		callback(uint16(src), uint16(dst))
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
