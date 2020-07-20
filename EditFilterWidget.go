package main

import (
	"io/ioutil"

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

	buttonRow := widgets.NewQHBoxLayout()
	buttonRow.SetAlign(core.Qt__AlignLeft)

	okButton := widgets.NewQPushButton2("Apply", nil)
	buttonRow.AddWidget(okButton, 0, 0)
	okButton.ConnectReleased(func() {
		filterScript := filterInput.ToPlainText()
		useExtraInfo := extraInfo.CheckState() == core.Qt__Checked
		window.Close()
		callback(filterScript, useExtraInfo)
	})
	loadFromFileButton := widgets.NewQPushButton2("Open file...", nil)
	buttonRow.AddWidget(loadFromFileButton, 0, 0)
	loadFromFileButton.ConnectReleased(func() {
		file := widgets.QFileDialog_GetOpenFileName(window, "Open filter script file", "", "Lua filter scripts (*.lua)", "", 0)
		if file != "" {
			script, err := ioutil.ReadFile(file)
			if err != nil {
				println("failed to open file:", err.Error())
				return
			}
			filterInput.SetPlainText(string(script))
		}
	})

	buttonRowWidget := widgets.NewQWidget(nil, 0)
	buttonRowWidget.SetLayout(buttonRow)
	layout.AddWidget(buttonRowWidget, 0, 0)

	window.SetLayout(layout)
	window.Show()

	return nil
}
