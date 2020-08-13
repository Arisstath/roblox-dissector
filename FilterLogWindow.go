package main

import (
	"github.com/gotk3/gotk3/gtk"
)

type FilterLogWindow struct {
	win     *gtk.Window
	textBuf *gtk.TextBuffer
}

func NewFilterLogWindow(title string) (*FilterLogWindow, error) {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, err
	}
	win.SetTitle(title)

	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}

	textBuf, err := gtk.TextBufferNew(nil)
	if err != nil {
		return nil, err
	}
	logView, err := gtk.TextViewNewWithBuffer(textBuf)
	if err != nil {
		return nil, err
	}
	logView.SetProperty("monospace", true)
	logView.SetEditable(false)
	logView.SetVExpand(true)
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.Add(logView)
	box.Add(scrolled)

	clearButton, err := gtk.ButtonNewWithLabel("Clear")
	if err != nil {
		return nil, err
	}
	clearButton.Connect("clicked", func() {
		textBuf.Delete(textBuf.GetStartIter(), textBuf.GetEndIter())
	})
	box.Add(clearButton)

	win.Add(box)
	return &FilterLogWindow{
		win:     win,
		textBuf: textBuf,
	}, nil

}

func (logWin *FilterLogWindow) AppendLog(log string) {
	logWin.textBuf.Insert(logWin.textBuf.GetEndIter(), log)
}

func (logWin *FilterLogWindow) Show() {
	logWin.win.ShowAll()
}

func (logWin *FilterLogWindow) Hide() {
	logWin.win.Hide()
}
