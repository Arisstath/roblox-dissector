
package main
import "github.com/therecipe/qt/widgets"

func NewPlayerProxyWidget(parent widgets.QWidget_ITF, settings *PlayerProxySettings, callback func(*PlayerProxySettings)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose HTTPS server settings...")
	layout := NewTopAlignLayout()

	certfileLabel := NewLabel("Certfile location:")
	certfileTextBox := widgets.NewQLineEdit2(settings.Certfile, nil)
	certbrowseButton := widgets.NewQPushButton2("Browse...", nil)
	certbrowseButton.ConnectReleased(func() {
		certfileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find certfile...", "", "Certfiles (*.crt)", "", 0))
	})
	layout.AddWidget(certfileLabel, 0, 0)
	layout.AddWidget(certfileTextBox, 0, 0)
	layout.AddWidget(certbrowseButton, 0, 0)

	keyfileLabel := NewLabel("Keyfile location:")
	keyfileTextBox := widgets.NewQLineEdit2(settings.Keyfile, nil)
	keybrowseButton := widgets.NewQPushButton2("Browse...", nil)
	keybrowseButton.ConnectReleased(func() {
		keyfileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find keyfile...", "", "Keyfiles (*.pem)", "", 0))
	})
	layout.AddWidget(keyfileLabel, 0, 0)
	layout.AddWidget(keyfileTextBox, 0, 0)
	layout.AddWidget(keybrowseButton, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Destroy(true, true)
		settings.Certfile = certfileTextBox.Text()
		settings.Keyfile = keyfileTextBox.Text()
		callback(settings)
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
