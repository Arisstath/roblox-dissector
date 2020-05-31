package main

import (
	"fmt"
	"os"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/xml"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func showChildren(rootNode *gui.QStandardItem, children []*datamodel.Instance) {
	for _, instance := range children {
		row := showReplicationInstance(instance, instance.Parent())
		instance.PropertiesMutex.RLock()
		if len(instance.Properties) > 0 {
			row[0].AppendRow([]*gui.QStandardItem{showProperties(instance.Properties, 0)})
		}
		instance.PropertiesMutex.RUnlock()
		if len(instance.Children) > 0 {
			childrenRootItem := NewQStandardItemF("%d children", len(instance.Children))
			showChildren(childrenRootItem, instance.Children)
			row[0].AppendRow([]*gui.QStandardItem{childrenRootItem, nil, nil, nil, nil, nil})
		}
		rootNode.AppendRow(row)
	}
}

type InstanceProperty struct {
	Instance *rbxfile.Instance
	Name     string
}

func findScripts(instances []*rbxfile.Instance, propertyList []InstanceProperty) []InstanceProperty {
	for _, instance := range instances {
		for name, property := range instance.Properties {
			thisType := property.Type()
			if thisType == datamodel.TypeSignedProtectedString {
				propertyList = append(propertyList, InstanceProperty{
					Instance: instance,
					Name:     name,
				})
			}
		}
		propertyList = findScripts(instance.Children, propertyList)
	}
	return propertyList
}

func dumpScripts(location string, instances []*rbxfile.Instance, parent widgets.QWidget_ITF) error {
	instList := findScripts(instances, nil)

	progressDialog := widgets.NewQProgressDialog2("Dumping scripts...", "Cancel", 0, len(instList), parent, 0)
	progressDialog.SetWindowTitle("Dumping scripts")
	progressDialog.SetWindowModality(core.Qt__WindowModal)

	for i, instProp := range instList {
		file, err := os.Create(fmt.Sprintf("%s/%s.%d.rbxc", location, instProp.Instance.GetFullName(), i))
		if err != nil {
			progressDialog.Cancel()
			return err
		}
		_, err = file.Write([]byte(instProp.Instance.Properties[instProp.Name].(datamodel.ValueSignedProtectedString).Value))
		if err != nil {
			progressDialog.Cancel()
			return err
		}
		err = file.Close()
		if err != nil {
			progressDialog.Cancel()
			return err
		}

		// HACK: We clear the script source here to prevent issues with the XML parser
		// This may be unexpected in the case of Studio, where the source code is not protected
		delete(instProp.Instance.Properties, instProp.Name)

		if progressDialog.WasCanceled() {
			return nil
		}
		progressDialog.SetValue(i)
	}
	progressDialog.SetValue(len(instList))
	return nil
}

func DumpDataModel(context *peer.CommunicationContext, parent widgets.QWidget_ITF) {
	writableClone := context.DataModel.ToRbxfile()

	location := widgets.QFileDialog_GetExistingDirectory(parent, "Select directory to dump DataModel to", "", 0)
	writer, err := os.OpenFile(location+"/datamodel.rbxlx", os.O_RDWR|os.O_CREATE, 0666)
	defer writer.Close()
	if err != nil {
		widgets.QMessageBox_Critical(parent, "Error while creating RBXLX", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		return
	}

	if LumikideEnabled {
		err = LumikideProcessContext(parent, context, writableClone)
		if err != nil {
			widgets.QMessageBox_Critical(parent, "Error while processing with Lumikide", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
			return
		}
	} else {
		println("Lumikide disabled at compile time, skipping processing.")
	}

	scriptData, err := os.OpenFile(location+"/scriptKeys", os.O_RDWR|os.O_CREATE, 0666)
	defer scriptData.Close()
	if err != nil {
		widgets.QMessageBox_Critical(parent, "Error while dumping script keys", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		return
	}

	_, err = fmt.Fprintf(scriptData, "Script key: %d\nCore script key: %d", context.ScriptKey, context.CoreScriptKey)
	if err != nil {
		widgets.QMessageBox_Critical(parent, "Error while dumping script keys", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		return
	}

	err = dumpScripts(location, writableClone.Instances, parent)
	if err != nil {
		widgets.QMessageBox_Critical(parent, "Error while dumping scripts", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		return
	}

	err = xml.Serialize(writer, nil, writableClone)
	if err != nil {
		widgets.QMessageBox_Critical(parent, "Error while serializing place", err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		return
	}
}

func NewDataModelBrowser(context *peer.CommunicationContext, dataModel *datamodel.DataModel, parent widgets.QWidget_ITF) {
	subWindow := widgets.NewQWidget(parent, core.Qt__Window)
	subWindowLayout := widgets.NewQVBoxLayout2(subWindow)
	subWindowLayout.SetAlign(core.Qt__AlignTop)

	subWindow.SetWindowTitle("Data Model")

	takeSnapshotButton := widgets.NewQPushButton2("Dump DataModel (RBXLX) and scripts (RBXC)...", nil)
	takeSnapshotButton.ConnectReleased(func() {
		DumpDataModel(context, subWindow)
	})

	subWindowLayout.AddWidget(takeSnapshotButton, 0, 0)

	instanceList := widgets.NewQTreeView(subWindow)
	standardModel := NewProperSortModel(subWindow)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Reference", "Parent", "Path"})

	rootNode := standardModel.InvisibleRootItem()
	showChildren(rootNode, dataModel.Instances)
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	subWindowLayout.AddWidget(instanceList, 0, 0)

	subWindow.Show()
}
