package packets

import (
	"errors"
	"fmt"

	"github.com/gskartwii/rbxfile"
)

// ChangeProperty describes an ID_CHANGE_PROPERTY data subpacket.
type ChangeProperty struct {
	// Instance that had the property change
	Instance    *rbxfile.Instance
	Bool1       bool
	Int1        int32
	DoWriteInt1 bool
	// Name of the property
	PropertyName string
	// New value
	Value rbxfile.Value
}

func (thisBitstream *PacketReaderBitstream) DecodeChangeProperty(reader PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	layer := &ChangeProperty{}

	referent, err := thisBitstream.readObject(reader.Caches())
	if err != nil {
		return layer, err
	}
	if referent.IsNull() {
		return layer, errors.New("self is null in repl property")
	}
	instance, err := reader.Context().InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return layer, err
	}
	layer.Instance = instance
	instance.Properties[layer.PropertyName] = layer.Value

	propertyIDx, err := thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}

	layer.Bool1, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	if layer.Bool1 && reader.IsClient() {
		layer.Int1, err = thisBitstream.readSintUTF8()
		if err != nil {
			return layer, err
		}
	}

	context := reader.Context()
	if int(propertyIDx) == int(len(context.StaticSchema.Properties)) { // explicit Parent property system
		var referent Referent
		referent, err = thisBitstream.readObject(reader.Caches())
		parent, err := context.InstancesByReferent.TryGetInstance(referent)
		if err != nil {
			return layer, errors.New("parent doesn't exist in repl property")
		}
		result := rbxfile.ValueReference{parent}
		layer.Value = result
		layer.PropertyName = "Parent"

		if referent.IsNull() { // NULL is a valid referent; think about :Remove()!
			return layer, instance.SetParent(nil)
		}
		err = parent.AddChild(instance)
		return layer, err
	}

	if int(propertyIDx) > int(len(context.StaticSchema.Properties)) {
		return layer, fmt.Errorf("prop idx %d is higher than %d", propertyIDx, len(context.StaticSchema.Properties))
	}
	schema := context.StaticSchema.Properties[propertyIDx]
	layer.PropertyName = schema.Name

	layer.Value, err = DecodeReplicationProperty(reader, thisBitstream, layers, schema)

	return layer, err
}

func (layer *ChangeProperty) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
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
		if writer.ToClient() {
			err = stream.writeSintUTF8(layer.Int1)
			if err != nil {
				return err
			}
		}

		return stream.writeObject(layer.Value.(rbxfile.ValueReference).Instance, writer.Caches())
	}

	err = stream.writeUint16BE(uint16(context.StaticSchema.PropertiesByName[layer.Instance.ClassName+"."+layer.PropertyName]))
	if err != nil {
		return err
	}

	err = stream.writeBoolByte(layer.Bool1)
	if err != nil {
		return err
	}
	if writer.ToClient() { // TODO: Serializers should be able to access PacketWriter
		err = stream.writeSintUTF8(layer.Int1)
		if err != nil {
			return err
		}
	}

	//println("serializing property", layer.PropertyName, layer.Instance.Name(), layer.Value.String())
	if layer.Instance == nil {
		return errors.New("cannot serialize property because instance is nil")
	}

	// Shun Go for silently ignoring nil map values and just returning 0 instead
	// TODO improve this
	propertyID, ok := context.StaticSchema.PropertiesByName[layer.Instance.ClassName+"."+layer.PropertyName]
	if !ok {
		return errors.New("unrecognized property " + layer.Instance.ClassName + "." + layer.PropertyName)
	}

	// TODO: A different system for this?
	err = SerializeReplicationProperty(layer.Value, writer, stream, context.StaticSchema.Properties[propertyID])
	return err
}

func (ChangeProperty) Type() uint8 {
	return 3
}
func (ChangeProperty) TypeString() string {
	return "ID_REPLIC_PROP"
}
