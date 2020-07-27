package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dreadl0ck/gopcap"
	"github.com/google/gopacket/pcap"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func invalidUi(name string) error {
	return errors.New("invalid ui (" + name + ")")
}

type DissectorWindow struct {
	*gtk.Window

	tabs                 *gtk.Notebook
	forgetAcksItem       *gtk.CheckMenuItem
	tabIndexToSession    []*CaptureSession
	tabIndexToListViewer []*PacketListViewer
	sessionRefCount      map[*CaptureSession]uint

	stopButton            *gtk.ToolButton
	pauseButton           *gtk.ToolButton
	browseDataModelButton *gtk.ToolButton
}

func ShowError(wdg gtk.IWidget, err error, extrainfo string) {
	widget := wdg.ToWidget()
	parentWindow, topLevelErr := widget.GetToplevel()
	if topLevelErr != nil {
		println("failed to find parent window:", topLevelErr.Error())
		return
	}
	dialog := gtk.MessageDialogNew(
		parentWindow.(gtk.IWindow),
		gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_OK,
		"%s: %s",
		extrainfo,
		err.Error(),
	)
	dialog.SetTitle("Error")
	dialog.SetIconFromFile("res/app-icon.ico")
	dialog.Connect("response", (*gtk.MessageDialog).Destroy)
	dialog.ShowAll()
	dialog.Run()
}

func (win *DissectorWindow) ShowCaptureError(err error, extrainfo string) {
	ShowError(win, err, extrainfo)
}

func (win *DissectorWindow) AppendClosablePage(title string, session *CaptureSession, listViewer *PacketListViewer) {
	titleLabel, err := gtk.LabelNew(title)
	if err != nil {
		win.ShowCaptureError(err, "Accepting new listviewer")
		return
	}
	titleHBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 4)
	if err != nil {
		win.ShowCaptureError(err, "Accepting new listviewer")
		return
	}
	titleHBox.PackStart(titleLabel, true, true, 0)

	closeButton, err := gtk.ButtonNew()
	if err != nil {
		win.ShowCaptureError(err, "Accepting new listviewer")
		return
	}
	buttonImg, err := gtk.ImageNewFromIconName("edit-delete", gtk.ICON_SIZE_BUTTON)
	closeButton.SetRelief(gtk.RELIEF_NONE)
	closeButton.SetFocusOnClick(false)
	closeButton.Add(buttonImg)
	titleHBox.PackEnd(closeButton, false, false, 0)
	titleHBox.ShowAll()
	win.sessionRefCount[session] += 1

	closeButton.Connect("clicked", func() {
		pageNum := win.tabs.PageNum(listViewer.mainWidget)
		if pageNum == -1 {
			return
		}
		win.tabs.RemovePage(pageNum)

		oldSession := win.tabIndexToSession[pageNum]
		win.tabIndexToListViewer = append(win.tabIndexToListViewer[:pageNum], win.tabIndexToListViewer[pageNum+1:]...)
		win.tabIndexToSession = append(win.tabIndexToSession[:pageNum], win.tabIndexToSession[pageNum+1:]...)
		win.sessionRefCount[oldSession] -= 1

		if win.sessionRefCount[oldSession] == 0 {
			delete(win.sessionRefCount, oldSession)
			oldSession.StopCapture()
		}
		win.UpdateActionsEnabled()
	})

	win.tabIndexToListViewer = append(win.tabIndexToListViewer, listViewer)
	win.tabIndexToSession = append(win.tabIndexToSession, session)
	win.tabs.AppendPage(listViewer.mainWidget, titleHBox)
}

