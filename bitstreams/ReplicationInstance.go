package bitstreams

import (
	"errors"
	"fmt"

	"github.com/gskartwii/rbxfile"
	"github.com/gskartwii/roblox-dissector/util"
)

func is2ndRoundType(typeId uint8) bool {
	id := uint32(typeId)
	return ((id-3) > 0x1F || ((1<<(id-3))&uint32(0xC200000F)) == 0) && (id != 1) // thank you ARM compiler for optimizing this <3
}

func DecodeReplicationInstance(reader util.PacketReader, thisBitstream InstanceReader) (util.DeserializedInstance, error) {
	var err error
    var instance util.DeserializedInstance
	var referent util.Reference

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return instance, errors.New("while parsing self: " + err.Error())
	}
	if referent.IsNull {
		return instance, errors.New("self is nil in decodeReplicationInstance")
	}
	instance, err = reader.CreateInstance(referent)
	if err != nil {
		return instance, err
	}
    thisInstance := instance.Instance

	schemaIDx, err := thisBitstream.ReadUint16BE()
	if int(schemaIDx) > len(reader.Schema().Instances) {
		return instance, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(reader.Schema().Instances))
	}
	schema := reader.Schema().Instances[schemaIDx]
	thisInstance.ClassName = schema.Name

	_, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return instance, err
	}
	thisInstance.Properties = make(map[string]rbxfile.Value, len(schema.Properties))

	err = thisBitstream.ReadProperties(schema.Properties, thisInstance.Properties, reader)
	if err != nil {
		return instance, err
	}

    parentRef, err := thisBitstream.ReadObject(reader)
	if err != nil {
		return instance, errors.New("while parsing parent: " + err.Error())
	}
	if parentRef.IsNull {
		return instance, errors.New("parent is null")
	}

	parent, err := reader.TryGetInstance(parentRef)
    instance.Parent = parent.Reference
	if parent.Instance != nil {
		return instance, parent.AddChild(thisInstance)
	}
	if err != nil && !thisInstance.IsService {
		return instance, errors.New("not service yet parent doesn't exist") // the parents of services don't exist
	}

	return instance, nil
}

func SerializeReplicationInstance(reference util.Reference, writer util.PacketWriter, stream InstanceWriter) error {
    // TODO: PacketWriter implement GetInstance() and CreateInstance()?
    instance, err := writer.TryGetInstance(reference)
    if err != nil {
        return err
    }
	if instance.Instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err = stream.WriteObject(instance.Reference, writer)
	if err != nil {
		return err
	}

	schemaIdx := uint16(writer.Schema().ClassesByName[instance.ClassName])
	err = stream.WriteUint16BE(schemaIdx)
	if err != nil {
		return err
	}
	err = stream.WriteBoolByte(true) // ???
	if err != nil {
		return err
	}

	schema := writer.Schema().Instances[schemaIdx]
	err = stream.WriteProperties(schema.Properties, instance.Properties, writer)
	if err != nil {
		return err
	}

	return stream.WriteObject(instance.Parent, writer)
}
