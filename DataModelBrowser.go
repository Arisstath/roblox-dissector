package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/rbxfile/bin"
import "github.com/Gskartwii/roblox-dissector/peer"
import "os"
import "fmt"

func showChildren(rootNode *gui.QStandardItem, children []*rbxfile.Instance) {
	for _, instance := range children {
		row := showReplicationInstance(instance)
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
		instance.PropertiesMutex.Lock()
		for name, property := range instance.Properties {
			if property == nil {
				delete(instance.Properties, name)
				continue
			}
			thisType := property.Type()
			if thisType == rbxfile.TypeProtectedString {
				instance.PropertiesMutex.Unlock()
				println("dumping protectedstring", instance.ClassName, name, thisType.String())
				instance.PropertiesMutex.Lock()
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
		instance.PropertiesMutex.Unlock()
		i = dumpScripts(instance.Children, i)
	}
	return i
}

func stripInvalidTypes(instances []*rbxfile.Instance, defaultValues DefaultValues, i int) int {
	for _, instance := range instances {
		instance.PropertiesMutex.Lock()
		color, ok := instance.Properties["Color3uint8"]
		if ok {
			if _, ok = color.(rbxfile.ValueDefault); !ok {
				col := color.(rbxfile.ValueColor3uint8)
				instance.Properties["Color"] = rbxfile.ValueColor3{
					R: float32(col.R) / 255,
					G: float32(col.R) / 255,
					B: float32(col.B) / 255,
				}
			}
		}

		for name, property := range instance.Properties {
			if property == nil {
				delete(instance.Properties, name)
				continue
			}
			thisType := property.Type()
			if thisType == rbxfile.TypeDefault {
				class, ok := defaultValues[instance.ClassName]
				if !ok {
					delete(instance.Properties, name)
					continue
				}
				value, ok := class[name]
				if !ok {
					delete(instance.Properties, name)
					continue
				}
				instance.Properties[name] = value
			} else if thisType >= rbxfile.TypeNumberSequenceKeypoint ||
				thisType == rbxfile.TypeVector2int16 {
				delete(instance.Properties, name)
				continue
			} else if thisType == rbxfile.TypeProtectedString {
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
		instance.PropertiesMutex.Unlock()
		i = stripInvalidTypes(instance.Children, defaultValues, i)
	}
	return i
}

func NewDataModelBrowser(context *peer.CommunicationContext, dataModel *rbxfile.Root, defaultValues DefaultValues) {
	subWindow := widgets.NewQWidget(window, core.Qt__Window)
	subWindowLayout := widgets.NewQVBoxLayout2(subWindow)

	subWindow.SetWindowTitle("Data Model")

	children := dataModel.Copy()

	takeSnapshotButton := widgets.NewQPushButton2("Save as RBXL...", nil)
	takeSnapshotButton.ConnectPressed(func() {
		location := widgets.QFileDialog_GetSaveFileName(subWindow, "Save as RBXL...", "", "Roblox place files (*.rbxl)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		defer writer.Close()
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		writableClone := children.Copy()
		stripInvalidTypes(writableClone.Instances, defaultValues, 0)

		err = bin.SerializePlace(writer, nil, writableClone)
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

	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	showChildren(rootNode, children.Instances)
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	subWindowLayout.AddWidget(instanceList, 0, 0)

	subWindow.Show()
}
