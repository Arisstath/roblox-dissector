package peer
import "github.com/gskartwii/rbxfile"

type ReplicationEvent struct {
	UnknownInt uint32
	Arguments []rbxfile.Value
}

func (schema *StaticEventSchema) Decode(packet *UDPPacket, context *CommunicationContext) (*ReplicationEvent, error) {
	var err error
    var thisVal rbxfile.Value
	thisBitstream := packet.Stream

	event := &ReplicationEvent{}
	event.Arguments = make([]rbxfile.Value, len(schema.Arguments))
	for i, argSchema := range schema.Arguments {
		thisVal, err = readSerializedValue(false, argSchema.Type, thisBitstream, context)
		event.Arguments[i] = thisVal
		if err != nil {
			return event, err
		}
	}

	return event, nil
}
