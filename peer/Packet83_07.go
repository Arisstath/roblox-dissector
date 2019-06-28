package peer

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet83_07 represents ID_EVENT
type Packet83_07 struct {
	// Instance that the event was invoked on
	Instance *datamodel.Instance
	Schema   *NetworkEventSchema
	// Description about the invocation
	Event *ReplicationEvent
}

func (thisStream *extendedReader) DecodePacket83_07(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	layer := &Packet83_07{}

	reference, err := thisStream.readObject(reader.Caches())
	if err != nil {
		return layer, err
	}
	if reference.IsNull {
		return layer, errors.New("self is nil in decode repl event")
	}
	layer.Instance, err = reader.Context().InstancesByReference.TryGetInstance(reference)
	if err != nil {
		return layer, err
	}

	eventIDx, err := thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}

	context := reader.Context()
	if int(eventIDx) > int(len(context.NetworkSchema.Events)) {
		return layer, fmt.Errorf("event idx %d is higher than %d", eventIDx, len(context.NetworkSchema.Events))
	}

	deferred := newDeferredStrings(reader)
	schema := context.NetworkSchema.Events[eventIDx]
	layer.Schema = schema
	layers.Root.Logger.Println("Decoding event", schema.Name)
	layer.Event, err = schema.Decode(reader, thisStream, layers, deferred)
	if err != nil {
		return layer, err
	}

	err = thisStream.resolveDeferredStrings(deferred)
	if err != nil {
		return layer, err
	}

	return layer, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_07) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err := stream.writeObject(layer.Instance, writer.Caches())
	if err != nil {
		return err
	}

	err = stream.writeUint16BE(uint16(layer.Schema.NetworkID))
	if err != nil {
		return err
	}

	deferred := newWriteDeferredStrings(writer)
	err = layer.Schema.Serialize(layer.Event, writer, stream, deferred)
	if err != nil {
		return err
	}

	return stream.resolveDeferredStrings(deferred)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_07) Type() uint8 {
	return 7
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_07) TypeString() string {
	return "ID_REPLIC_EVENT"
}

func (layer *Packet83_07) String() string {
	return fmt.Sprintf("ID_REPLIC_EVENT: %s::%s", layer.Instance.GetFullName(), layer.Schema.Name)
}
