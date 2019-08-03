package peer

import (
	"errors"

	"github.com/robloxapi/rbxfile"
)

// Decode deserializes a network property change packet
func (schema *NetworkPropertySchema) Decode(reader PacketReader, stream serializeReader, layers *PacketLayers, deferred deferredStrings) (rbxfile.Value, error) {
	val, err := stream.ReadSerializedValue(reader, schema.Type, schema.EnumID, deferred)
	if err != nil {
		return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
	}
	if val.Type() != rbxfile.TypeProtectedString {
		layers.Root.Logger.Println("read", schema.Name, val.String())
	}
	return val, nil
}

// Serialize serializes a property change packet to its network format
func (schema *NetworkPropertySchema) Serialize(value rbxfile.Value, writer PacketWriter, stream serializeWriter, deferred writeDeferredStrings) error {
	return stream.WriteSerializedValue(value, writer, schema.Type, deferred)
}
