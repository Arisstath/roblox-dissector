package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewEditFilterWidget(parent widgets.QWidget_ITF, oldFilter string, oldUseExtraInfo bool, callback func(string, bool)) error {
	window := widgets.NewQWidget(parent, core.Qt__Window)
	window.SetWindowTitle("Input Lua filter script for packets...")

	layout := NewTopAlignLayout()
	extraInfo := widgets.NewQCheckBox2("Use extra info", nil)
	extraInfo.SetChecked(oldUseExtraInfo)
	layout.AddWidget(extraInfo, 0, 0)

	filterInput := widgets.NewQPlainTextEdit(nil)
	doc := filterInput.Document()
	font := doc.DefaultFont()
	font.SetFamily("Source Code Pro")
	font.SetStyleHint(gui.QFont__Monospace, gui.QFont__PreferDefault)
	doc.SetDefaultFont(font)
	filterInput.SetPlainText(oldFilter)
	layout.AddWidget(filterInput, 0, 0)

	okButton := widgets.NewQPushButton2("Apply", nil)
	layout.AddWidget(okButton, 0, 0)
	okButton.ConnectReleased(func() {
    	filterScript := filterInput.ToPlainText()
		useExtraInfo := extraInfo.CheckState() == core.Qt__Checked
		window.Close()
		callback(filterScript, useExtraInfo)
	})

	window.SetLayout(layout)
	window.Show()

	return nil
}
