package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

// Describes a global service from ID_SET_GLOBALS (Packet81Layer)
type Packet81LayerItem struct {
	Schema   *StaticInstanceSchema
	Instance *datamodel.Instance
	Bool1    bool
	Bool2    bool
}

// ID_SET_GLOBALS - server -> client
type Packet81Layer struct {
	// Is streaming enabled?
	StreamJob bool
	// Is Filtering enabled?
	FilteringEnabled     bool
	AllowThirdPartySales bool
	CharacterAutoSpawn   bool
	// Server's scope
	ReferentString string
	Int1           uint32
	Int2           uint32
	// List of services to be set
	Items []*Packet81LayerItem
}

func NewPacket81Layer() *Packet81Layer {
	return &Packet81Layer{}
}

func (thisBitstream *extendedReader) DecodePacket81Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket81Layer()

	var err error

	layer.StreamJob, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.FilteringEnabled, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.AllowThirdPartySales, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.CharacterAutoSpawn, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	stringLen, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.ReferentString, err = thisBitstream.readASCII(int(stringLen))
	if err != nil {
		return layer, err
	}
	// This assignment is justifiable because a call to readJoinObject() below depends on it
	reader.Context().InstanceTopScope = layer.ReferentString
	if !reader.Context().IsStudio {
		layer.Int1, err = thisBitstream.readUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Int2, err = thisBitstream.readUint32BE()
		if err != nil {
			return layer, err
		}

		reader.Context().Int1 = layer.Int1
		reader.Context().Int2 = layer.Int2
	}

	arrayLen, err := thisBitstream.readUintUTF8()
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
		referent, err := thisBitstream.readJoinObject(context)
		if err != nil {
			return layer, err
		}

		classID, err := thisBitstream.readUint16BE()
		if err != nil {
			return layer, err
		}

		if int(classID) > len(context.StaticSchema.Instances) {
			return layer, fmt.Errorf("class idx %d is higher than %d", classID, len(context.StaticSchema.Instances))
		}

		schema := context.StaticSchema.Instances[classID]
		thisItem.Schema = schema
		instance, err := context.InstancesByReferent.CreateInstance(referent)
		if err != nil {
			return layer, err
		}
		instance.ClassName = schema.Name
		instance.IsService = true
		thisItem.Instance = instance

		thisItem.Bool1, err = thisBitstream.readBoolByte()
		if err != nil {
			return layer, err
		}
		thisItem.Bool2, err = thisBitstream.readBoolByte()
		if err != nil {
			return layer, err
		}
		layer.Items[i] = thisItem
	}
	return layer, nil
}

func (layer *Packet81Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x81)
	if err != nil {
		return err
	}

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

	err = stream.writeUint32BE(uint32(len(layer.ReferentString)))
	if err != nil {
		return err
	}
	err = stream.writeASCII(layer.ReferentString)
	if err != nil {
		return err
	}

	if !writer.Context().IsStudio {
		err = stream.writeUint32BE(layer.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(layer.Int2)
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
		err = stream.writeBoolByte(item.Bool1)
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(item.Bool2)
		if err != nil {
			return err
		}
	}
	return nil
}

func (layer *Packet81Layer) String() string {
	return fmt.Sprintf("ID_SET_GLOBALS: %d services", len(layer.Items))
}

func (Packet81Layer) TypeString() string {
	return "ID_SET_GLOBALS"
}
