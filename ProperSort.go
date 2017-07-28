package main
import "github.com/therecipe/qt/gui"

func NewProperSortModel(parent core.QObject_ITF) *gui.QStandardItemModel {
	return gui.NewQStandardItemModel(parent)
}
