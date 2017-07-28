package main
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"

func NewProperSortModel(parent core.QObject_ITF) *gui.QStandardItemModel {
	return gui.NewQStandardItemModel(parent)
}
