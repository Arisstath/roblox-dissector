package peer

import (
	"errors"

	"github.com/gskartwii/rbxfile"
)

func (schema StaticPropertySchema) Decode(reader PacketReader, stream SerializeReader) (rbxfile.Value, error) {
	val, err := stream.ReadSerializedValue(reader, schema.Type, schema.EnumID)
	if err != nil {
		return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
	}
	if val.Type().String() != "ProtectedString" {
		packet.Logger.Println("read", schema.Name, val.String())
	}
	return val, nil
}

// TODO: Better system?
func (schema StaticPropertySchema) Serialize(value rbxfile.Value, writer PacketWriter, stream SerializeWriter) error {
	return stream.WriteSerializedValue(value, writer, schema.Type)
}
