package main
import "github.com/therecipe/qt/widgets"

func NewStudioChooser(parent widgets.QWidget_ITF, settings *StudioSettings, callback func(*StudioSettings)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose Studio location...")
	layout := widgets.NewQVBoxLayout()
	
	fileLabel := NewQLabelF("Studio location:")
	fileTextBox := widgets.NewQLineEdit2(settings.Location, nil)
	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectPressed(func() {
		fileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find RobloxStudioBeta.exe...", "", "Roblox Studio (*.exe)", "", 0))
	})
	layout.AddWidget(fileLabel, 0, 0)
	layout.AddWidget(fileTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	flagsLabel := NewQLabelF("Command line flags:")
	flags := widgets.NewQLineEdit2(settings.Flags, nil)
	layout.AddWidget(flagsLabel, 0, 0)
	layout.AddWidget(flags, 0, 0)

	addressLabel := NewQLabelF("Server address:")
	address := widgets.NewQLineEdit2(settings.Address, nil)
	layout.AddWidget(addressLabel, 0, 0)
	layout.AddWidget(address, 0, 0)

	portLabel := NewQLabelF("Port number:")
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddWidget(portLabel, 0, 0)
	layout.AddWidget(port, 0, 0)

	rbxlFileLabel := NewQLabelF("RBXL location:")
	rbxlFileTextBox := widgets.NewQLineEdit2(settings.RBXL, nil)
	browseRBXLButton := widgets.NewQPushButton2("Browse...", nil)
	browseRBXLButton.ConnectPressed(func() {
		rbxlFileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find place file...", "", "Roblox place files (*.rbxl *.rbxlx)", "", 0))
	})
	layout.AddWidget(rbxlFileLabel, 0, 0)
	layout.AddWidget(rbxlFileTextBox, 0, 0)
	layout.AddWidget(browseRBXLButton, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectPressed(func() {
		window.Destroy(true, true)
		settings.Location = fileTextBox.Text()
		settings.Flags = flags.Text()
		settings.Address = address.Text()
		settings.Port = port.Text()
		settings.RBXL = rbxlFileTextBox.Text()
		callback(settings)
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
