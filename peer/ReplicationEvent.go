package peer

import "github.com/gskartwii/rbxfile"

// ReplicationEvent describes an event invocation replication packet.
type ReplicationEvent struct {
	Arguments []rbxfile.Value
}

func (schema *StaticEventSchema) Decode(reader PacketReader, thisBitstream SerializeReader, layers *PacketLayers) (*ReplicationEvent, error) {
	var err error
	var thisVal rbxfile.Value

	event := &ReplicationEvent{}
	event.Arguments = make([]rbxfile.Value, len(schema.Arguments))
	for i, argSchema := range schema.Arguments {
		thisVal, err = thisBitstream.ReadSerializedValue(reader, argSchema.Type, argSchema.EnumID)
		event.Arguments[i] = thisVal
		if err != nil {
			return event, err
		}
	}

	return event, nil
}

func (schema *StaticEventSchema) Serialize(event *ReplicationEvent, writer PacketWriter, stream SerializeWriter) error {
	for i, argSchema := range schema.Arguments {
		//println("Writing argument", argSchema.Type)
		err := stream.WriteSerializedValue(event.Arguments[i], writer, argSchema.Type)
		if err != nil {
			return err
		}
	}
	return nil
}
