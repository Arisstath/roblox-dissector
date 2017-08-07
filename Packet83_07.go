package main
import "github.com/google/gopacket"
import "errors"
import "fmt"

type Packet83_07 struct {
	Object1 Object
	EventName string
	Schema *EventSchemaItem
	Event *ReplicationEvent
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
