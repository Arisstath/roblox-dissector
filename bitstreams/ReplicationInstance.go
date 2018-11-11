package bitstreams

import (
	"errors"
	"fmt"

	"github.com/gskartwii/rbxfile"
)

func is2ndRoundType(typeId uint8) bool {
	id := uint32(typeId)
	return ((id-3) > 0x1F || ((1<<(id-3))&uint32(0xC200000F)) == 0) && (id != 1) // thank you ARM compiler for optimizing this <3
}

func DecodeReplicationInstance(reader PacketReader, thisBitstream InstanceReader, layers *PacketLayers) (*Reference, error) {
	var err error
	var referent *Reference
	context := reader.Context()

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return nil, errors.New("while parsing self: " + err.Error())
	}
	if referent.IsNull() {
		return nil, errors.New("self is nil in decodeReplicationInstance")
	}
	thisInstance, err := context.InstancesByReferent.CreateInstance(referent)
	if err != nil {
		return nil, err
	}

	schemaIDx, err := thisBitstream.ReadUint16BE()
	if int(schemaIDx) > len(context.StaticSchema.Instances) {
		return referent, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(context.StaticSchema.Instances))
	}
	schema := context.StaticSchema.Instances[schemaIDx]
	thisInstance.ClassName = schema.Name
	layers.Root.Logger.Println("will parse", referent, schema.Name, len(schema.Properties))

	unkBool, err := thisBitstream.ReadBoolByte()
	if err != nil {
		return referent, err
	}
	layers.Root.Logger.Println("unkbool:", unkBool)
	thisInstance.Properties = make(map[string]rbxfile.Value, len(schema.Properties))

	err = thisBitstream.ReadProperties(schema.Properties, thisInstance.Properties, reader)
	if err != nil {
		return referent, err
	}

    parentRef, err := thisBitstream.ReadObject(reader)
	if err != nil {
		return referent, errors.New("while parsing parent: " + err.Error())
	}
	if referent.IsNull() {
		return referent, errors.New("parent is null")
	}
	if len(referent) > 0x50 {
		layers.Root.Logger.Println("Parent: (invalid), ", len(referent))
	} else {
		layers.Root.Logger.Println("Parent: ", referent)
	}

	context.InstancesByReferent.AddInstance(referent, thisInstance)
	parent, err := context.InstancesByReferent.TryGetInstance(parentRef)
	if parent != nil {
		return referent, parent.AddChild(thisInstance)
	}
	if err != nil && !thisInstance.IsService {
		return referent, errors.New("not service yet parent doesn't exist") // the parents of services don't exist
	}

	return referent, nil
}

func SerializeReplicationInstance(reference *Reference, writer PacketWriter, stream InstanceWriter) error {
    // TODO: PacketWriter implement GetInstance() and CreateInstance()?
    instance, err := writer.Context().InstancesByReferent.TryGetInstance(reference)
    if err != nil {
        return err
    }
	if instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err = stream.WriteObject(instance, writer)
	if err != nil {
		return err
	}

	context := writer.Context()
	schemaIdx := uint16(context.StaticSchema.ClassesByName[instance.ClassName])
	err = stream.WriteUint16BE(schemaIdx)
	if err != nil {
		return err
	}
	err = stream.WriteBoolByte(true) // ???
	if err != nil {
		return err
	}

	schema := context.StaticSchema.Instances[schemaIdx]
	err = stream.WriteProperties(schema.Properties, instance.Properties, writer)
	if err != nil {
		return err
	}

	return stream.WriteObject(instance.Parent(), writer)
}
