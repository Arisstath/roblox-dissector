// +build !pango_1_44

package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

func newWrappingLabelF(fmtS string, rest ...interface{}) (*gtk.Label, error) {
	label, err := gtk.LabelNew(fmt.Sprintf(fmtS, rest...))
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)

	label.SetLineWrap(true)
	label.SetLineWrapMode(pango.WRAP_WORD_CHAR)

	return label, nil
}
