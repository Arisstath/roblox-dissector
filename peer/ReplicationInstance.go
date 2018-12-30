package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

func is2ndRoundType(typeId uint8) bool {
	id := uint32(typeId)
	return ((id-3) > 0x1F || ((1<<(id-3))&uint32(0xC200000F)) == 0) && (id != 1) // thank you ARM compiler for optimizing this <3
}

func decodeReplicationInstance(reader PacketReader, thisBitstream InstanceReader, layers *PacketLayers) (*datamodel.Instance, error) {
	var err error
	var referent datamodel.Reference
	context := reader.Context()

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return nil, errors.New("while parsing self: " + err.Error())
	}
	if referent.IsNull {
		return nil, errors.New("self is nil in decodeReplicationInstance")
	}
	thisInstance, err := context.InstancesByReferent.CreateInstance(referent)
	if err != nil {
		return nil, err
	}

	schemaIDx, err := thisBitstream.readUint16BE()
	if int(schemaIDx) > len(context.StaticSchema.Instances) {
		return thisInstance, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(context.StaticSchema.Instances))
	}
	schema := context.StaticSchema.Instances[schemaIDx]
	thisInstance.ClassName = schema.Name
	layers.Root.Logger.Println("will parse", referent.String(), schema.Name, len(schema.Properties))

	unkBool, err := thisBitstream.readBoolByte()
	if err != nil {
		return thisInstance, err
	}
	layers.Root.Logger.Println("delete on disconnect:", unkBool)

	err = thisBitstream.ReadProperties(schema.Properties, thisInstance.Properties, reader)
	if err != nil {
		return thisInstance, err
	}

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return thisInstance, errors.New("while parsing parent: " + err.Error())
	}
	if referent.IsNull {
		return thisInstance, errors.New("parent is null")
	}
	if len(referent.String()) > 0x50 {
		layers.Root.Logger.Println("Parent: (invalid), ", len(referent.String()))
	} else {
		layers.Root.Logger.Println("Parent: ", referent.String())
	}

	context.InstancesByReferent.AddInstance(thisInstance.Ref, thisInstance)
	parent, err := context.InstancesByReferent.TryGetInstance(referent)
	if parent != nil {
		return thisInstance, parent.AddChild(thisInstance)
	}
	if err != nil && !thisInstance.IsService {
		println("couldn't find parent for", thisInstance.Ref.String(), referent.String())
		return thisInstance, err // the parents of services don't exist
	}

	return thisInstance, nil
}

func serializeReplicationInstance(instance *datamodel.Instance, writer PacketWriter, stream InstanceWriter) error {
	var err error
	if instance == nil {
		return errors.New("self is nil in serialize repl inst")
	}
	err = stream.WriteObject(instance, writer)
	if err != nil {
		return err
	}

	context := writer.Context()
	schemaIdx := uint16(context.StaticSchema.ClassesByName[instance.ClassName])
	err = stream.writeUint16BE(schemaIdx)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(true) // ???
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
