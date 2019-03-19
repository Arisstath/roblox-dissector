package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

// ID_EVENT
type Packet83_07 struct {
	// Instance that the event was invoked on
	Instance *datamodel.Instance
	Schema   *StaticEventSchema
	// Description about the invocation
	Event *ReplicationEvent
}

func (thisBitstream *extendedReader) DecodePacket83_07(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	layer := &Packet83_07{}

	referent, err := thisBitstream.readObject(reader.Caches())
	if err != nil {
		return layer, err
	}
	if referent.IsNull {
		return layer, errors.New("self is nil in decode repl event")
	}
	layer.Instance, err = reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return layer, err
	}

	eventIDx, err := thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}

	context := reader.Context()
	if int(eventIDx) > int(len(context.StaticSchema.Events)) {
		return layer, fmt.Errorf("event idx %d is higher than %d", eventIDx, len(context.StaticSchema.Events))
	}

	schema := context.StaticSchema.Events[eventIDx]
	layer.Schema = schema
	layers.Root.Logger.Println("Decoding event", schema.Name)
	layer.Event, err = schema.Decode(reader, thisBitstream, layers)
	if err != nil {
		return layer, err
	}

	return layer, err
}

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

	return layer.Schema.Serialize(layer.Event, writer, stream)
}

func (Packet83_07) Type() uint8 {
	return 7
}
func (Packet83_07) TypeString() string {
	return "ID_REPLIC_EVENT"
}

func (layer *Packet83_07) String() string {
	return fmt.Sprintf("ID_REPLIC_EVENT: %s::%s", layer.Instance.GetFullName(), layer.Schema.Name)
}
