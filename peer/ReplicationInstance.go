package peer

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

// ReplicationInstance describes a network instance creation packet
type ReplicationInstance struct {
	Instance           *datamodel.Instance
	Properties         map[string]rbxfile.Value
	Parent             *datamodel.Instance
	Schema             *NetworkInstanceSchema
	DeleteOnDisconnect bool
}

func is2ndRoundType(typeID uint8) bool {
	id := uint32(typeID)
	return ((id-3) > 0x1F || ((1<<(id-3))&uint32(0xC200000F)) == 0) && (id != 1) // thank you ARM compiler for optimizing this <3
}

func decodeReplicationInstance(reader PacketReader, thisStream instanceReader, layers *PacketLayers, deferred deferredStrings, readDefers bool) (*ReplicationInstance, error) {
	var err error
	repInstance := &ReplicationInstance{
		Properties: make(map[string]rbxfile.Value),
	}
	var reference datamodel.Reference
	context := reader.Context()

	reference, err = thisStream.ReadObject(reader)
	if err != nil {
		return nil, errors.New("while parsing self: " + err.Error())
	}
	if reference.IsNull {
		return nil, errors.New("self is nil in decodeReplicationInstance")
	}
	thisInstance, err := context.InstancesByReference.CreateInstance(reference)
	if err != nil {
		return nil, err
	}
	repInstance.Instance = thisInstance

	schemaIDx, err := thisStream.readUint16BE()
	if int(schemaIDx) > len(context.NetworkSchema.Instances) {
		return repInstance, fmt.Errorf("class idx %d is higher than %d", schemaIDx, len(context.NetworkSchema.Instances))
	}
	schema := context.NetworkSchema.Instances[schemaIDx]
	repInstance.Schema = schema
	thisInstance.ClassName = schema.Name
	layers.Root.Logger.Println("will parse", reference.String(), schema.Name, len(schema.Properties))

	repInstance.DeleteOnDisconnect, err = thisStream.readBoolByte()
	if err != nil {
		return repInstance, err
	}

	err = thisStream.ReadProperties(schema.Properties, repInstance.Properties, reader, deferred)
	if err != nil {
		return repInstance, err
	}

	if readDefers {
		err = thisStream.resolveDeferredStrings(deferred)
		if err != nil {
			return repInstance, err
		}
	}

	reference, err = thisStream.ReadObject(reader)
	if err != nil {
		return repInstance, errors.New("while parsing parent: " + err.Error())
	}
	if len(reference.String()) > 0x50 {
		layers.Root.Logger.Println("Parent: (invalid), ", len(reference.String()))
	} else {
		layers.Root.Logger.Println("Parent: ", reference.String())
	}
	parent, err := context.InstancesByReference.TryGetInstance(reference)
	if err != nil {
		// service parents aren't excepted to exist
		if err == datamodel.ErrInstanceDoesntExist && thisInstance.IsService {
			// create a dummy instance for DataModel
			parent, err := context.InstancesByReference.CreateInstance(reference)
			if err != nil {
				return nil, err
			}
			parent.ClassName = "DataModel"
			repInstance.Parent = parent
		} else {
			return repInstance, err
		}
	}
	repInstance.Parent = parent

	return repInstance, nil
}

// Serialize serializes an instance creation packet to its network format
func (instance *ReplicationInstance) Serialize(writer PacketWriter, stream instanceWriter, deferred writeDeferredStrings, writeDefers bool) error {
	var err error
	if instance == nil || instance.Instance == nil {
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

	err = stream.WriteProperties(instance.Schema.Properties, instance.Properties, writer, deferred)
	if err != nil {
		return err
	}

	if writeDefers {
		err = stream.resolveDeferredStrings(deferred)
		if err != nil {
			return err
		}
	}

	return stream.WriteObject(instance.Parent, writer)
}
