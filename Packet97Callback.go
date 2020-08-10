package main

import (
	"fmt"
	"os"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket97(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet97Layer)

	progress := 0
	maxProgress := len(MainLayer.Schema.Enums) + len(MainLayer.Schema.Instances)
	progressDialog := widgets.NewQProgressDialog2(fmt.Sprintf("Building schema view (%d enums, %d classes)", len(MainLayer.Schema.Enums), len(MainLayer.Schema.Instances)), "", 0, maxProgress, nil, 0)
	progressDialog.SetCancelButton(nil)
	progressDialog.SetWindowFilePath("Building schema view")

	labelForEnumSchema := NewQLabelF("Enum schema (%d entries):", len(MainLayer.Schema.Enums))
	layerLayout.AddWidget(labelForEnumSchema, 0, 0)

	enumSchemaList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Size"})

	enumSchemaRootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.Schema.Enums {
		nameItem := NewStringItem(item.Name)
		sizeItem := NewUintItem(item.BitSize)
		enumSchemaRootNode.AppendRow([]*gui.QStandardItem{nameItem, sizeItem})
		progress++
		progressDialog.SetValue(progress)
	}

	enumSchemaList.SetModel(standardModel)
	enumSchemaList.SetSelectionMode(0)
	enumSchemaList.SetSortingEnabled(true)
	enumSchemaList.SortByColumn(0, core.Qt__AscendingOrder)
	layerLayout.AddWidget(enumSchemaList, 0, 0)

	labelForInstanceSchema := NewQLabelF("Instance schema (%d entries):", len(MainLayer.Schema.Instances))
	layerLayout.AddWidget(labelForInstanceSchema, 0, 0)
	instanceSchemaList := widgets.NewQTreeView(nil)
	instanceModel := NewProperSortModel(nil)
	instanceModel.SetHorizontalHeaderLabels([]string{"Name", "Type"})
	instanceSchemaRootNode := instanceModel.InvisibleRootItem()

	for _, item := range MainLayer.Schema.Instances {
		nameItem := NewStringItem(item.Name)
		instanceRow := []*gui.QStandardItem{nameItem, nil}

		propertySchemaItem := NewQStandardItemF("Property schema (%d entries)", len(item.Properties))

		for _, property := range item.Properties {
			propertyNameItem := NewStringItem(property.Name)
			propertyTypeItem := NewStringItem(property.TypeString)

			propertyRow := []*gui.QStandardItem{propertyNameItem, propertyTypeItem}
			propertySchemaItem.AppendRow(propertyRow)
		}
		nameItem.AppendRow([]*gui.QStandardItem{propertySchemaItem})

		eventSchemaItem := NewQStandardItemF("Event schema (%d entries)", len(item.Events))
		nameItem.AppendRow([]*gui.QStandardItem{eventSchemaItem})

		for _, event := range item.Events {
			eventNameItem := NewQStandardItemF("%s (%d arguments)", event.Name, len(event.Arguments))

			eventRow := []*gui.QStandardItem{eventNameItem, nil}

			for index, thisArgument := range event.Arguments {
				eventArgumentNameItem := NewQStandardItemF("Event argument %d", index)
				eventArgumentTypeItem := NewStringItem(thisArgument.TypeString)

				eventSubIntRow := []*gui.QStandardItem{eventArgumentNameItem, eventArgumentTypeItem}
				eventNameItem.AppendRow(eventSubIntRow)
			}

			eventSchemaItem.AppendRow(eventRow)
		}

		instanceSchemaRootNode.AppendRow(instanceRow)
		progress++
		progressDialog.SetValue(progress)
	}

	instanceSchemaList.SetModel(instanceModel)
	instanceSchemaList.SetSelectionMode(0)
	instanceSchemaList.SetSortingEnabled(true)
	instanceSchemaList.SortByColumn(0, core.Qt__AscendingOrder)
	layerLayout.AddWidget(instanceSchemaList, 0, 0)

	labelForPrefixes := NewQLabelF("Preshared Content prefixes (%d entries):", len(MainLayer.Schema.ContentPrefixes))
	layerLayout.AddWidget(labelForPrefixes, 0, 0)
	prefixList := widgets.NewQTreeView(nil)
	prefixModel := NewProperSortModel(nil)
	prefixModel.SetHorizontalHeaderLabels([]string{"Name"})
	prefixRootNode := prefixModel.InvisibleRootItem()
	for _, prefix := range MainLayer.Schema.ContentPrefixes {
		prefixRootNode.AppendRow([]*gui.QStandardItem{NewQStandardItemF("%s", prefix)})
	}
	prefixList.SetModel(prefixModel)
	prefixList.SetSelectionMode(0)
	layerLayout.AddWidget(prefixList, 0, 0)

	dumpButton := widgets.NewQPushButton2("Dump...", nil)
	dumpButton.ConnectReleased(func() {
		schemaLocation := widgets.QFileDialog_GetSaveFileName(dumpButton, "Save schema...", "", "TXT files (*.txt)", "", 0)
		schemaFile, err := os.OpenFile(schemaLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		err = MainLayer.Schema.Dump(schemaFile)
		if err != nil {
			println("while encoding:", err.Error())
		}

	})
	layerLayout.AddWidget(dumpButton, 0, 0)
}
