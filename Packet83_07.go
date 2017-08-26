package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "errors"
import "fmt"
import "github.com/gskartwii/rbxfile"

type Packet83_07 struct {
	Instance *rbxfile.Instance
	EventName string
	Event *ReplicationEvent
}

func (this Packet83_07) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Object: %s", this.Instance.Reference), 0, 0)
	layout.AddWidget(NewQLabelF("Event name: %s", this.EventName), 0, 0)
	layout.AddWidget(NewQLabelF("Arguments:"), 0, 0)

	argumentList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Type", "Value"})
	rootNode := standardModel.InvisibleRootItem()

	for _, argument := range this.Event.Arguments {
		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF(argument.Type().String()),
			NewQStandardItemF("%s", argument.String()),
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
    referent, err := thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

    eventIDx, err := thisBitstream.ReadUint16BE()
    if err != nil {
        return layer, err
    }

    if int(eventIDx) > int(len(context.StaticSchema.Events)) {
        return layer, errors.New(fmt.Sprintf("event idx %d is higher than %d", eventIDx, len(context.StaticSchema.Events)))
    }

    schema := context.StaticSchema.Events[eventIDx]
    layer.EventName = schema.Name
    layer.Event, err = schema.Decode(thisBitstream, context, packet)
    context.InstancesByReferent.OnAddInstance(referent, func(instance *rbxfile.Instance) {
        layer.Instance = instance
    })
    return layer, err
}
