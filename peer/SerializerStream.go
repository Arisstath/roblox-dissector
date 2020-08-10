package peer

import (
	"errors"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

type serializeReader interface {
	ReadSerializedValue(reader PacketReader, valType uint8, enumID uint16, deferred deferredStrings) (rbxfile.Value, error)
	ReadObject(reader PacketReader) (datamodel.Reference, error)

	// We must also ask for the following methods for compatibility reasons.
	// Any better way to do this? I can't tell Go that the interface
	// will always implement everything from *extendedReader...
	readUint16BE() (uint16, error)
	readBoolByte() (bool, error)
	readUint8() (uint8, error)
}
type instanceReader interface {
	serializeReader
	ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader, deferred deferredStrings) error
	resolveDeferredStrings(deferred deferredStrings) error
}

type serializeWriter interface {
	WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valType uint8, deferred writeDeferredStrings) error
	WriteObject(object *datamodel.Instance, writer PacketWriter) error

	writeUint16BE(uint16) error
	writeBoolByte(bool) error
	WriteByte(byte) error
}
type instanceWriter interface {
	serializeWriter
	WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter, deferred writeDeferredStrings) error
	resolveDeferredStrings(deferred writeDeferredStrings) error
}

func (b *extendedReader) ReadSerializedValue(reader PacketReader, valueType uint8, enumID uint16, deferred deferredStrings) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PropertyTypeString:
		result, err = b.readNewPString(reader.Caches())
	case PropertyTypeProtectedString0:
		result, err = b.readNewProtectedString(reader.Caches())
	case PropertyTypeProtectedString1:
		result, err = b.readNewProtectedString(reader.Caches())
	case PropertyTypeProtectedString2:
		result, err = b.readNewProtectedString(reader.Caches())
	case PropertyTypeProtectedString3:
		result, err = b.readNewProtectedString(reader.Caches())
	case PropertyTypeLuauString:
		result, err = b.readLuauProtectedString(reader.Caches())
	case PropertyTypeInstance:
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
			var instance *datamodel.Instance
			instance, err = reader.Context().InstancesByReference.CreateInstance(reference)
			result = datamodel.ValueReference{Instance: instance, Reference: reference}
		}
	case PropertyTypeContent:
		result, err = b.readNewContent(reader.Context())
	case PropertyTypeTuple:
		result, err = b.readNewTuple(reader, deferred)
	case PropertyTypeArray:
		result, err = b.readNewArray(reader, deferred)
	case PropertyTypeDictionary:
		result, err = b.readNewDictionary(reader, deferred)
	case PropertyTypeMap:
		result, err = b.readNewMap(reader, deferred)
	default:
		return b.readSerializedValueGeneric(reader, valueType, enumID, deferred)
	}
	return result, err
}
func (b *extendedReader) ReadObject(reader PacketReader) (datamodel.Reference, error) {
	return b.readObject(reader.Context(), reader.Caches())
}
func (b *extendedReader) ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader, deferred deferredStrings) error {
	for i := 0; i < 2; i++ {
		propertyIndex, err := b.readUint8()
		last := "none"
		for err == nil && propertyIndex != 0xFF {
			if int(propertyIndex) > len(schema) {
				return errors.New("prop index oob, last was " + last)
			}

			value, err := b.ReadSerializedValue(reader, schema[propertyIndex].Type, schema[propertyIndex].EnumID, deferred)
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

func (b *extendedWriter) WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valueType uint8, deferred writeDeferredStrings) error {
	var err error
	switch valueType {
	case PropertyTypeString:
		err = b.writeNewPString(val.(rbxfile.ValueString), writer.Caches())
	case PropertyTypeProtectedString0:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PropertyTypeProtectedString1:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PropertyTypeProtectedString2:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PropertyTypeProtectedString3:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case PropertyTypeLuauString:
		err = b.writeLuauProtectedString(val.(datamodel.ValueSignedProtectedString), writer.Caches())
	case PropertyTypeInstance:
		err = b.WriteObject(val.(datamodel.ValueReference).Instance, writer)
	case PropertyTypeContent:
		err = b.writeNewContent(val.(rbxfile.ValueContent), writer.Context())
	case PropertyTypeTuple:
		err = b.writeNewTuple(val.(datamodel.ValueTuple), writer, deferred)
	case PropertyTypeArray:
		err = b.writeNewArray(val.(datamodel.ValueArray), writer, deferred)
	case PropertyTypeDictionary:
		err = b.writeNewDictionary(val.(datamodel.ValueDictionary), writer, deferred)
	case PropertyTypeMap:
		err = b.writeNewMap(val.(datamodel.ValueMap), writer, deferred)
	default:
		return b.writeSerializedValueGeneric(val, valueType, deferred)
	}
	return err
}
func (b *extendedWriter) WriteObject(object *datamodel.Instance, writer PacketWriter) error {
	return b.writeObject(object, writer.Context(), writer.Caches())
}
func (b *extendedWriter) WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter, deferred writeDeferredStrings) error {
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

		err = schema[i].Serialize(value, writer, b, deferred)
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

		err = schema[i].Serialize(value, writer, b, deferred)
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

func (b *joinSerializeReader) ReadSerializedValue(reader PacketReader, valueType uint8, enumID uint16, deferred deferredStrings) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PropertyTypeString:
		result, err = b.readNewPString()
	case PropertyTypeProtectedString0:
		result, err = b.readNewProtectedString()
	case PropertyTypeProtectedString1:
		result, err = b.readNewProtectedString()
	case PropertyTypeProtectedString2:
		result, err = b.readNewProtectedString()
	case PropertyTypeProtectedString3:
		result, err = b.readNewProtectedString()
	case PropertyTypeLuauString:
		result, err = b.readLuauProtectedString()
	case PropertyTypeInstance:
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
	case PropertyTypeContent:
		result, err = b.readNewContent()
	default:
		return b.extendedReader.readSerializedValueGeneric(reader, valueType, enumID, deferred)
	}
	return result, err
}
func (b *joinSerializeReader) ReadObject(reader PacketReader) (datamodel.Reference, error) {
	return b.readJoinObject(reader.Context())
}
func (b *joinSerializeReader) ReadProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, reader PacketReader, deferred deferredStrings) error {
	propertyIndex, err := b.readUint8()
	last := "none"
	for err == nil && propertyIndex != 0xFF {
		if int(propertyIndex) > len(schema) {
			return errors.New("prop index oob, last was " + last)
		}

		value, err := b.ReadSerializedValue(reader, schema[propertyIndex].Type, schema[propertyIndex].EnumID, deferred)
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

func (b *joinSerializeWriter) WriteSerializedValue(val rbxfile.Value, writer PacketWriter, valueType uint8, deferred writeDeferredStrings) error {
	var err error
	switch valueType {
	case PropertyTypeString:
		err = b.writeNewPString(val.(rbxfile.ValueString))
	case PropertyTypeProtectedString0:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PropertyTypeProtectedString1:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PropertyTypeProtectedString2:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PropertyTypeProtectedString3:
		err = b.writeNewProtectedString(val.(rbxfile.ValueProtectedString))
	case PropertyTypeLuauString:
		err = b.writeLuauProtectedString(val.(datamodel.ValueSignedProtectedString))
	case PropertyTypeInstance:
		err = b.WriteObject(val.(datamodel.ValueReference).Instance, writer)
	case PropertyTypeContent:
		err = b.writeNewContent(val.(rbxfile.ValueContent))
	default:
		return b.writeSerializedValueGeneric(val, valueType, deferred)
	}
	return err
}
func (b *joinSerializeWriter) WriteObject(object *datamodel.Instance, writer PacketWriter) error {
	return b.extendedWriter.writeJoinObject(object, writer.Context())
}

func (b *joinSerializeWriter) WriteProperties(schema []*NetworkPropertySchema, properties map[string]rbxfile.Value, writer PacketWriter, deferred writeDeferredStrings) error {
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

		err = schema[i].Serialize(value, writer, b, deferred)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(0xFF)
	return err
}
