package bitstreams

import (
	"errors"

	"github.com/gskartwii/rbxfile"
    "github.com/gskartwii/roblox-dissector/schema"
    "github.com/gskartwii/roblox-dissector/util"
)

type SerializeReader interface {
	ReadSerializedValue(reader util.PacketReader, valType uint8, enumId uint16) (rbxfile.Value, error)
	ReadObject(reader util.PacketReader) (util.Reference, error)

	// We must also ask for the following methods for compatibility reasons.
	// Any better way to do this? I can't tell Go that the interface
	// will always implement everything from *BitstreamReader...
	ReadUint16BE() (uint16, error)
	ReadBoolByte() (bool, error)
	ReadUint8() (uint8, error)
}
type InstanceReader interface {
	SerializeReader
	ReadProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, reader util.PacketReader) error
}

type SerializeWriter interface {
	WriteSerializedValue(val rbxfile.Value, writer util.PacketWriter, valType uint8) error
	WriteObject(object util.Reference, writer util.PacketWriter) error

	WriteUint16BE(uint16) error
	WriteBoolByte(bool) error
	WriteByte(uint8) error
}
type InstanceWriter interface {
	SerializeWriter
	WriteProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, writer util.PacketWriter) error
}

func (b *BitstreamReader) ReadSerializedValue(reader util.PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case schema.PROP_TYPE_STRING:
		result, err = b.ReadNewPString(reader.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_0:
		result, err = b.ReadNewProtectedString(reader.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_1:
		result, err = b.ReadNewProtectedString(reader.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_2:
		result, err = b.ReadNewProtectedString(reader.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_3:
		result, err = b.ReadNewProtectedString(reader.Caches())
	case schema.PROP_TYPE_INSTANCE:
		var referent util.Reference
		referent, err = b.ReadObject(reader)
		if err != nil {
			return nil, err
		}
		// Note: NULL is a valid referent!
		instance, _ := reader.Context().InstancesByReferent.TryGetInstance(referent)
		result = rbxfile.ValueReference{instance}
	case schema.PROP_TYPE_CONTENT:
		result, err = b.ReadNewContent(reader.Caches())
	case schema.PROP_TYPE_SYSTEMADDRESS:
		result, err = b.ReadSystemAddress(reader.Caches())
	case schema.PROP_TYPE_TUPLE:
		result, err = b.ReadNewTuple(reader)
	case schema.PROP_TYPE_ARRAY:
		result, err = b.ReadNewArray(reader)
	case schema.PROP_TYPE_DICTIONARY:
		result, err = b.ReadNewDictionary(reader)
	case schema.PROP_TYPE_MAP:
		result, err = b.ReadNewMap(reader)
	default:
		return b.ReadSerializedValueGeneric(reader, valueType, enumId)
	}
	return result, err
}
func (b *BitstreamReader) ReadObject(reader util.PacketReader) (util.Reference, error) {
	return b.ReadReference(reader.Caches())
}
func (b *BitstreamReader) ReadProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, reader util.PacketReader) error {
	for i := 0; i < 2; i++ {
		propertyIndex, err := b.ReadUint8()
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
			propertyIndex, err = b.ReadUint8()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BitstreamWriter) WriteSerializedValue(val rbxfile.Value, writer util.PacketWriter, valueType uint8) error {
	if val == nil {
		return nil
	}
	var err error
	switch valueType {
	case schema.PROP_TYPE_STRING:
		err = b.WriteNewPString(val.(rbxfile.ValueString), writer.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_0:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_1:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_2:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case schema.PROP_TYPE_PROTECTEDSTRING_3:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), writer.Caches())
	case schema.PROP_TYPE_INSTANCE:
		err = b.WriteObject(val.(rbxfile.ValueReference).Instance, writer.Caches())
	case schema.PROP_TYPE_CONTENT:
		err = b.WriteNewContent(val.(rbxfile.ValueContent), writer.Caches())
	case schema.PROP_TYPE_SYSTEMADDRESS:
		err = b.WriteSystemAddress(val.(rbxfile.ValueSystemAddress), writer.Caches())
	case schema.PROP_TYPE_TUPLE:
		err = b.WriteNewTuple(val.(rbxfile.ValueTuple), writer)
	case schema.PROP_TYPE_ARRAY:
		err = b.WriteNewArray(val.(rbxfile.ValueArray), writer)
	case schema.PROP_TYPE_DICTIONARY:
		err = b.WriteNewDictionary(val.(rbxfile.ValueDictionary), writer)
	case schema.PROP_TYPE_MAP:
		err = b.WriteNewMap(val.(rbxfile.ValueMap), writer)
	default:
		return b.WriteSerializedValueGeneric(val, valueType)
	}
	return err
}
func (b *BitstreamWriter) WriteObject(object util.Reference, writer util.PacketWriter) error {
	return b.WriteReference(object, writer.Caches())
}
func (b *BitstreamWriter) WriteProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, writer util.PacketWriter) error {
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

		//println("serializing", i, name, value.String())
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

type JoinSerializeReader struct {
	*BitstreamReader
}

func (b *JoinSerializeReader) ReadNewContent() (rbxfile.ValueContent, error) {
	res, err := b.ReadNewPString()
	return rbxfile.ValueContent(res), err
}
func (b *JoinSerializeReader) ReadNewProtectedString() (rbxfile.ValueProtectedString, error) {
	res, err := b.ReadNewPString()
	return rbxfile.ValueProtectedString(res), err
}
func (b *JoinSerializeReader) ReadNewPString() (rbxfile.ValueString, error) {
	val, err := b.ReadVarLengthString()
	return rbxfile.ValueString(val), err
}
func (b *JoinSerializeReader) ReadContent() (rbxfile.ValueContent, error) {
	var result string
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return rbxfile.ValueContent(result), err
	}
	result, err = b.ReadASCII(int(stringLen))
	return rbxfile.ValueContent(result), err
}
func (b *JoinSerializeReader) ReadSystemAddress() (rbxfile.ValueSystemAddress, error) {
	var err error
	thisAddress := rbxfile.ValueSystemAddress("0.0.0.0:0")
	thisAddr := net.UDPAddr{}
	thisAddr.IP = make([]byte, 4)
	err = b.bytes(thisAddr.IP, 4)
	if err != nil {
		return thisAddress, err
	}
	for i := 0; i < 4; i++ {
		thisAddr.IP[i] = thisAddr.IP[i] ^ 0xFF // bitwise NOT
	}

	port, err := b.ReadUint16BE()
	thisAddr.Port = int(port)
	if err != nil {
		return thisAddress, err
	}
	return rbxfile.ValueSystemAddress(thisAddr.String()), nil
}

func (b *JoinSerializeReader) ReadSerializedValue(reader util.PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case schema.PROP_TYPE_STRING:
		result, err = b.ReadNewPString()
	case schema.PROP_TYPE_PROTECTEDSTRING_0:
		result, err = b.ReadNewProtectedString()
	case schema.PROP_TYPE_PROTECTEDSTRING_1:
		result, err = b.ReadNewProtectedString()
	case schema.PROP_TYPE_PROTECTEDSTRING_2:
		result, err = b.ReadNewProtectedString()
	case schema.PROP_TYPE_PROTECTEDSTRING_3:
		result, err = b.ReadNewProtectedString()
	case schema.PROP_TYPE_INSTANCE:
		var referent util.Reference
		referent, err = b.ReadJoinObject(reader.Context())
		if err != nil {
			return nil, err
		}
		// Note: NULL is a valid referent!
		instance, _ := reader.Context().InstancesByReferent.TryGetInstance(referent)
		result = rbxfile.ValueReference{instance}
	case schema.PROP_TYPE_CONTENT:
		result, err = b.ReadNewContent()
	case schema.PROP_TYPE_SYSTEMADDRESS:
		result, err = b.ReadSystemAddress()
	default:
		return b.BitstreamReader.ReadSerializedValueGeneric(reader, valueType, enumId)
	}
	return result, err
}
func (b *JoinSerializeReader) ReadObject(reader util.PacketReader) (util.Reference, error) {
	return b.ReadJoinObject(reader.Context())
}
func (b *JoinSerializeReader) ReadProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, reader util.PacketReader) error {
	propertyIndex, err := b.ReadUint8()
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
		propertyIndex, err = b.ReadUint8()
	}
	return err
}

