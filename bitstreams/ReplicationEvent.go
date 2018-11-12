package bitstreams

import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/schema"
import "github.com/gskartwii/roblox-dissector/util"

// ReplicationEvent describes an event invocation replication packet.
type ReplicationEvent struct {
	Arguments []rbxfile.Value
}

func DecodeReplicationEvent(reader util.PacketReader, thisBitstream SerializeReader, schema *schema.StaticEventSchema) (*ReplicationEvent, error) {
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

func SerializeReplicationEvent(event *ReplicationEvent, writer util.PacketWriter, stream SerializeWriter, schema *schema.StaticEventSchema) error {
	for i, argSchema := range schema.Arguments {
		//println("Writing argument", argSchema.Type)
		err := stream.WriteSerializedValue(event.Arguments[i], writer, argSchema.Type)
		if err != nil {
			return err
		}
	}
	return nil
}
