package peer

import (
	"errors"

	"github.com/robloxapi/rbxfile"
)

// ReplicationEvent describes an event invocation replication packet.
type ReplicationEvent struct {
	Arguments []rbxfile.Value
}

// Decode deserializes a network event invocation packet
func (schema *NetworkEventSchema) Decode(reader PacketReader, thisStream serializeReader, layers *PacketLayers, deferred deferredStrings) (*ReplicationEvent, error) {
	var err error
	var thisVal rbxfile.Value

	event := &ReplicationEvent{}
	event.Arguments = make([]rbxfile.Value, len(schema.Arguments))
	for i, argSchema := range schema.Arguments {
		thisVal, err = thisStream.ReadSerializedValue(reader, argSchema.Type, argSchema.EnumID, deferred)
		if err != nil {
			return event, err
		}

		event.Arguments[i] = thisVal
	}

	return event, nil
}

// Serialize serializes an event invocation packet to its network format
func (schema *NetworkEventSchema) Serialize(event *ReplicationEvent, writer PacketWriter, stream serializeWriter, deferred writeDeferredStrings) error {
	if len(event.Arguments) != len(schema.Arguments) {
		return errors.New("invalid number of event arguments")
	}
	for i, argSchema := range schema.Arguments {
		//println("Writing argument", argSchema.Type)
		if event.Arguments[i] != nil {
			err := stream.WriteSerializedValue(event.Arguments[i], writer, argSchema.Type, deferred)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
