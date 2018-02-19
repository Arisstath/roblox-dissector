package main
import "github.com/therecipe/qt/widgets"
import "github.com/gskartwii/roblox-dissector/peer"
import "strconv"

func NewClientStartWidget(parent widgets.QWidget_ITF, settings *peer.CustomClient, callback func(uint32)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start server...")
	layout := widgets.NewQVBoxLayout()
	trackerIDLabel := NewQLabelF("Browser tracker ID:")
	trackerID := widgets.NewQLineEdit2(strconv.Itoa(int(settings.BrowserTrackerId)), nil)
	layout.AddWidget(trackerIDLabel, 0, 0)
	layout.AddWidget(trackerID, 0, 0)

	goldenHashLabel := NewQLabelF("Golden hash:")
	goldenHash := widgets.NewQLineEdit2(strconv.Itoa(int(settings.GoldenHash)), nil)
	layout.AddWidget(goldenHashLabel, 0, 0)
	layout.AddWidget(goldenHash, 0, 0)

	osPlatformLabel := NewQLabelF("OS platform:")
	osPlatform := widgets.NewQLineEdit2(settings.OsPlatform, nil)
	layout.AddWidget(osPlatformLabel, 0, 0)
	layout.AddWidget(osPlatform, 0, 0)

	dataModelHashLabel := NewQLabelF("DataModel hash:")
	dataModelHash := widgets.NewQLineEdit2(settings.DataModelHash, nil)
	layout.AddWidget(dataModelHashLabel, 0, 0)
	layout.AddWidget(dataModelHash, 0, 0)

	securityKeyLabel := NewQLabelF("Security key:")
	securityKey := widgets.NewQLineEdit2(settings.SecurityKey, nil)
	layout.AddWidget(securityKeyLabel, 0, 0)
	layout.AddWidget(securityKey, 0, 0)

	placeIdLabel := NewQLabelF("Place id:")
	placeId := widgets.NewQLineEdit2("0", nil)
	layout.AddWidget(placeIdLabel, 0, 0)
	layout.AddWidget(placeId, 0, 0)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectPressed(func() {
		window.Destroy(true, true)
		browserTrackerId, _ := strconv.Atoi(trackerID.Text())
		settings.BrowserTrackerId = uint64(browserTrackerId)
		goldenHashVal, _ := strconv.Atoi(goldenHash.Text())
		settings.GoldenHash = uint32(goldenHashVal)
		settings.OsPlatform = osPlatform.Text()
		settings.DataModelHash = dataModelHash.Text()
		settings.SecurityKey = securityKey.Text()
		placeIdVal, _ := strconv.Atoi(placeId.Text())
		callback(uint32(placeIdVal))
	})
	layout.AddWidget(startButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
