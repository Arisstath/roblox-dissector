package main

import (
	"fmt"
	"strconv"

	"github.com/olebedev/emitter"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type HTTPMessageViewer struct {
	*widgets.QWidget
	Headers      *widgets.QTableView
	HeadersModel *gui.QStandardItemModel
	Body         *widgets.QPlainTextEdit
}

func NewHTTPMessageViewer(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *HTTPMessageViewer {
	viewer := &HTTPMessageViewer{
		QWidget: widgets.NewQWidget(parent, flags),
	}
	layout := NewTopAlignLayout()

	headersTable := widgets.NewQTableView(viewer)
	viewer.Headers = headersTable
	viewer.HeadersModel = NewProperSortModel(headersTable)
	headersTable.SetModel(viewer.HeadersModel)
	layout.AddWidget(headersTable, 0, 0)

	body := widgets.NewQPlainTextEdit(viewer)
	layout.AddWidget(body, 0, 0)
	viewer.Body = body

	viewer.SetLayout(layout)
	return viewer
}

type HTTPLayerViewer struct {
	*widgets.QWidget
	RequestViewer  *HTTPMessageViewer
	ResponseViewer *HTTPMessageViewer
}

func (viewer *HTTPLayerViewer) Update(layer *HTTPLayer) {
	reqViewer := viewer.RequestViewer
	reqViewer.HeadersModel.Clear()
	reqViewer.HeadersModel.SetHorizontalHeaderLabels([]string{"Name", "Value"})
	reqViewer.HeadersModel.AppendRow([]*gui.QStandardItem{
		NewStringItem("<REQUEST>"),
		NewStringItem(fmt.Sprintf("%s %s", layer.Request.Method, layer.Request.URL.String())),
	})
	for name, valSet := range layer.Request.Header {
		for _, val := range valSet {
			reqViewer.HeadersModel.AppendRow([]*gui.QStandardItem{
				NewStringItem(name),
				NewStringItem(val),
			})
		}
	}
	reqViewer.Body.SetPlainText(string(layer.RequestBody))

	respViewer := viewer.ResponseViewer
	respViewer.HeadersModel.Clear()
	respViewer.HeadersModel.SetHorizontalHeaderLabels([]string{"Name", "Value"})
	respViewer.HeadersModel.AppendRow([]*gui.QStandardItem{
		NewStringItem("<RESPONSE>"),
		NewStringItem(layer.Response.Status),
	})
	for name, valSet := range layer.Response.Header {
		for _, val := range valSet {
			respViewer.HeadersModel.AppendRow([]*gui.QStandardItem{
				NewStringItem(name),
				NewStringItem(val),
			})
		}
	}
	respViewer.Body.SetPlainText(string(layer.ResponseBody))
}

func NewHTTPLayerViewer(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *HTTPLayerViewer {
	viewer := &HTTPLayerViewer{
		QWidget: widgets.NewQWidget(parent, flags),
	}
	layout := NewTopAlignLayout()
	viewer.RequestViewer = NewHTTPMessageViewer(viewer, 0)
	viewer.ResponseViewer = NewHTTPMessageViewer(viewer, 0)

	mainSplitter := widgets.NewQSplitter(viewer)
	mainSplitter.SetOrientation(core.Qt__Horizontal)
	mainSplitter.AddWidget(viewer.RequestViewer)
	mainSplitter.AddWidget(viewer.ResponseViewer)

	layout.AddWidget(mainSplitter, 0, 0)
	viewer.SetLayout(layout)

	return viewer
}

type HTTPViewer struct {
	*widgets.QWidget
	TreeView      *widgets.QTreeView
	StandardModel *gui.QStandardItemModel
	RootNode      *gui.QStandardItem
	Conversation  *Conversation
	LayerIndex    int
	Layers        map[int]*HTTPLayer // TODO: Use slice

	DefaultLayerViewer *HTTPLayerViewer
}

func NewHTTPViewer(parent widgets.QWidget_ITF, flags core.Qt__WindowType) *HTTPViewer {
	viewer := &HTTPViewer{
		QWidget: widgets.NewQWidget(parent, flags),
		Layers:  make(map[int]*HTTPLayer),
	}
	viewer.DefaultLayerViewer = NewHTTPLayerViewer(viewer, 0)
	layout := NewTopAlignLayout()

	standardModel := NewProperSortModel(viewer)
	standardModel.SetHorizontalHeaderLabels([]string{
		"Index", "Method", "URL", "Status",
	})
	viewer.StandardModel = standardModel
	treeView := widgets.NewQTreeView(viewer)
	treeView.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
	treeView.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	treeView.SetSortingEnabled(true)
	treeView.SetUniformRowHeights(true)
	treeView.SetModel(standardModel)
	treeView.SelectionModel().ConnectCurrentRowChanged(viewer.RowChanged)
	viewer.TreeView = treeView
	viewer.RootNode = standardModel.InvisibleRootItem()

	mainSplitter := widgets.NewQSplitter(viewer)
	mainSplitter.SetOrientation(core.Qt__Vertical)
	mainSplitter.AddWidget(treeView)
	mainSplitter.AddWidget(viewer.DefaultLayerViewer)
	layout.AddWidget(mainSplitter, 0, 0)

	viewer.SetLayout(layout)

	return viewer
}

func (viewer *HTTPViewer) RowChanged(index, _ *core.QModelIndex) {
	realSelectedValue, _ := strconv.Atoi(viewer.StandardModel.Item(index.Row(), 0).Data(0).ToString())
	if viewer.Layers[realSelectedValue] != nil {
		thisLayer := viewer.Layers[realSelectedValue]
		viewer.DefaultLayerViewer.Update(thisLayer)
	}
}

func (viewer *HTTPViewer) BindToConversation(conv *HTTPConversation) {
	conv.Layers().On("http", func(e *emitter.Event) {
		thisLayer := e.Args[0].(*HTTPLayer)
		thisURL := thisLayer.Request.URL
		// TODO: Hacky! Should create a copy of the URL
		thisURL.Host = thisLayer.OriginalHost
		thisURL.Scheme = thisLayer.OriginalScheme

		MainThreadRunner.RunOnMain(func() {
			viewer.Layers[viewer.LayerIndex] = thisLayer

			viewer.RootNode.AppendRow([]*gui.QStandardItem{
				NewUintItem(viewer.LayerIndex),
				NewStringItem(thisLayer.Request.Method),
				NewStringItem(thisURL.String()),
				NewStringItem(thisLayer.Response.Status),
			})
			viewer.LayerIndex++
		})
		<-MainThreadRunner.Wait
	}, emitter.Void)
}
