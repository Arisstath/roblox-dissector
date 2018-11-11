package bitstreams

import (
	"errors"

	"github.com/gskartwii/rbxfile"
    "github.com/gskartwii/roblox-dissector/schema"
)

func DecodeReplicationProperty(reader PacketReader, stream SerializeReader, layers *PacketLayers, schema schema.StaticPropertySchema) (rbxfile.Value, error) {
	val, err := stream.ReadSerializedValue(reader, schema.Type, schema.EnumID)
	if err != nil {
		return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
	}
	if val.Type().String() != "ProtectedString" {
		layers.Root.Logger.Println("read", schema.Name, val.String())
	}
	return val, nil
}

// TODO: Better system?
func SerializeReplicationProperty(value rbxfile.Value, writer PacketWriter, stream SerializeWriter, schema schema.StaticPropertySchema) error {
	return stream.WriteSerializedValue(value, writer, schema.Type)
}
