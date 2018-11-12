package bitstreams

import (
	"errors"

	"github.com/gskartwii/rbxfile"
    "github.com/gskartwii/roblox-dissector/schema"
    "github.com/gskartwii/roblox-dissector/util"
)

func DecodeReplicationProperty(reader util.PacketReader, stream SerializeReader, schema schema.StaticPropertySchema) (rbxfile.Value, error) {
	val, err := stream.ReadSerializedValue(reader, schema.Type, schema.EnumID)
	if err != nil {
		return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
	}
	return val, nil
}

// TODO: Better system?
func SerializeReplicationProperty(value rbxfile.Value, writer util.PacketWriter, stream SerializeWriter, schema schema.StaticPropertySchema) error {
	return stream.WriteSerializedValue(value, writer, schema.Type)
}
