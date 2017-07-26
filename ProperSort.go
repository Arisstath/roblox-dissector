package main
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"
import "strconv"
import "strings"

type ProperSortModel struct {
	core.QSortFilterProxyModel
	source *gui.QStandardItemModel
}

func NewProperSortModel(parent core.QObject_ITF) *ProperSortModel {
	source := gui.NewQStandardItemModel(nil)
	new := &ProperSortModel{
		*core.NewQSortFilterProxyModel(parent),
		source,
	}
	new.SetSourceModel(source)
	new.SetDynamicSortFilter(true)
	new.SetSortRole(0)
	return new
}

func (m *ProperSortModel) LessThan(source_left core.QModelIndex_ITF, source_right core.QModelIndex_ITF) bool {
	println("Proper sorting called")
	leftData := m.source.Data(source_left, 0)
	rightData := m.source.Data(source_right, 0)
	if leftData.Type() != core.QVariant__String {
		println("Don't know how to sort:", leftData.Type())
		return true
	} else {
		leftInt, err1 := strconv.Atoi(leftData.ToString())
		rightInt, err2 := strconv.Atoi(rightData.ToString())
		if err1 != nil || err2 != nil {
			return strings.Compare(leftData.ToString(), rightData.ToString()) == -1
		}
		return leftInt < rightInt
	}
}

func (m *ProperSortModel) SetHorizontalHeaderLabels(labels []string) {
	m.source.SetHorizontalHeaderLabels(labels)
}

func (m *ProperSortModel) InvisibleRootItem() *gui.QStandardItem {
	return m.source.InvisibleRootItem()
}
