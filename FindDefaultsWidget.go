package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"

func NewFindDefaultsWidget(parent widgets.QWidget_ITF, settings *DefaultsSettings, callback func(*DefaultsSettings)()) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Find default property values...")
	layout := widgets.NewQVBoxLayout()

	filesLabel := NewQLabelF("Property value locations:")
	files := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(files)
	standardModel.SetHorizontalHeaderLabels([]string{"File"})
	rootNode := standardModel.InvisibleRootItem()

	for _, file := range settings.Files {
		rootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("%s", file)})
	}
	files.SetSelectionMode(3)
	files.SetModel(standardModel)

	layout.AddWidget(filesLabel, 0, 0)
	layout.AddWidget(files, 0, 0)

	addButton := widgets.NewQPushButton2("Add", window)
	delButton := widgets.NewQPushButton2("Remove", window)

	addButton.ConnectPressed(func() {
		names := widgets.QFileDialog_GetOpenFileNames(window, "Find property dump...", "", "Roblox model files (*.rbxm)", "", 0)
		for _, file := range names {
			rootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("%s", file)})
		}
	})
	delButton.ConnectPressed(func() {
		for len(files.SelectionModel().SelectedIndexes()) > 0 {
			index := files.SelectionModel().SelectedIndexes()[0]
			standardModel.RemoveRow(index.Row(), index.Parent())
		}
	})

	layout.AddWidget(addButton, 0, 0)
	layout.AddWidget(delButton, 0, 0)

	okButton := widgets.NewQPushButton2("OK", window)
	okButton.ConnectPressed(func() {
		window.Destroy(true, true)
		settings.Files = make([]string, standardModel.RowCount(core.NewQModelIndex()))
		for i := 0; i < standardModel.RowCount(core.NewQModelIndex()); i++ {
			settings.Files[i] = standardModel.Item(i, 0).Data(0).ToString()
		}
		callback(settings)
	})
	layout.AddWidget(okButton, 0, 0)

	window.SetLayout(layout)
	window.Show()
}
