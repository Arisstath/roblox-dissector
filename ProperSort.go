package main
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"
import "strconv"
import "strings"

type PropertSortItem struct {
	gui.QStandardItem
}

func NewProperSortItem(content string) *PropertSortItem {
	return &PropertSortItem{gui.NewQStandardItem2(content)}
}

func (m *ProperSortItem) LessThan(source_left core.QModelIndex_ITF, source_right core.QModelIndex_ITF) bool {
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
