package peer

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

// Packet83_03 represents ID_CHANGE_PROPERTY
type Packet83_03 struct {
	// Instance that had the property change
	Instance   *datamodel.Instance
	HasVersion bool
	Version    int32
	Schema     *NetworkPropertySchema
	// New value
	Value rbxfile.Value
}

func (thisStream *extendedReader) DecodePacket83_03(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	layer := &Packet83_03{}

	reference, err := thisStream.ReadObject(reader)
	if err != nil {
		return layer, err
	}
	if reference.IsNull {
		return layer, errors.New("self is null in repl property")
	}
	layer.Instance, err = reader.Context().InstancesByReference.TryGetInstance(reference)
	if err != nil {
		return layer, err
	}

	propertyIDx, err := thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}

	layer.HasVersion, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	// If this packet was written by the client, read version
	if layer.HasVersion && reader.IsClient() {
		layer.Version, err = thisStream.readSintUTF8()
		if err != nil {
			return layer, err
		}
	}

	context := reader.Context()
	if int(propertyIDx) == int(len(context.NetworkSchema.Properties)) { // explicit Parent property system
		var reference datamodel.Reference
		reference, err = thisStream.ReadObject(reader)
		if err != nil {
			return layer, err
		}
		// CreateInstance: allow forward references in ID_REPLIC_PROP
		result := datamodel.ValueReference{Reference: reference}
		refInstance, err := context.InstancesByReference.CreateInstance(reference)
		if err != nil {
			return layer, err
		}
		result.Instance = refInstance
		layer.Value = result
		layer.Schema = nil

		return layer, err
	}

	if int(propertyIDx) > int(len(context.NetworkSchema.Properties)) {
		return layer, fmt.Errorf("prop idx %d is higher than %d", propertyIDx, len(context.NetworkSchema.Properties))
	}
	schema := context.NetworkSchema.Properties[propertyIDx]
	layer.Schema = schema

	deferred := newDeferredStrings(reader)
	layer.Value, err = schema.Decode(reader, thisStream, layers, deferred)
	if err != nil {
		return layer, err
	}

	err = thisStream.resolveDeferredStrings(deferred)
	if err != nil {
		return layer, err
	}

	return layer, nil
}

// Serialize implements Packet83Subpacke.Serialize()
func (layer *Packet83_03) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("self is nil in serialize repl prop")
	}

	err := stream.WriteObject(layer.Instance, writer)
	if err != nil {
		return err
	}

	context := writer.Context()
	if layer.Schema == nil { // assume Parent property
		err = stream.writeUint16BE(uint16(len(context.NetworkSchema.Properties)))
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(layer.HasVersion)
		if err != nil {
			return err
		}
		// If this packet is to the server, write version
		if layer.HasVersion && !writer.ToClient() {
			err = stream.writeSintUTF8(layer.Version)
			if err != nil {
				return err
			}
		}

		return stream.WriteObject(layer.Value.(datamodel.ValueReference).Instance, writer)
	}

	err = stream.writeUint16BE(layer.Schema.NetworkID)
	if err != nil {
		return err
	}

	err = stream.writeBoolByte(layer.HasVersion)
	if err != nil {
		return err
	}
	if layer.HasVersion && !writer.ToClient() {
		err = stream.writeSintUTF8(layer.Version)
		if err != nil {
			return err
		}
	}

	// TODO: A different system for this?
	deferred := newWriteDeferredStrings(writer)
	err = layer.Schema.Serialize(layer.Value, writer, stream, deferred)
	if err != nil {
		return err
	}
	return stream.resolveDeferredStrings(deferred)
}

// Type implements Packet83Subpacket.Type()
func (Packet83_03) Type() uint8 {
	return 3
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_03) TypeString() string {
	return "ID_REPLIC_PROP"
}

func (layer *Packet83_03) String() string {
	var propName string
	if layer.Schema == nil {
		propName = "Parent"
	} else {
		propName = layer.Schema.Name
	}
	return fmt.Sprintf("ID_REPLIC_PROP: %s: %s[%s]", layer.Instance.Ref.String(), layer.Instance.Name(), propName)
}
