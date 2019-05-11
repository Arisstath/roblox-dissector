package main

import (
	"github.com/therecipe/qt/widgets"
)

type FileBrowseLayout struct {
	*widgets.QHBoxLayout

	LineEdit     *widgets.QLineEdit
	BrowseButton *widgets.QPushButton
}

func (layout *FileBrowseLayout) FileName() string {
	return layout.LineEdit.Text()
}

func NewFileBrowseLayout(parent widgets.QWidget_ITF, directory bool, defaultText string, browseTitle string, browseFilter string) *FileBrowseLayout {
	layout := &FileBrowseLayout{
		QHBoxLayout:  widgets.NewQHBoxLayout2(parent),
		LineEdit:     widgets.NewQLineEdit2(defaultText, nil),
		BrowseButton: widgets.NewQPushButton2("Browse...", nil),
	}

	layout.BrowseButton.ConnectReleased(func() {
		var newText string
		if directory {
			newText = widgets.QFileDialog_GetExistingDirectory(parent, browseTitle, "", 0)
		} else {
			newText = widgets.QFileDialog_GetOpenFileName(parent, browseTitle, "", browseFilter, "", 0)
		}
		layout.LineEdit.SetText(newText)
	})

	layout.AddWidget(layout.LineEdit, 0, 0)
	layout.AddWidget(layout.BrowseButton, 0, 0)

	return layout
}
