package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "./peer"

func ShowPacket97(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(peer.Packet97Layer)

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)

	labelForEnumSchema := NewQLabelF("Enum schema (%d entries):", len(MainLayer.Schema.Enums))
	layerLayout.AddWidget(labelForEnumSchema, 0, 0)

	enumSchemaList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Size"})

	enumSchemaRootNode := standardModel.InvisibleRootItem()
	for _, item := range MainLayer.Schema.Enums {
		nameItem := NewQStandardItemF(item.Name)
		sizeItem := NewQStandardItemF("%d", item.BitSize)
		enumSchemaRootNode.AppendRow([]*gui.QStandardItem{nameItem, sizeItem})
	}

	enumSchemaList.SetModel(standardModel)
	enumSchemaList.SetSelectionMode(0)
	enumSchemaList.SetSortingEnabled(true)
	layerLayout.AddWidget(enumSchemaList, 0, 0)


	labelForInstanceSchema := NewQLabelF("Instance schema (%d entries):", len(MainLayer.Schema.Instances))
	layerLayout.AddWidget(labelForInstanceSchema, 0, 0)
	instanceSchemaList := widgets.NewQTreeView(nil)
	instanceModel := NewProperSortModel(nil)
	instanceModel.SetHorizontalHeaderLabels([]string{"Name", "Type"})
	instanceSchemaRootNode := instanceModel.InvisibleRootItem()

	for _, item := range MainLayer.Schema.Instances {
		nameItem := NewQStandardItemF(item.Name)
		instanceRow := []*gui.QStandardItem{nameItem, nil}

		propertySchemaItem := NewQStandardItemF("Property schema (%d entries)", len(item.Properties))

		for _, property := range item.Properties {
			propertyNameItem := NewQStandardItemF(property.Name)
			propertyTypeItem := NewQStandardItemF(property.TypeString)

			propertyRow := []*gui.QStandardItem{propertyNameItem, propertyTypeItem}
			propertySchemaItem.AppendRow(propertyRow)
		}
		nameItem.AppendRow([]*gui.QStandardItem{propertySchemaItem})

		eventSchemaItem := NewQStandardItemF("Event schema (%d entries)", len(item.Events))
		nameItem.AppendRow([]*gui.QStandardItem{eventSchemaItem})

		for _, event := range item.Events {
			eventNameItem := NewQStandardItemF("%s (%d arguments)", event.Name, len(event.Arguments))

			eventRow := []*gui.QStandardItem{eventNameItem, nil}

			for _, thisArgument := range event.Arguments {
				eventArgumentNameItem := NewQStandardItemF("Event argument")
				eventArgumentTypeItem := NewQStandardItemF(thisArgument.TypeString)

				eventSubIntRow := []*gui.QStandardItem{eventArgumentNameItem, eventArgumentTypeItem}
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
