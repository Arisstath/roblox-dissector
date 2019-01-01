package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

// Packet83_03 describes an ID_CHANGE_PROPERTY data subpacket.
type Packet83_03 struct {
	// Instance that had the property change
	Instance    *datamodel.Instance
	Bool1       bool
	Int1        int32
	DoWriteInt1 bool
	// Name of the property
	PropertyName string
	// New value
	Value rbxfile.Value
}

func (thisBitstream *extendedReader) DecodePacket83_03(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	layer := &Packet83_03{}

	referent, err := thisBitstream.readObject(reader.Caches())
	if err != nil {
		return layer, err
	}
	if referent.IsNull {
		return layer, errors.New("self is null in repl property")
	}
	instance, err := reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return layer, err
	}
	layer.Instance = instance

	propertyIDx, err := thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}

	layer.Bool1, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	// If this packet was written by the client, read version
	if layer.Bool1 && reader.IsClient() {
		layer.Int1, err = thisBitstream.readSintUTF8()
		if err != nil {
			return layer, err
		}
	}

	context := reader.Context()
	if int(propertyIDx) == int(len(context.StaticSchema.Properties)) { // explicit Parent property system
		var referent datamodel.Reference
		referent, err = thisBitstream.readObject(reader.Caches())
		parent, err := context.InstancesByReferent.TryGetInstance(referent)
		if err != nil {
			return layer, errors.New("parent doesn't exist in repl property")
		}
		layers.Root.Logger.Printf("Parent: %s -> %s\n", referent, parent.GetFullName())
		result := datamodel.ValueReference{Instance: parent, Reference: referent}
		layer.Value = result
		layer.PropertyName = "Parent"

		if referent.IsNull { // NULL is a valid referent; think about :Remove()!
			instance.SetParent(nil)
			return layer, nil
		}
		err = parent.AddChild(instance)
		return layer, err
	}

	if int(propertyIDx) > int(len(context.StaticSchema.Properties)) {
		return layer, fmt.Errorf("prop idx %d is higher than %d", propertyIDx, len(context.StaticSchema.Properties))
	}
	schema := context.StaticSchema.Properties[propertyIDx]
	layer.PropertyName = schema.Name

	layer.Value, err = schema.Decode(reader, thisBitstream, layers)
	instance.Set(layer.PropertyName, layer.Value)

	return layer, err
}

func (layer *Packet83_03) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if layer.Instance == nil {
		return errors.New("self is nil in serialize repl prop")
	}

	err := stream.writeObject(layer.Instance, writer.Caches())
	if err != nil {
		return err
	}

	context := writer.Context()
	if layer.PropertyName == "Parent" { // explicit system for this
		err = stream.writeUint16BE(uint16(len(context.StaticSchema.Properties)))
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(layer.Bool1)
		if err != nil {
			return err
		}
		// If this packet is to the server, write version
		if layer.Bool1 && !writer.ToClient() {
			err = stream.writeSintUTF8(layer.Int1)
			if err != nil {
				return err
			}
		}

		return stream.writeObject(layer.Value.(datamodel.ValueReference).Instance, writer.Caches())
	}

	err = stream.writeUint16BE(uint16(context.StaticSchema.PropertiesByName[layer.Instance.ClassName+"."+layer.PropertyName]))
	if err != nil {
		return err
	}

	err = stream.writeBoolByte(layer.Bool1)
	if err != nil {
		return err
	}
	if layer.Bool1 && !writer.ToClient() {
		err = stream.writeSintUTF8(layer.Int1)
		if err != nil {
			return err
		}
	}

	println("serializing property", layer.PropertyName, layer.Instance.Name(), layer.Value.String(), uint16(context.StaticSchema.PropertiesByName[layer.Instance.ClassName+"."+layer.PropertyName]))

	// Shun Go for silently ignoring nil map values and just returning 0 instead
	// TODO improve this
	propertyID, ok := context.StaticSchema.PropertiesByName[layer.Instance.ClassName+"."+layer.PropertyName]
	if !ok {
		return errors.New("unrecognized property " + layer.Instance.ClassName + "." + layer.PropertyName)
	}

	// TODO: A different system for this?
	err = context.StaticSchema.Properties[propertyID].Serialize(layer.Value, writer, stream)
	return err
}

func (Packet83_03) Type() uint8 {
	return 3
}
func (Packet83_03) TypeString() string {
	return "ID_REPLIC_PROP"
}
