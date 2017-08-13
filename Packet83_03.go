package main
import "github.com/google/gopacket"
import "errors"
import "fmt"
import "github.com/therecipe/qt/widgets"

type Packet83_03 struct {
	Object1 Object
	Bool1 bool
	PropertyName string
	Schema *PropertySchemaItem
	Value *ReplicationProperty
}

func (this Packet83_03) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Referent: %s", this.Object1.Show()), 0, 0)
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Property name: %s", this.PropertyName), 0, 0)
	layout.AddWidget(NewQLabelF("Property type: %s", this.Schema.Type), 0, 0)
	layout.AddWidget(NewQLabelF("Property value: %s", this.Value.Show()), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func DecodePacket83_03(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, propertySchema []*PropertySchemaItem) (interface{}, error) {
	var err error
	layer := &Packet83_03{}
	layer.Object1, err = thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

	propertyIDx, err := thisBitstream.Bits(0xB)
	if err != nil {
		return layer, err
	}
	realIDx := (propertyIDx & 0x7 << 8) | propertyIDx >> 3

	if int(realIDx) > int(len(propertySchema)) {
		return layer, errors.New(fmt.Sprintf("prop idx %d is higher than %d", realIDx, len(propertySchema)))
	}

	schema := propertySchema[realIDx]
	layer.PropertyName = schema.Name
	println(DebugInfo2(context, packet, false), "Our prop: ", layer.PropertyName)

	layer.Bool1, err = thisBitstream.ReadBool()
	if err != nil {
		return layer, err
	}

	layer.Value, err = schema.Decode(ROUND_UPDATE, thisBitstream, context, packet)
	layer.Schema = schema
	return layer, err
}
