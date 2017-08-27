package peer
import "errors"
import "fmt"

type Packet83_07 struct {
	Object1 Object
	EventName string
	Event *ReplicationEvent
}

func DecodePacket83_07(packet *UDPPacket, context *CommunicationContext, eventSchema []*EventSchemaItem) (interface{}, error) {
	var err error
	layer := &Packet83_07{}
	thisBitstream := packet.Stream
	layer.Object1, err = thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

	if !context.UseStaticSchema {
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

		layer.Event, err = schema.Decode(packet, context)
		return layer, err
	} else {
		eventIDx, err := thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}

		if int(eventIDx) > int(len(context.StaticSchema.Events)) {
			return layer, errors.New(fmt.Sprintf("event idx %d is higher than %d", eventIDx, len(context.StaticSchema.Events)))
		}

		schema := context.StaticSchema.Events[eventIDx]
		layer.EventName = schema.Name
		layer.Event, err = schema.Decode(packet, context)
		return layer, err
	}
}
