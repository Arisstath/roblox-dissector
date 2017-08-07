package main
import "github.com/google/gopacket"
import "errors"
import "fmt"

type Packet83_03 struct {
	Object1 Object
	Bool1 bool
	PropertyName string
	Schema *PropertySchemaItem
	Value PropertyValue
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
