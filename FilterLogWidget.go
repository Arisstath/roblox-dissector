package main

import (
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type FilterLogWidget struct {
    *widgets.QWidget
    LogBox *widgets.QPlainTextEdit
}

func NewFilterLogWidget(parent widgets.QWidget_ITF, title string) *FilterLogWidget {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Filter log: " + title)
	layout := NewTopAlignLayout()

	logBox := widgets.NewQPlainTextEdit(nil)
	doc := logBox.Document()
	font := doc.DefaultFont()
	font.SetFamily("Source Code Pro")
	font.SetStyleHint(gui.QFont__Monospace, gui.QFont__PreferDefault)
	doc.SetDefaultFont(font)
	layout.AddWidget(logBox, 0, 0)

	clearButton := widgets.NewQPushButton2("Clear", nil)
	clearButton.ConnectReleased(func() {
		logBox.SetPlainText("")
	})
	layout.AddWidget(clearButton, 0, 0)

	window.SetLayout(layout)

	return &FilterLogWidget{
    	QWidget: window,

		LogBox: logBox,
	}
}

func (widget *FilterLogWidget) AppendLog(log string) {
    widget.LogBox.AppendPlainText(log)
}
