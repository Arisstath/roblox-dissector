package main

import (
	"fmt"
	"os"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/xml"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func showChildren(rootNode *gui.QStandardItem, children []*datamodel.Instance) {
	for _, instance := range children {
		row := showReplicationInstance(instance, instance.Parent())
		if len(instance.Children) > 0 {
			childrenRootItem := NewQStandardItemF("%d children", len(instance.Children))
			showChildren(childrenRootItem, instance.Children)
			row[0].AppendRow([]*gui.QStandardItem{childrenRootItem, nil, nil, nil, nil, nil})
		}
		rootNode.AppendRow(row)
	}
}

func dumpScripts(instances []*rbxfile.Instance, i int) int {
	for _, instance := range instances {
		for name, property := range instance.Properties {
			thisType := property.Type()
			if thisType == rbxfile.TypeProtectedString {
				println("dumping protectedstring", instance.ClassName, name, thisType.String())
				file, err := os.Create(fmt.Sprintf("dumps/%s.%d", instance.GetFullName(), i))
				if err != nil {
					println(err.Error())
					continue
				}
				i++
				_, err = file.Write([]byte(instance.Properties[name].(rbxfile.ValueProtectedString)))
				if err != nil {
					println(err.Error())
					continue
				}
				err = file.Close()
				if err != nil {
					println(err.Error())
					continue
				}
			}
		}
		i = dumpScripts(instance.Children, i)
	}
	return i
}

func NewDataModelBrowser(context *peer.CommunicationContext, dataModel *datamodel.DataModel, defaultValues DefaultValues) {
	subWindow := widgets.NewQWidget(window, core.Qt__Window)
	subWindowLayout := widgets.NewQVBoxLayout2(subWindow)
	subWindowLayout.SetAlign(core.Qt__AlignTop)

	subWindow.SetWindowTitle("Data Model")

	writableClone := dataModel.ToRbxfile()

	takeSnapshotButton := widgets.NewQPushButton2("Save as RBXL...", nil)
	takeSnapshotButton.ConnectPressed(func() {
		location := widgets.QFileDialog_GetSaveFileName(subWindow, "Save as RBXLX...", "", "Roblox place files (*.rbxlx)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		defer writer.Close()
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		dumpScripts(writableClone.Instances, 0)

		err = xml.Serialize(writer, nil, writableClone)
		if err != nil {
			println("while serializing place:", err.Error())
			return
		}

		scriptData, err := os.OpenFile("dumps/scriptKeys", os.O_RDWR|os.O_CREATE, 0666)
		defer scriptData.Close()
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}

		_, err = fmt.Fprintf(scriptData, "Int 1: %d\nInt 2: %d", context.Int1, context.Int2)
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}
	})
	subWindowLayout.AddWidget(takeSnapshotButton, 0, 0)

	instanceList := widgets.NewQTreeView(subWindow)
	standardModel := NewProperSortModel(subWindow)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	showChildren(rootNode, dataModel.Instances)
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	subWindowLayout.AddWidget(instanceList, 0, 0)

	subWindow.Show()
}
