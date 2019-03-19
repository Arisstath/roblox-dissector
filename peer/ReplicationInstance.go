package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

type ReplicationInstance struct {
	Instance           *datamodel.Instance
	Schema             *StaticInstanceSchema
	DeleteOnDisconnect bool
}

func NewReplicationInstance() *ReplicationInstance {
	return &ReplicationInstance{}
}

func is2ndRoundType(typeId uint8) bool {
	id := uint32(typeId)
	return ((id-3) > 0x1F || ((1<<(id-3))&uint32(0xC200000F)) == 0) && (id != 1) // thank you ARM compiler for optimizing this <3
}

func decodeReplicationInstance(reader PacketReader, thisBitstream InstanceReader, layers *PacketLayers) (*ReplicationInstance, error) {
	var err error
	repInstance := NewReplicationInstance()
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
	repInstance = thisInstance

	schemaIDx, err := thisBitstream.readUint16BE()
	if int(schemaIDx) > len(context.StaticSchema.Instances) {
		return repInstance, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(context.StaticSchema.Instances))
	}
	schema := context.StaticSchema.Instances[schemaIDx]
	repInstance.Schema = schema
	thisInstance.ClassName = schema.Name
	layers.Root.Logger.Println("will parse", referent.String(), schema.Name, len(schema.Properties))

	thisInstance.DeleteOnDisconnect, err = thisBitstream.readBoolByte()
	if err != nil {
		return repInstance, err
	}

	err = thisBitstream.ReadProperties(schema.Properties, repInstance.Properties, reader)
	if err != nil {
		return repInstance, err
	}

	referent, err = thisBitstream.ReadObject(reader)
	if err != nil {
		return repInstance, errors.New("while parsing parent: " + err.Error())
	}
	if referent.IsNull {
		return repInstance, errors.New("parent is null")
	}
	if len(referent.String()) > 0x50 {
		layers.Root.Logger.Println("Parent: (invalid), ", len(referent.String()))
	} else {
		layers.Root.Logger.Println("Parent: ", referent.String())
	}
	parent, err := context.InstancesByReferent.TryGetInstance(referent)
	if err != nil {
		return repInstance, err
	}
	parent.AddChild(thisInstance)

	return thisInstance, nil
}

func (instance *ReplicationInstance) Serialize(writer PacketWriter, stream InstanceWriter) error {
	var err error
	if instance == nil || instance.Instance {
		return errors.New("self is nil in serialize repl inst")
	}
	err = stream.WriteObject(instance.Instance, writer)
	if err != nil {
		return err
	}

	err = stream.writeUint16BE(instance.Schema.NetworkID)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(instance.DeleteOnDisconnect)
	if err != nil {
		return err
	}

	err = stream.WriteProperties(instance.Schema.Properties, instance.Instance.Properties, writer)
	if err != nil {
		return err
	}

	return stream.WriteObject(instance.Instance.Parent(), writer)
}
