package peer
import "github.com/gskartwii/rbxfile"

// ReplicationEvent describes an event invocation replication packet.
type ReplicationEvent struct {
	Arguments []rbxfile.Value
}

func (schema *StaticEventSchema) Decode(packet *UDPPacket, context *CommunicationContext) (*ReplicationEvent, error) {
	var err error
    var thisVal rbxfile.Value
	thisBitstream := packet.stream

	event := &ReplicationEvent{}
	event.Arguments = make([]rbxfile.Value, len(schema.Arguments))
	for i, argSchema := range schema.Arguments {
		thisVal, err = readSerializedValue(context.IsClient(packet.Source), false, argSchema.EnumID, argSchema.Type, thisBitstream, context)
		event.Arguments[i] = thisVal
		if err != nil {
			return event, err
		}
	}

	return event, nil
}

func (schema *StaticEventSchema) serialize(isClient bool, event *ReplicationEvent, context *CommunicationContext, stream *extendedWriter) error {
	for i, argSchema := range schema.Arguments {
		println("Writing argument", argSchema.Type)
		err := stream.writeSerializedValue(isClient, event.Arguments[i], false, argSchema.Type, context)
		if err != nil {
			return err
		}
	}
	return nil
}
