package packets

import (
	"errors"
	"fmt"

	"github.com/gskartwii/rbxfile"
)
import "github.com/gskartwii/roblox-dissector/util"

// Describes a global service from ID_SET_GLOBALS (TopReplication)
type TopReplicationItem struct {
	// Class ID, according to ID_NEW_SCHEMA (SchemaPacket)
	ClassID  uint16
	Instance *rbxfile.Instance
	Bool1    bool
	Bool2    bool
}

// ID_SET_GLOBALS - server -> client
type TopReplication struct {
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
	Items []*TopReplicationItem
}

func NewTopReplication() *TopReplication {
	return &TopReplication{}
}

func (thisBitstream *PacketReaderBitstream) DecodeTopReplication(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewTopReplication()
	
	var err error

	layer.StreamJob, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	layer.FilteringEnabled, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	layer.AllowThirdPartySales, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	layer.CharacterAutoSpawn, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	stringLen, err := thisBitstream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	layer.ReferentString, err = thisBitstream.ReadASCII(int(stringLen))
	if err != nil {
		return layer, err
	}
	reader.Context().InstanceTopScope = layer.ReferentString
	if !reader.Context().IsStudio {
		layer.Int1, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return layer, err
		}

		reader.Context().Int1 = layer.Int1
		reader.Context().Int2 = layer.Int2
	}

	arrayLen, err := thisBitstream.ReadUintUTF8()
	if err != nil {
		return layer, err
	}
	if arrayLen > 0x1000 {
		return layer, errors.New("sanity check: exceeded maximum preschema len")
	}

	context := reader.Context()
	context.DataModel = &rbxfile.Root{make([]*rbxfile.Instance, arrayLen)}

	layer.Items = make([]*TopReplicationItem, arrayLen)
	for i := 0; i < int(arrayLen); i++ {
		thisItem := &TopReplicationItem{}
		referent, err := thisBitstream.ReadJoinObject(context)
		if err != nil {
			return layer, err
		}

		thisItem.ClassID, err = thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}

		if int(thisItem.ClassID) > len(context.StaticSchema.Instances) {
			return layer, fmt.Errorf("class idx %d is higher than %d", thisItem.ClassID, len(context.StaticSchema.Instances))
		}

		className := context.StaticSchema.Instances[thisItem.ClassID].Name
		thisService := &rbxfile.Instance{
			ClassName:  className,
			Reference:  string(referent),
			Properties: make(map[string]rbxfile.Value, 0),
			IsService:  true,
		}
		context.DataModel.Instances[i] = thisService
		context.InstancesByReferent.AddInstance(referent, thisService)
		thisItem.Instance = thisService

		thisItem.Bool1, err = thisBitstream.ReadBoolByte()
		if err != nil {
			return layer, err
		}
		thisItem.Bool2, err = thisBitstream.ReadBoolByte()
		if err != nil {
			return layer, err
		}
		layer.Items[i] = thisItem
	}
	return layer, nil
}

func (layer *TopReplication) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x81)
	if err != nil {
		return err
	}

	err = stream.WriteBool(layer.StreamJob)
	if err != nil {
		return err
	}
	err = stream.WriteBool(layer.FilteringEnabled)
	if err != nil {
		return err
	}
	err = stream.WriteBool(layer.AllowThirdPartySales)
	if err != nil {
		return err
	}
	err = stream.WriteBool(layer.CharacterAutoSpawn)
	if err != nil {
		return err
	}

	err = stream.WriteUint32BE(uint32(len(layer.ReferentString)))
	if err != nil {
		return err
	}
	err = stream.WriteASCII(layer.ReferentString)
	if err != nil {
		return err
	}

	// FIXME: assumes Studio

	err = stream.WriteUintUTF8(uint32(len(layer.Items)))
	for _, item := range layer.Items {
		err = stream.WriteJoinObject(item.Instance, writer.Context())
		if err != nil {
			return err
		}
		err = stream.WriteUint16BE(item.ClassID)
		if err != nil {
			return err
		}
		err = stream.WriteBool(item.Bool1)
		if err != nil {
			return err
		}
		err = stream.WriteBool(item.Bool2)
		if err != nil {
			return err
		}
	}
	return nil
}
