package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/rbxfile"
)

// ID_EVENT
type ReplicateEvent struct {
	// Instance that the event was invoked on
	Instance *rbxfile.Instance
	// Name of the event
	EventName string
	// Description about the invocation
	Event *ReplicationEvent
}

func (thisBitstream *extendedReader) DecodeReplicateEvent(reader PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	layer := &ReplicateEvent{}

	referent, err := thisBitstream.readObject(reader.Caches())
	if err != nil {
		return layer, err
	}
	if referent.IsNull() {
		return layer, errors.New("self is nil in decode repl event")
	}
	instance, err := reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return layer, err
	}
	layer.Instance = instance

	eventIDx, err := thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}

	context := reader.Context()
	if int(eventIDx) > int(len(context.StaticSchema.Events)) {
		return layer, fmt.Errorf("event idx %d is higher than %d", eventIDx, len(context.StaticSchema.Events))
	}

	schema := context.StaticSchema.Events[eventIDx]
	layer.EventName = schema.Name
	layers.Root.Logger.Println("Decoding event", layer.EventName)
	layer.Event, err = schema.Decode(reader, thisBitstream, layers)
	if err != nil {
		return layer, err
	}
	return layer, err
}

func (layer *ReplicateEvent) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err := stream.writeObject(layer.Instance, writer.Caches())
	if err != nil {
		return err
	}

	context := writer.Context()
	eventSchemaID, ok := context.StaticSchema.EventsByName[layer.Instance.ClassName+"."+layer.EventName]
	if !ok {
		return errors.New("Invalid event: " + layer.Instance.ClassName + "." + layer.EventName)
	}
	err = stream.writeUint16BE(uint16(eventSchemaID))
	if err != nil {
		return err
	}

	schema := context.StaticSchema.Events[uint16(eventSchemaID)]
	//println("Writing event", schema.Name, schema.InstanceSchema.Name)

	return schema.Serialize(layer.Event, writer, stream)
}

func (ReplicateEvent) Type() uint8 {
	return 7
}
func (ReplicateEvent) TypeString() string {
	return "ID_REPLIC_EVENT"
}
