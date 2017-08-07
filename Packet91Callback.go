package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"

func ShowPacket91(packetType byte, packet gopacket.Packet, context *CommunicationContext, layers *PacketLayers) {
	MainLayer := layers.Main.(Packet91Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	labelForEnumSchema := NewQLabelF("Enum schema (%d entries):", len(MainLayer.EnumSchema))
	layerLayout.AddWidget(labelForEnumSchema, 0, 0)

	enumSchemaList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"ID", "Name", "Size"})
	
	enumSchemaRootNode := standardModel.InvisibleRootItem()
	for id, item := range MainLayer.EnumSchema {
		idItem := NewQStandardItemF("%d", id)
		nameItem := NewQStandardItemF(item.Name)
		sizeItem := NewQStandardItemF("%d", item.BitSize)
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
	instanceModel.SetHorizontalHeaderLabels([]string{"Name", "Common ID", "Type", "Type from dictionary", "Is creatable?", "Property replicates?", "Is enum?", "Size in bits"})
	instanceSchemaRootNode := instanceModel.InvisibleRootItem()

	for _, item := range MainLayer.InstanceSchema {
		nameItem := NewQStandardItemF(item.Name)
		commonIDItem := NewQStandardItemF("%d", item.CommonID)
		creatableItem := NewQStandardItemF("%v", item.IsCreatable)

		instanceRow := []*gui.QStandardItem{nameItem, commonIDItem, nil, nil, creatableItem}

		propertySchemaItem := NewQStandardItemF("Property schema (%d entries)", len(item.PropertySchema))
		nameItem.AppendRow([]*gui.QStandardItem{propertySchemaItem})

		for _, property := range item.PropertySchema {
			propertyNameItem := NewQStandardItemF(property.Name)
			propertyCommonIDItem := NewQStandardItemF("%d", property.CommonID)
			propertyTypeItem := NewQStandardItemF(property.Type)
			propertyDictionaryTypeItem := NewQStandardItemF(property.DictionaryType)
			propertyReplicatesItem := NewQStandardItemF("%v", property.Replicates)
			propertyIsEnumItem := NewQStandardItemF("%v", property.IsEnum)
			propertyBitSizeItem := NewQStandardItemF("%d", property.BitSize)


			propertyRow := []*gui.QStandardItem{propertyNameItem, propertyCommonIDItem, propertyTypeItem, propertyDictionaryTypeItem, nil, propertyReplicatesItem, propertyIsEnumItem, propertyBitSizeItem}
			propertySchemaItem.AppendRow(propertyRow)
		}
		
		eventSchemaItem := NewQStandardItemF("Event schema (%d entries)", len(item.EventSchema))
		nameItem.AppendRow([]*gui.QStandardItem{eventSchemaItem})

		for _, event := range item.EventSchema {
			eventNameItem := NewQStandardItemF("%s (%d arguments)", event.Name, len(event.ArgumentTypes))
			eventCommonIDItem := NewQStandardItemF("%d", event.CommonID)

			eventRow := []*gui.QStandardItem{eventNameItem, eventCommonIDItem}

			for _, thisArgument := range event.ArgumentTypes {
				eventArgumentNameItem := NewQStandardItemF("Event argument")
				eventArgumentTypeItem := NewQStandardItemF(thisArgument)

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
