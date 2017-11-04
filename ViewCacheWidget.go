package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/core"
import "github.com/therecipe/qt/gui"
import "github.com/gskartwii/roblox-dissector/peer"

func NewCacheList(cache *peer.StringCache) widgets.QWidget_ITF {
	cacheWidget := widgets.NewQWidget(nil, 0)
	cacheLayout := widgets.NewQVBoxLayout()
	cacheList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Key", "Value"})

	rootNode := standardModel.InvisibleRootItem()
	for i := 0; i < 0x80; i++ {
		value := cache.Values[i]
		if value == nil {
			continue
		}

		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF("%d", i),
			NewQStandardItemF("%s", value.(string)),
		})
	}
	cacheList.SetModel(standardModel)
	cacheList.SetSelectionMode(0)
	cacheList.SetSortingEnabled(true)

	cacheLayout.AddWidget(cacheList, 0, 0)
	cacheLayout.AddWidget(NewQLabelF("%d", cache.LastWrite()), 0, 0)
	cacheWidget.SetLayout(cacheLayout)

	return cacheWidget
}

func NewViewCacheWidget(parent widgets.QWidget_ITF, context *peer.CommunicationContext) {
	subWindow := widgets.NewQWidget(window, core.Qt__Window)
	subWindow.SetWindowTitle("Cache Viewer")
	subWindowLayout := widgets.NewQVBoxLayout2(subWindow)
	tabWidget := widgets.NewQTabWidget(nil)

	tabWidget.AddTab(NewCacheList(&context.ClientCaches.String), "Client string cache")
	tabWidget.AddTab(NewCacheList(&context.ServerCaches.String), "Server string cache")
	tabWidget.AddTab(NewCacheList(&context.ClientCaches.Object), "Client object cache")
	tabWidget.AddTab(NewCacheList(&context.ServerCaches.Object), "Server object cache")

	subWindowLayout.AddWidget(tabWidget, 0, 0)
	subWindow.Show()
}
