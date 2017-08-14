package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "errors"
import "fmt"

type Packet83_07 struct {
	Object1 Object
	EventName string
	Schema *EventSchemaItem
	Event *ReplicationEvent
}

func (this Packet83_07) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Referent: %s", this.Object1.Show()), 0, 0)
	layout.AddWidget(NewQLabelF("Event name: %s", this.EventName), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown int: %d", this.Event.UnknownInt), 0, 0)
	layout.AddWidget(NewQLabelF("Arguments:"), 0, 0)

	argumentList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Type", "Value"})
	rootNode := standardModel.InvisibleRootItem()

	for i, argument := range this.Event.Arguments {
		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF(this.Schema.ArgumentTypes[i]),
			NewQStandardItemF("%s", argument.Show()),
		})
	}

	argumentList.SetModel(standardModel)
	argumentList.SetSelectionMode(0)
	argumentList.SetSortingEnabled(true)
	layout.AddWidget(argumentList, 0, 0)
	widget.SetLayout(layout)
	
	return widget
}

func DecodePacket83_07(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, eventSchema []*EventSchemaItem) (interface{}, error) {
	var err error
	layer := &Packet83_07{}
	layer.Object1, err = thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

	eventIDx, err := thisBitstream.Bits(0x9)
	if err != nil {
		return layer, err
	}
	realIDx := (eventIDx & 1 << 8) | eventIDx >> 1

	if int(realIDx) > int(len(eventSchema)) {
		return layer, errors.New(fmt.Sprintf("event idx %d is higher than %d", realIDx, len(eventSchema)))
	}

	schema := eventSchema[realIDx]
	layer.EventName = schema.Name
	println(DebugInfo2(context, packet, false), "Our event: ", layer.EventName)

	layer.Event, err = schema.Decode(thisBitstream, context, packet)
	layer.Schema = schema
	return layer, err
}