func (win *DissectorWindow) UpdateActionsForPage(curPage int) {
	if curPage == -1 {
		pauseButtonIcon, err := win.pauseButton.GetIconWidget()
		if err != nil {
			println("error while switching pause button")
			return
		}
		var iconName = "media-playback-pause-symbolic"
		pauseButtonIcon.(*gtk.Image).SetFromIconName(iconName, gtk.ICON_SIZE_BUTTON)
		win.stopButton.SetSensitive(false)
		win.pauseButton.SetSensitive(false)
		win.browseDataModelButton.SetSensitive(false)
		return
	}

	curSession := win.tabIndexToSession[curPage]
	curViewer := win.tabIndexToListViewer[curPage]
	win.stopButton.SetSensitive(curSession.IsCapturing)
	win.pauseButton.SetSensitive(curSession.IsCapturing)
	win.browseDataModelButton.SetSensitive(true)

	pauseButtonIcon, err := win.pauseButton.GetIconWidget()
	if err != nil {
		println("error while switching pause button")
		return
	}
	var iconName = "media-playback-pause-symbolic"
	if !curViewer.updatePassthrough && curSession.IsCapturing {
		iconName = "media-playback-start-symbolic"
	}
	pauseButtonIcon.(*gtk.Image).SetFromIconName(iconName, gtk.ICON_SIZE_BUTTON)
}

func (win *DissectorWindow) UpdateActionsEnabled() {
	win.UpdateActionsForPage(win.tabs.GetCurrentPage())
}

func (win *DissectorWindow) StopClicked() {
	curPage := win.tabs.GetCurrentPage()
	currSession := win.tabIndexToSession[curPage]
	currSession.StopCapture()
	win.UpdateActionsEnabled()
}

func (win *DissectorWindow) PauseClicked() {
	curPage := win.tabs.GetCurrentPage()
	currViewer := win.tabIndexToListViewer[curPage]
	currViewer.ToggleUpdatePassthrough()
	win.UpdateActionsEnabled()
}

func (win *DissectorWindow) BrowseDataModelClicked() {
	// nop
}

func (win *DissectorWindow) CaptureFromPcapDevice(name string) {
	handle, err := pcap.OpenLive(name, 2000, false, 1*time.Second)
	if err != nil {
		println("error starting capture", err.Error())
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	context, cancelFunc := context.WithCancel(context.Background())
	session, err := NewCaptureSession(name, cancelFunc, func(session *CaptureSession, listViewer *PacketListViewer, err error) {
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		listViewer.mainWidget.ShowAll()
		win.AppendClosablePage(listViewer.title, session, listViewer)

		windowHeight := win.GetAllocatedHeight()
		paneHeight := int(0.6 * float64(windowHeight))
		listViewer.mainWidget.SetPosition(paneHeight)
		listViewer.mainWidget.SetWideHandle(true)
	})
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	session.ForgetAcks = win.forgetAcksItem.GetActive()

	go func() {
		err = CaptureFromHandle(context, session, handle)
		if err != nil {
			win.ShowCaptureError(err, "Starting capture")
			return
		}
	}()
}

func (win *DissectorWindow) CaptureFromFile(filename string) {
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}

	progressDialog, err := gtk.DialogNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressDialog.SetTitle("File capture in progress")
	contentArea, err := progressDialog.GetContentArea()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressBar, err := gtk.ProgressBarNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating progress dialog")
		return
	}
	progressBar.Pulse()
	progressBar.SetShowText(true)
	progressBar.SetText("Scanning for packets...")
	contentArea.Add(progressBar)
	progressDialog.SetDeletable(false) // Hide "close" button
	progressDialog.SetModal(true)
	progressDialog.SetSizeRequest(200, 32)
	progressDialog.ShowAll()

	context, cancelFunc := context.WithCancel(context.Background())
	session, err := NewCaptureSession(filename, cancelFunc, func(session *CaptureSession, listViewer *PacketListViewer, err error) {
		if err != nil {
			win.ShowCaptureError(err, "Accepting new listviewer")
			return
		}
		listViewer.mainWidget.ShowAll()
		win.AppendClosablePage(listViewer.title, session, listViewer)

		windowHeight := win.GetAllocatedHeight()
		paneHeight := int(0.6 * float64(windowHeight))
		listViewer.mainWidget.SetPosition(paneHeight)
		listViewer.mainWidget.SetWideHandle(true)
	})
	if err != nil {
		win.ShowCaptureError(err, "Starting capture")
		return
	}
	session.ForgetAcks = win.forgetAcksItem.GetActive()

	countPackets := 0.0
	frac := 0.0

	lastUpdate := time.Now()
	session.ProgressCallback = func(progress int) {
		if progress == -1 {
			progressDialog.Close()
			win.UpdateActionsEnabled()
			return
		}

		now := time.Now()
		if now.Sub(lastUpdate) < 500*time.Millisecond {
			return
		}
		lastUpdate = now

		frac = float64(progress) / countPackets
		progressBar.SetFraction(frac)
		progressBar.SetText(fmt.Sprintf("Reading packets: %.1f %%", frac*100))

		// Force redraw of this progress bar
		// We only force this every 100 ms, so it shouldn't be too bad
		glib.MainContextDefault().Iteration(false)
	}

	go func() {
		count, err := gopcap.Count(filename)
		if err != nil {
			glib.IdleAdd(func() bool {
				progressDialog.Close()
				return false
			})
		} else {
			countPackets = float64(count)
		}

		err = CaptureFromHandle(context, session, handle)
		session.ReportDone()
		if err != nil {
			win.ShowCaptureError(err, "Starting capture")
			return
		}
	}()
}