type JoinSerializeWriter struct {
	*BitstreamWriter
}

func (b *JoinSerializeWriter) WriteSerializedValue(val rbxfile.Value, writer util.PacketWriter, valueType uint8) error {
	if val == nil {
		return nil
	}
	var err error
	switch valueType {
	case schema.PROP_TYPE_STRING:
		err = b.WriteNewPString(val.(rbxfile.ValueString))
	case schema.PROP_TYPE_PROTECTEDSTRING_0:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString))
	case schema.PROP_TYPE_PROTECTEDSTRING_1:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString))
	case schema.PROP_TYPE_PROTECTEDSTRING_2:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString))
	case schema.PROP_TYPE_PROTECTEDSTRING_3:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString))
	case schema.PROP_TYPE_INSTANCE:
		err = b.WriteObject(val.(rbxfile.ValueReference).Instance, writer)
	case schema.PROP_TYPE_CONTENT:
		err = b.WriteNewContent(val.(rbxfile.ValueContent))
	case schema.PROP_TYPE_SYSTEMADDRESS:
		err = b.WriteSystemAddress(val.(rbxfile.ValueSystemAddress))
	default:
		return b.WriteSerializedValueGeneric(val, valueType)
	}
	return err
}
func (b *JoinSerializeWriter) WriteObject(reference util.Reference, writer util.PacketWriter) error {
	return b.BitstreamWriter.WriteJoinObject(object, writer.Context())
}
func (b *JoinSerializeWriter) WriteProperties(schema []schema.StaticPropertySchema, properties map[string]rbxfile.Value, writer util.PacketWriter) error {
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

		//println("serializing", i, name, value.String())
		err = schema[i].Serialize(value, writer, b)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(0xFF)
	return err
}

func (b *JoinSerializeWriter) WriteNewPString(val rbxfile.ValueString) error {
	return b.BitstreamWriter.WritePStringNoCache(val)
}
func (b *JoinSerializeWriter) WriteNewProtectedString(val rbxfile.ValueProtectedString) error {
	return b.BitstreamWriter.WritePStringNoCache(rbxfile.ValueString(val))
}
func (b *JoinSerializeWriter) WriteNewContent(val rbxfile.ValueContent) error {
	return b.WriteUint32AndString(string(val))
}
