package main
import "github.com/google/gopacket"
import "errors"
import "fmt"
import "github.com/therecipe/qt/widgets"
import "github.com/gskartwii/rbxfile"

type Packet83_03 struct {
	Instance *rbxfile.Instance
	Bool1 bool
	PropertyName string
	Value rbxfile.Value
}

func (this Packet83_03) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Object: %s", this.Instance.Reference), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Property name: %s", this.PropertyName), 0, 0)
	layout.AddWidget(NewQLabelF("Property type: %s", this.Value.Type().String()), 0, 0)
	layout.AddWidget(NewQLabelF("Property value: %s", this.Value.String()), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func DecodePacket83_03(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, propertySchema []*PropertySchemaItem) (interface{}, error) {
	var err error
	layer := &Packet83_03{}
    referent, err := thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}
    instance, ok := context.InstancesByReferent[referent]
    layer.Instance = instance
    if !ok {
        return layer, errors.New("invalid rebind: " + string(referent))
    }

    propertyIDx, err := thisBitstream.ReadUint16BE()
    if err != nil {
        return layer, err
    }

    if int(propertyIDx) >= int(len(context.StaticSchema.Properties)) {
        return layer, errors.New(fmt.Sprintf("prop idx %d is higher than %d", propertyIDx, len(context.StaticSchema.Properties)))
    }
    schema := context.StaticSchema.Properties[propertyIDx]
    layer.PropertyName = schema.Name

    layer.Bool1, err = thisBitstream.ReadBool()
    if err != nil {
        return layer, err
    }

    layer.Value, err = schema.Decode(ROUND_UPDATE, thisBitstream, context, packet)
    instance.Properties[layer.PropertyName] = layer.Value
    return layer, err
}
