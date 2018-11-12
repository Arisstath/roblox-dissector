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

func DecodeReplicationInstance(reader util.PacketReader, thisBitstream InstanceReader) (util.Reference, error) {
	var err error
	var referent util.Reference

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return referent, errors.New("while parsing self: " + err.Error())
	}
	if referent.IsNull {
		return referent, errors.New("self is nil in decodeReplicationInstance")
	}
	thisInstance, err := reader.CreateInstance(referent)
	if err != nil {
		return referent, err
	}

	schemaIDx, err := thisBitstream.ReadUint16BE()
	if int(schemaIDx) > len(reader.Schema().Instances) {
		return referent, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(reader.Schema().Instances))
	}
	schema := reader.Schema().Instances[schemaIDx]
	thisInstance.ClassName = schema.Name

	unkBool, err := thisBitstream.ReadBoolByte()
	if err != nil {
		return referent, err
	}
	thisInstance.Properties = make(map[string]rbxfile.Value, len(schema.Properties))

	err = thisBitstream.ReadProperties(schema.Properties, thisInstance.Properties, reader)
	if err != nil {
		return referent, err
	}

    parentRef, err := thisBitstream.ReadObject(reader)
	if err != nil {
		return referent, errors.New("while parsing parent: " + err.Error())
	}
	if parentRef.IsNull {
		return referent, errors.New("parent is null")
	}

	parent, err := reader.TryGetInstance(parentRef)
	if parent != nil {
		return referent, parent.AddChild(thisInstance)
	}
	if err != nil && !thisInstance.IsService {
		return referent, errors.New("not service yet parent doesn't exist") // the parents of services don't exist
	}

	return referent, nil
}

func SerializeReplicationInstance(reference util.Reference, writer util.PacketWriter, stream InstanceWriter) error {
    // TODO: PacketWriter implement GetInstance() and CreateInstance()?
    instance, err := writer.TryGetInstance(reference)
    if err != nil {
        return err
    }
	if instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err = stream.WriteObject(reference, writer)
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

	return stream.WriteObject(instance.Parent(), writer)
}
