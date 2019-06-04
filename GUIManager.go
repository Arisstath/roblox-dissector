package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Cheating
type mainThreadHelper struct {
	core.QObject

	_    func()       `constructor:"init"`
	_    func(func()) `signal:"runOnMain"`
	Wait chan struct{}
}

func (helper *mainThreadHelper) init() {
	// send doesn't need to block, only receive
	// that's why we can make the cap 1
	// this also helps prevent deadlocks
	helper.Wait = make(chan struct{}, 1)
}

func (helper *mainThreadHelper) runOnMain(f func()) {
	f()
	helper.Wait <- struct{}{}
}

var MainThreadRunner = NewMainThreadHelper(nil)

func init() {
	MainThreadRunner.ConnectRunOnMain(MainThreadRunner.runOnMain)
}

func NewTopAlignLayout() *widgets.QVBoxLayout {
	layout := widgets.NewQVBoxLayout()
	layout.SetAlign(core.Qt__AlignTop)
	return layout
}
func NewLabel(content string) *widgets.QLabel {
	return widgets.NewQLabel2(content, nil, 0)
}
func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewStringItem(content string) *gui.QStandardItem {
	ret := gui.NewQStandardItem2(content)
	ret.SetEditable(false)
	return ret
}
func NewIntItem(content interface{}) *gui.QStandardItem {
	var normalized int
	switch content.(type) {
	case int:
		normalized = content.(int)
	case int8:
		normalized = int(content.(int8))
	case int16:
		normalized = int(content.(int16))
	case int32:
		normalized = int(content.(int32))
	case int64:
		normalized = int(content.(int64))
	case uint:
		normalized = int(content.(uint))
	case uint8:
		normalized = int(content.(uint8))
	case uint16:
		normalized = int(content.(uint16))
	case uint32:
		normalized = int(content.(uint32))
	case uint64:
		normalized = int(content.(uint64))
	}

	ret := gui.NewQStandardItem()
	ret.SetData(core.NewQVariant7(normalized), 0)
	ret.SetEditable(false)
	return ret
}
func NewUintItem(content interface{}) *gui.QStandardItem {
	var normalized uint
	switch content.(type) {
	case int:
		normalized = uint(content.(int))
	case int8:
		normalized = uint(content.(int8))
	case int16:
		normalized = uint(content.(int16))
	case int32:
		normalized = uint(content.(int32))
	case int64:
		normalized = uint(content.(int64))
	case uint:
		normalized = content.(uint)
	case uint8:
		normalized = uint(content.(uint8))
	case uint16:
		normalized = uint(content.(uint16))
	case uint32:
		normalized = uint(content.(uint32))
	case uint64:
		normalized = uint(content.(uint64))
	}

	ret := gui.NewQStandardItem()
	ret.SetData(core.NewQVariant8(normalized), 0)
	ret.SetEditable(false)
	return ret
}
func NewQStandardItemF(format string, args ...interface{}) *gui.QStandardItem {
	ret := gui.NewQStandardItem2(fmt.Sprintf(format, args...))
	ret.SetEditable(false)
	return ret
}

func paintItems(row []*gui.QStandardItem, color *gui.QColor) {
	for i := 0; i < len(row); i++ {
		row[i].SetBackground(gui.NewQBrush3(color, core.Qt__SolidPattern))
	}
}

func showCritical(title, message string) {
	widgets.QMessageBox_Critical(nil, title, message, widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
}

func GUIMain() {
	widgets.NewQApplication(len(os.Args), os.Args)
	window := NewDissectorWindow(nil, 0)
	window.ShowMaximized()

	joinFlag := flag.String("join", "", "roblox-dissector:<authTicket>:<placeID>:<browserTrackerID>")
	flag.Parse()
	if *joinFlag != "" {
		println("Received protocol invocation?")
		window.StartClient(*joinFlag)
	}
	openFile := flag.Arg(0)
	// React to command line arg
	if openFile != "" {
		window.CaptureFromFile(openFile, false)
	}

	widgets.QApplication_Exec()
}
