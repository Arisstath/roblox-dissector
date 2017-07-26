package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "strconv"
import "fmt"

func ShowPacket91(data []byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet91Layer)

	layerLayout := NewBasicPacketViewer(data, packet, context, layers)

	labelForEnumSchema := NewQLabelF("Enum schema (%d entries):", len(MainLayer.EnumSchema))
	layerLayout.AddWidget(labelForEnumSchema, 0, 0)

	enumSchemaList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"ID", "Name", "Size"})
	
	enumSchemaRootNode := standardModel.InvisibleRootItem()
	for id, item := range MainLayer.EnumSchema {
		idItem := gui.NewQStandardItem2(strconv.Itoa(int(id)))
		nameItem := gui.NewQStandardItem2(item.Name)
		sizeItem := gui.NewQStandardItem2(strconv.Itoa(int(item.BitSize)))
		idItem.SetEditable(false)
		nameItem.SetEditable(false)
		sizeItem.SetEditable(false)
		enumSchemaRootNode.AppendRow([]*gui.QStandardItem{idItem, nameItem, sizeItem}) 
	}

	enumSchemaList.SetModel(standardModel)
	enumSchemaList.SetSelectionMode(0)
	enumSchemaList.SetSortingEnabled(true)
	layerLayout.AddWidget(enumSchemaList, 0, 0)


	labelForInstanceSchema := NewQLabelF("Instance schema (%d entries):", len(MainLayer.InstanceSchema))
	layerLayout.AddWidget(labelForInstanceSchema, 0, 0)
	instanceSchemaList := widgets.NewQTreeView(nil)
	instanceModel := NewProperSortModel(nil)
	instanceModel.SetHorizontalHeaderLabels([]string{"Name", "Common ID", "Type", "Type from dictionary", "Is creatable?", "Unknown bool 1", "Is enum?", "Size in bits"})
	instanceSchemaRootNode := instanceModel.InvisibleRootItem()

	for _, item := range MainLayer.InstanceSchema {
		nameItem := gui.NewQStandardItem2(item.Name)
		commonIDItem := gui.NewQStandardItem2(strconv.Itoa(int(item.CommonID)))
		creatableItem := gui.NewQStandardItem2(strconv.FormatBool(item.IsCreatable))
		nameItem.SetEditable(false)
		commonIDItem.SetEditable(false)
		creatableItem.SetEditable(false)

		instanceRow := []*gui.QStandardItem{nameItem, commonIDItem, nil, nil, creatableItem}

		propertySchemaString := fmt.Sprintf("Property schema (%d entries)", len(item.PropertySchema))
		propertySchemaItem := gui.NewQStandardItem2(propertySchemaString)
		propertySchemaItem.SetEditable(false)
		nameItem.AppendRow([]*gui.QStandardItem{propertySchemaItem})

		for _, property := range item.PropertySchema {
			propertyNameItem := gui.NewQStandardItem2(property.Name)
			propertyCommonIDItem := gui.NewQStandardItem2(strconv.Itoa(int(property.CommonID)))
			propertyTypeItem := gui.NewQStandardItem2(property.Type)
			propertyDictionaryTypeItem := gui.NewQStandardItem2(property.DictionaryType)
			propertyBool1Item := gui.NewQStandardItem2(strconv.FormatBool(property.Bool1))
			propertyIsEnumItem := gui.NewQStandardItem2(strconv.FormatBool(property.IsEnum))
			propertyBitSizeItem := gui.NewQStandardItem2(strconv.Itoa(int(property.BitSize)))

			propertyNameItem.SetEditable(false)
			propertyCommonIDItem.SetEditable(false)
			propertyTypeItem.SetEditable(false)
			propertyDictionaryTypeItem.SetEditable(false)
			propertyBool1Item.SetEditable(false)
			propertyIsEnumItem.SetEditable(false)
			propertyBitSizeItem.SetEditable(false)

			propertyRow := []*gui.QStandardItem{propertyNameItem, propertyCommonIDItem, propertyTypeItem, propertyDictionaryTypeItem, nil, propertyBool1Item, propertyIsEnumItem, propertyBitSizeItem}
			propertySchemaItem.AppendRow(propertyRow)
		}
		
		eventSchemaString := fmt.Sprintf("Event schema (%d entries)", len(item.EventSchema))
		eventSchemaItem := gui.NewQStandardItem2(eventSchemaString)
		eventSchemaItem.SetEditable(false)
		nameItem.AppendRow([]*gui.QStandardItem{eventSchemaItem})

		for _, event := range item.EventSchema {
			eventNameString := fmt.Sprintf("%s (%d arguments)", event.Name, len(event.ArgumentTypes))
			eventNameItem := gui.NewQStandardItem2(eventNameString)
			eventCommonIDItem := gui.NewQStandardItem2(strconv.Itoa(int(event.CommonID)))
			eventNameItem.SetEditable(false)
			eventCommonIDItem.SetEditable(false)

			eventRow := []*gui.QStandardItem{eventNameItem, eventCommonIDItem}

			for _, thisArgument := range event.ArgumentTypes {
				eventArgumentNameItem := gui.NewQStandardItem2("Event argument")
				eventArgumentTypeItem := gui.NewQStandardItem2(thisArgument)
				eventArgumentNameItem.SetEditable(false)
				eventArgumentTypeItem.SetEditable(false)

				eventSubIntRow := []*gui.QStandardItem{eventArgumentNameItem, nil, eventArgumentTypeItem}
				eventNameItem.AppendRow(eventSubIntRow)
			}

			eventSchemaItem.AppendRow(eventRow)
		}

		instanceSchemaRootNode.AppendRow(instanceRow)
	}

	instanceSchemaList.SetModel(instanceModel)
	instanceSchemaList.SetSelectionMode(0)
	instanceSchemaList.SetSortingEnabled(true)
	layerLayout.AddWidget(instanceSchemaList, 0, 0)
}
