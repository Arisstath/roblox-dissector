package main

/*
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

	for name, caches := range context.NamedCaches {
		tabWidget.AddTab(NewCacheList(&caches.String), fmt.Sprintf("%s string cache", name))
		tabWidget.AddTab(NewCacheList(&caches.Object), fmt.Sprintf("%s object cache", name))
	}

	subWindowLayout.AddWidget(tabWidget, 0, 0)
	subWindow.Show()
}
*/
