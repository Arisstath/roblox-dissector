package main
import "github.com/therecipe/qt/widgets"

func NewStudioChooser(parent widgets.QWidget_ITF, settings *StudioSettings, callback func(*StudioSettings)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose Studio location...")
	layout := NewTopAlignLayout()

	fileLabel := NewLabel("Studio location:")
	fileTextBox := widgets.NewQLineEdit2(settings.Location, nil)
	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectReleased(func() {
		fileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find RobloxStudioBeta.exe...", "", "Roblox Studio (*.exe)", "", 0))
	})
	layout.AddWidget(fileLabel, 0, 0)
	layout.AddWidget(fileTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	flagsLabel := NewLabel("Command line flags:")
	flags := widgets.NewQLineEdit2(settings.Flags, nil)
	layout.AddWidget(flagsLabel, 0, 0)
	layout.AddWidget(flags, 0, 0)

	addressLabel := NewLabel("Server address:")
	address := widgets.NewQLineEdit2(settings.Address, nil)
	layout.AddWidget(addressLabel, 0, 0)
	layout.AddWidget(address, 0, 0)

	portLabel := NewLabel("Port number:")
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddWidget(portLabel, 0, 0)
	layout.AddWidget(port, 0, 0)

	rbxlFileLabel := NewLabel("RBXL location:")
	rbxlFileTextBox := widgets.NewQLineEdit2(settings.RBXL, nil)
	browseRBXLButton := widgets.NewQPushButton2("Browse...", nil)
	browseRBXLButton.ConnectReleased(func() {
		rbxlFileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find place file...", "", "Roblox place files (*.rbxl *.rbxlx)", "", 0))
	})
	layout.AddWidget(rbxlFileLabel, 0, 0)
	layout.AddWidget(rbxlFileTextBox, 0, 0)
	layout.AddWidget(browseRBXLButton, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
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

func NewPlayerChooser(parent widgets.QWidget_ITF, settings *PlayerSettings, callback func(*PlayerSettings)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Choose Player location...")
	layout := NewTopAlignLayout()
	
	fileLabel := NewLabel("Player location:")
	fileTextBox := widgets.NewQLineEdit2(settings.Location, nil)
	browseButton := widgets.NewQPushButton2("Browse...", nil)
	browseButton.ConnectReleased(func() {
		fileTextBox.SetText(widgets.QFileDialog_GetOpenFileName(window, "Find RobloxPlayerBeta.exe...", "", "Roblox Player (*.exe)", "", 0))
	})
	layout.AddWidget(fileLabel, 0, 0)
	layout.AddWidget(fileTextBox, 0, 0)
	layout.AddWidget(browseButton, 0, 0)

	flagsLabel := NewLabel("Command line flags:")
	flags := widgets.NewQLineEdit2(settings.Flags, nil)
	layout.AddWidget(flagsLabel, 0, 0)
	layout.AddWidget(flags, 0, 0)

	gameIDLabel := NewLabel("Game ID:")
	gameID := widgets.NewQLineEdit2(settings.GameID, nil)
	layout.AddWidget(gameIDLabel, 0, 0)
	layout.AddWidget(gameID, 0, 0)

	trackerIDLabel := NewLabel("Browser tracker ID:")
	trackerID := widgets.NewQLineEdit2(settings.TrackerID, nil)
	layout.AddWidget(trackerIDLabel, 0, 0)
	layout.AddWidget(trackerID, 0, 0)

	authTicketLabel := NewLabel("Auth ticket:")
	authTicket := widgets.NewQLineEdit2(settings.AuthTicket, nil)
	layout.AddWidget(authTicketLabel, 0, 0)
	layout.AddWidget(authTicket, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Destroy(true, true)
		settings.Location = fileTextBox.Text()
		settings.Flags = flags.Text()
		settings.GameID = gameID.Text()
		settings.TrackerID = trackerID.Text()
		settings.AuthTicket = authTicket.Text()
		callback(settings)
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
