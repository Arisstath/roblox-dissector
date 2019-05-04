package peer

import (
	"errors"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

type SerializeReader interface {
	ReadSerializedValue(reader PacketReader, valType uint8, enumId uint16) (rbxfile.Value, error)
	ReadObject(reader PacketReader) (datamodel.Reference, error)

	// We must also ask for the following methods for compatibility reasons.
	// Any better way to do this? I can't tell Go that the interface
	// will always implement everything from *extendedReader...
	readUint16BE() (uint16, error)
	readBoolByte() (bool, error)
	readUint8() (uint8, error)
}
type InstanceReader interface {
	SerializeReader
	ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader) error
}

type SerializeWriter interface {
	WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valType uint8) error
	WriteObject(object *datamodel.Instance, writer PacketWriter) error

	writeUint16BE(uint16) error
	writeBoolByte(bool) error
	WriteByte(byte) error
}
type InstanceWriter interface {
	SerializeWriter
	WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter) error
}

func (b *extendedReader) ReadSerializedValue(reader PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PROP_TYPE_STRING:
		result, err = b.readNewPString(reader.Caches())
	case PROP_TYPE_PROTECTEDSTRING_0:
		result, err = b.readNewProtectedString(reader.Caches())
	case PROP_TYPE_PROTECTEDSTRING_1:
		result, err = b.readNewProtectedString(reader.Caches())
	case PROP_TYPE_PROTECTEDSTRING_2:
		result, err = b.readNewProtectedString(reader.Caches())
	case PROP_TYPE_PROTECTEDSTRING_3:
		result, err = b.readNewProtectedString(reader.Caches())
	case PROP_TYPE_INSTANCE:
		var reference datamodel.Reference
		reference, err = b.ReadObject(reader)
		if err != nil {
			return nil, err
		}
		// Note: NULL is a valid reference!
		if reference.IsNull {
			result = datamodel.ValueReference{Instance: nil, Reference: reference}
		} else {
			// CreateInstance: allow forward references in ID_NEW_INST or ID_PROP
			// TODO: too tolerant?
			var instance *datamodel.Instance
			instance, err = reader.Context().InstancesByReference.CreateInstance(reference)
			result = datamodel.ValueReference{Instance: instance, Reference: reference}
		}
	case PROP_TYPE_CONTENT:
		result, err = b.readNewContent(reader.Caches())
	case PROP_TYPE_SYSTEMADDRESS:
		result, err = b.readSystemAddress(reader.Caches())
	case PROP_TYPE_TUPLE:
		result, err = b.readNewTuple(reader)
	case PROP_TYPE_ARRAY:
		result, err = b.readNewArray(reader)
	case PROP_TYPE_DICTIONARY:
		result, err = b.readNewDictionary(reader)
	case PROP_TYPE_MAP:
		result, err = b.readNewMap(reader)
	default:
		return b.readSerializedValueGeneric(reader, valueType, enumId)
	}
	return result, err
}
func (b *extendedReader) ReadObject(reader PacketReader) (datamodel.Reference, error) {
	return b.readObject(reader.Caches())
}
func (b *extendedReader) ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader) error {
	for i := 0; i < 2; i++ {
		propertyIndex, err := b.readUint8()
		last := "none"
		for err == nil && propertyIndex != 0xFF {
			if int(propertyIndex) > len(schema) {
				return errors.New("prop index oob, last was " + last)
			}

			value, err := b.ReadSerializedValue(reader, schema[propertyIndex].Type, schema[propertyIndex].EnumID)
			if err != nil {
				return err
			}
			properties[schema[propertyIndex].Name] = value
			last = schema[propertyIndex].Name
			propertyIndex, err = b.readUint8()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedWriter) WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valueType uint8) error {
	var err error
	switch valueType {
	case PROP_TYPE_STRING:
		err = b.writeNewPString(val.(rbxfile.ValueString), writer.Caches())
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PROP_TYPE_INSTANCE:
		err = b.writeObject(val.(datamodel.ValueReference).Instance, writer.Caches())
	case PROP_TYPE_CONTENT:
		err = b.writeNewContent(val.(rbxfile.ValueContent), writer.Caches())
	case PROP_TYPE_SYSTEMADDRESS:
		err = b.writeSystemAddress(val.(datamodel.ValueSystemAddress), writer.Caches())
	case PROP_TYPE_TUPLE:
		err = b.writeNewTuple(val.(datamodel.ValueTuple), writer)
	case PROP_TYPE_ARRAY:
		err = b.writeNewArray(val.(datamodel.ValueArray), writer)
	case PROP_TYPE_DICTIONARY:
		err = b.writeNewDictionary(val.(datamodel.ValueDictionary), writer)
	case PROP_TYPE_MAP:
		err = b.writeNewMap(val.(datamodel.ValueMap), writer)
	default:
		return b.writeSerializedValueGeneric(val, valueType)
	}
	return err
}
func (b *extendedWriter) WriteObject(object *datamodel.Instance, writer PacketWriter) error {
	return b.writeObject(object, writer.Caches())
}
func (b *extendedWriter) WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter) error {
	var err error
	for i := 0; i < len(schema); i++ {
		if is2ndRoundType(schema[i].Type) {
			continue
		}
		name := schema[i].Name
		value, ok := properties[name]

		if !ok {
			continue
		}

		err := b.WriteByte(uint8(i))
		if err != nil {
			return err
		}

		err = schema[i].Serialize(value, writer, b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(0xFF)
	if err != nil {
		return err
	}
	for i := 0; i < len(schema); i++ {
		if !is2ndRoundType(schema[i].Type) {
			continue
		}
		name := schema[i].Name
		value, ok := properties[name]

		if !ok {
			continue
		}

		err = b.WriteByte(uint8(i))
		if err != nil {
			return err
		}

		err = schema[i].Serialize(value, writer, b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(0xFF)
	return err
}

type joinSerializeReader struct {
	*extendedReader
}

func (b *joinSerializeReader) ReadSerializedValue(reader PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PROP_TYPE_STRING:
		result, err = b.readNewPString()
	case PROP_TYPE_PROTECTEDSTRING_0:
		result, err = b.readNewProtectedString()
	case PROP_TYPE_PROTECTEDSTRING_1:
		result, err = b.readNewProtectedString()
	case PROP_TYPE_PROTECTEDSTRING_2:
		result, err = b.readNewProtectedString()
	case PROP_TYPE_PROTECTEDSTRING_3:
		result, err = b.readNewProtectedString()
	case PROP_TYPE_INSTANCE:
		var reference datamodel.Reference
		reference, err = b.readJoinObject(reader.Context())
		if err != nil {
			return datamodel.ValueReference{Instance: nil, Reference: reference}, err
		}
		// Note: NULL is a valid reference!
		if reference.IsNull {
			result = datamodel.ValueReference{Instance: nil, Reference: reference}
			break
		}
		// CreateInstance: allow forward references
		var instance *datamodel.Instance
		instance, err = reader.Context().InstancesByReference.CreateInstance(reference)
		result = datamodel.ValueReference{Instance: instance, Reference: reference}
	case PROP_TYPE_CONTENT:
		result, err = b.readNewContent()
	case PROP_TYPE_SYSTEMADDRESS:
		result, err = b.readSystemAddress()
	default:
		return b.extendedReader.readSerializedValueGeneric(reader, valueType, enumId)
	}
	return result, err
}
func (b *joinSerializeReader) ReadObject(reader PacketReader) (datamodel.Reference, error) {
	return b.readJoinObject(reader.Context())
}
func (b *joinSerializeReader) ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader) error {
	propertyIndex, err := b.readUint8()
	last := "none"
	for err == nil && propertyIndex != 0xFF {
		if int(propertyIndex) > len(schema) {
			return errors.New("prop index oob, last was " + last)
		}

		value, err := b.ReadSerializedValue(reader, schema[propertyIndex].Type, schema[propertyIndex].EnumID)
		if err != nil {
			return err
		}
		properties[schema[propertyIndex].Name] = value
		last = schema[propertyIndex].Name
		propertyIndex, err = b.readUint8()
	}
	return err
}

type joinSerializeWriter struct {
	*extendedWriter
}

func (b *joinSerializeWriter) WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valueType uint8) error {
	var err error
	switch valueType {
	case PROP_TYPE_STRING:
		err = b.writeNewPString(val.(rbxfile.ValueString))
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PROP_TYPE_INSTANCE:
		err = b.WriteObject(val.(datamodel.ValueReference).Instance, writer)
	case PROP_TYPE_CONTENT:
		err = b.writeNewContent(val.(rbxfile.ValueContent))
	case PROP_TYPE_SYSTEMADDRESS:
		err = b.writeSystemAddress(val.(datamodel.ValueSystemAddress))
	default:
		return b.writeSerializedValueGeneric(val, valueType)
	}
	return err
}
func (b *joinSerializeWriter) WriteObject(object *datamodel.Instance, writer PacketWriter) error {
	return b.extendedWriter.writeJoinObject(object, writer.Context())
}
func (b *joinSerializeWriter) WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter) error {
	var err error
	for i := 0; i < len(schema); i++ {
		name := schema[i].Name
		value, ok := properties[name]

		if !ok {
			continue
		}

		err = b.WriteByte(uint8(i))
		if err != nil {
			return err
		}

		err = schema[i].Serialize(value, writer, b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(0xFF)
	return err
}
