package peer

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet81LayerItem describes a global service from ID_SET_GLOBALS (Packet81Layer)
type Packet81LayerItem struct {
	Schema        *NetworkInstanceSchema
	Instance      *datamodel.Instance
	WatchChanges  bool
	WatchChildren bool
}

// Packet81Layer represents ID_SET_GLOBALS - server -> client
type Packet81Layer struct {
	// Is streaming enabled?
	StreamJob bool
	// Is Filtering enabled?
	FilteringEnabled     bool
	AllowThirdPartySales bool
	CharacterAutoSpawn   bool
	// Server's scope
	ReferenceString string
	ScriptKey       uint32
	CoreScriptKey   uint32
	// List of services to be created
	Items []*Packet81LayerItem
}

func (thisStream *extendedReader) DecodePacket81Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet81Layer{}

	var err error

	layer.StreamJob, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.FilteringEnabled, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.AllowThirdPartySales, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.CharacterAutoSpawn, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	stringLen, err := thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.ReferenceString, err = thisStream.readASCII(int(stringLen))
	if err != nil {
		return layer, err
	}
	// This assignment is justifiable because a call to readJoinObject() below depends on it
	reader.Context().InstanceTopScope = layer.ReferenceString
	if !reader.Context().IsStudio {
		layer.ScriptKey, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
		layer.CoreScriptKey, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}

		reader.Context().ScriptKey = layer.ScriptKey
		reader.Context().CoreScriptKey = layer.CoreScriptKey
	}

	arrayLen, err := thisStream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	if arrayLen > 0x1000 {
		return layer, errors.New("sanity check: exceeded maximum preschema len")
	}

	context := reader.Context()

	layer.Items = make([]*Packet81LayerItem, arrayLen)
	for i := 0; i < int(arrayLen); i++ {
		thisItem := &Packet81LayerItem{}
		reference, err := thisStream.readJoinObject(context)
		if err != nil {
			return layer, err
		}

		classID, err := thisStream.readUint16BE()
		if err != nil {
			return layer, err
		}

		if int(classID) > len(context.NetworkSchema.Instances) {
			return layer, fmt.Errorf("class idx %d is higher than %d", classID, len(context.NetworkSchema.Instances))
		}

		schema := context.NetworkSchema.Instances[classID]
		thisItem.Schema = schema
		instance, err := context.InstancesByReference.CreateInstance(reference)
		if err != nil {
			return layer, err
		}
		instance.ClassName = schema.Name
		instance.IsService = true
		thisItem.Instance = instance

		thisItem.WatchChanges, err = thisStream.readBoolByte()
		if err != nil {
			return layer, err
		}
		thisItem.WatchChildren, err = thisStream.readBoolByte()
		if err != nil {
			return layer, err
		}
		layer.Items[i] = thisItem
	}
	return layer, nil
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet81Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	err = stream.writeBoolByte(layer.StreamJob)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.FilteringEnabled)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.AllowThirdPartySales)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.CharacterAutoSpawn)
	if err != nil {
		return err
	}

	err = stream.writeUint32BE(uint32(len(layer.ReferenceString)))
	if err != nil {
		return err
	}
	err = stream.writeASCII(layer.ReferenceString)
	if err != nil {
		return err
	}

	if !writer.Context().IsStudio {
		err = stream.writeUint32BE(layer.ScriptKey)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(layer.CoreScriptKey)
		if err != nil {
			return err
		}
	}

	err = stream.writeUintUTF8(uint32(len(layer.Items)))
	for _, item := range layer.Items {
		err = stream.writeJoinObject(item.Instance, writer.Context())
		if err != nil {
			return err
		}
		err = stream.writeUint16BE(item.Schema.NetworkID)
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(item.WatchChanges)
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(item.WatchChildren)
		if err != nil {
			return err
		}
	}
	return nil
}

func (layer *Packet81Layer) String() string {
	return fmt.Sprintf("ID_SET_GLOBALS: %d services", len(layer.Items))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet81Layer) TypeString() string {
	return "ID_SET_GLOBALS"
}

// Type implements RakNetPacket.Type()
func (Packet81Layer) Type() byte {
	return 0x81
}