func (win *DissectorWindow) PromptCaptureFromFile() {
	chooser, err := gtk.FileChooserNativeDialogNew("Choose PCAP file", win, gtk.FILE_CHOOSER_ACTION_OPEN, "Choose", "Cancel")
	if err != nil {
		win.ShowCaptureError(err, "Making chooser")
		return
	}
	filter, err := gtk.FileFilterNew()
	if err != nil {
		win.ShowCaptureError(err, "Creating filter")
		return
	}
	filter.AddPattern("*.pcap")
	filter.SetName("PCAP network capture files (*.pcap)")
	chooser.AddFilter(filter)
	resp := chooser.NativeDialog.Run()
	if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
		filename := chooser.GetFilename()
		win.CaptureFromFile(filename)
	}
}

func (win *DissectorWindow) PromptCaptureLive() {
	err := PromptInterfaceName(win.CaptureFromPcapDevice)
	if err != nil {
		win.ShowCaptureError(err, "Making interface chooser")
	}
}

func NewDissectorWindow() (*gtk.Window, error) {
	winBuilder, err := gtk.BuilderNewFromFile("res/dissectorwindow.ui")
	if err != nil {
		return nil, err
	}
	win, err := winBuilder.GetObject("dissectorwindow")
	if err != nil {
		return nil, err
	}

	wind, ok := win.(*gtk.Window)
	if !ok {
		return nil, invalidUi("mainwindow")
	}
	wind.SetTitle("Sala")
	wind.SetDefaultSize(800, 640)

	dwin := &DissectorWindow{
		Window:          wind,
		sessionRefCount: make(map[*CaptureSession]uint),
	}

	tabs, err := winBuilder.GetObject("conversationtabs")
	if err != nil {
		return nil, err
	}

	tabsNotebook, ok := tabs.(*gtk.Notebook)
	if !ok {
		return nil, invalidUi("convtabs")
	}
	dwin.tabs = tabsNotebook
	tabsNotebook.Connect("switch-page", func(_ *gtk.Notebook, _ gtk.IWidget, num int) {
		dwin.UpdateActionsForPage(num)
	})

	fromFileItem, err := winBuilder.GetObject("fromfileitem")
	if err != nil {
		return nil, err
	}
	fromFileMenuItem, ok := fromFileItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("fromfileitem")
	}
	fromFileMenuItem.Connect("activate", dwin.PromptCaptureFromFile)
	fromFileButton_, err := winBuilder.GetObject("fromfilebutton")
	if err != nil {
		return nil, err
	}
	fromFileButton, ok := fromFileButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("fromfilebutton")
	}
	fromFileButton.Connect("clicked", dwin.PromptCaptureFromFile)

	fromLiveItem, err := winBuilder.GetObject("frominterfaceitem")
	if err != nil {
		return nil, err
	}
	fromLiveMenuItem, ok := fromLiveItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("frominterfaceitem")
	}
	fromLiveMenuItem.Connect("activate", dwin.PromptCaptureLive)
	fromLiveButton_, err := winBuilder.GetObject("frominterfacebutton")
	if err != nil {
		return nil, err
	}
	fromLiveButton, ok := fromLiveButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("frominterfacebutton")
	}
	fromLiveButton.Connect("clicked", dwin.PromptCaptureLive)

	stopButton_, err := winBuilder.GetObject("stopcapturebutton")
	if err != nil {
		return nil, err
	}
	stopButton, ok := stopButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("stopcapturebutton")
	}
	stopButton.Connect("clicked", dwin.StopClicked)
	dwin.stopButton = stopButton
	pauseButton_, err := winBuilder.GetObject("pauseupdatebutton")
	if err != nil {
		return nil, err
	}
	pauseButton, ok := pauseButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("pauseupdatebutton")
	}
	pauseButton.Connect("clicked", dwin.PauseClicked)
	dwin.pauseButton = pauseButton
	browseDataModelButton_, err := winBuilder.GetObject("viewdatamodelbutton")
	if err != nil {
		return nil, err
	}
	browseDataModelButton, ok := browseDataModelButton_.(*gtk.ToolButton)
	if !ok {
		return nil, invalidUi("viewdatamodelbutton")
	}
	browseDataModelButton.Connect("clicked", dwin.BrowseDataModelClicked)
	dwin.browseDataModelButton = browseDataModelButton
	dwin.UpdateActionsEnabled()

	forgetAcksItem_, err := winBuilder.GetObject("forgetacksitem")
	if err != nil {
		return nil, err
	}
	forgetAcksItem, ok := forgetAcksItem_.(*gtk.CheckMenuItem)
	if !ok {
		return nil, invalidUi("forgetacksitem")
	}
	dwin.forgetAcksItem = forgetAcksItem

	divertItem, err := winBuilder.GetObject("fromdivertitem")
	if err != nil {
		return nil, err
	}
	divertMenuItem, ok := divertItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("divertmenuitem")
	}
	divertMenuItem.Connect("activate", func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		session, err := NewCaptureSession("<DIVERT>", cancelFunc, func(session *CaptureSession, listViewer *PacketListViewer, err error) {
			if err != nil {
				dwin.ShowCaptureError(err, "Accepting new listviewer")
				return
			}

			listViewer.mainWidget.ShowAll()
			dwin.AppendClosablePage(listViewer.title, session, listViewer)

			windowHeight := dwin.GetAllocatedHeight()
			paneHeight := int(0.6 * float64(windowHeight))
			listViewer.mainWidget.SetPosition(paneHeight)
			listViewer.mainWidget.SetWideHandle(true)
		})
		if err != nil {
			dwin.ShowCaptureError(err, "Starting WinDivert proxy")
		}
		err = CaptureFromDivert(ctx, session)
		if err != nil {
			dwin.ShowCaptureError(err, "Starting WinDivert proxy")
		}
	})

	aboutDialogItem, err := winBuilder.GetObject("aboutitem")
	if err != nil {
		return nil, err
	}
	aboutDialogMenuItem, ok := aboutDialogItem.(*gtk.MenuItem)
	if !ok {
		return nil, invalidUi("aboutitem")
	}
	aboutDialogMenuItem.Connect("activate", func() {
		dialog, err := gtk.AboutDialogNew()
		if err != nil {
			dwin.ShowCaptureError(err, "Show about dialog")
		}
		dialog.SetProgramName("Sala")
		dialog.SetVersion("v0.7.6")
		dialog.SetCopyright("© 2017 - 2020 Aleksi Hannula\nLicensed under the MIT license.")
		dialog.SetComments(`Codename "Maailman salaisuudet".

Sala is tool for dissecting Roblox network packets.`)
		dialog.SetWebsite("https://github.com/Gskartwii/roblox-dissector")
		dialog.SetWebsiteLabel("GitHub repository")
		dialog.SetAuthors([]string{"Aleksi Hannula", "Arisstath", "Alureon"})
		logo, err := gdk.PixbufNewFromFile("res/app-icon.ico")
		if err != nil {
			dwin.ShowCaptureError(err, "Show about dialog")
		}
		dialog.SetLogo(logo)
		dialog.SetIconFromFile("res/app-icon.ico")
		dialog.Connect("response", (*gtk.AboutDialog).Destroy)
		dialog.Run()
	})

	wind.SetIconFromFile("res/app-icon.ico")

	return wind, nil
}
