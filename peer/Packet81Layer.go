package peer
import "errors"
import "github.com/gskartwii/rbxfile"
import "fmt"

// Describes a global service from ID_SET_GLOBALS (Packet81Layer)
type Packet81LayerItem struct {
	// Class ID, according to ID_NEW_SCHEMA (Packet97Layer)
	ClassID uint16
	Instance *rbxfile.Instance
	Bool1 bool
	Bool2 bool
}

// ID_SET_GLOBALS - server -> client
type Packet81Layer struct {
	// Use NetworkOwner?
	DistributedPhysicsEnabled bool
	// Is streaming enabled?
	StreamJob bool
	// Is Filtering enabled?
	FilteringEnabled bool
	AllowThirdPartySales bool
	CharacterAutoSpawn bool
	// Server's scope
	ReferentString []byte
	Int1 uint32
	Int2 uint32
	// List of services to be set
	Items []*Packet81LayerItem
}

func NewPacket81Layer() *Packet81Layer {
	return &Packet81Layer{}
}

func decodePacket81Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket81Layer()
	thisBitstream := packet.stream
	var err error

	layer.DistributedPhysicsEnabled, err = thisBitstream.readBool()
	if err != nil {
		return layer, err
	}
	layer.StreamJob, err = thisBitstream.readBool()
	if err != nil {
		return layer, err
	}
	layer.FilteringEnabled, err = thisBitstream.readBool()
	if err != nil {
		return layer, err
	}
	layer.AllowThirdPartySales, err = thisBitstream.readBool()
	if err != nil {
		return layer, err
	}
	layer.CharacterAutoSpawn, err = thisBitstream.readBool()
	if err != nil {
		return layer, err
	}
	stringLen, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.ReferentString, err = thisBitstream.readString(int(stringLen))
	if err != nil {
		return layer, err
	}
	context.InstanceTopScope = string(layer.ReferentString)
	if !context.IsStudio {
		layer.Int1, err = thisBitstream.readUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Int2, err = thisBitstream.readUint32BE()
		if err != nil {
			return layer, err
		}
	}

    arrayLen, err := thisBitstream.readUintUTF8()
    if err != nil {
        return layer, err
    }
    if arrayLen > 0x1000 {
        return layer, errors.New("sanity check: exceeded maximum preschema len")
    }

    context.DataModel = &rbxfile.Root{make([]*rbxfile.Instance, arrayLen)}

    layer.Items = make([]*Packet81LayerItem, arrayLen)
    for i := 0; i < int(arrayLen); i++ {
        thisItem := &Packet81LayerItem{}
        referent, err := thisBitstream.readObject(context.IsClient(packet.Source), true, context)
        if err != nil {
            return layer, err
        }

        thisItem.ClassID, err = thisBitstream.readUint16BE()
        if err != nil {
            return layer, err
        }

        if int(thisItem.ClassID) > len(context.StaticSchema.Instances) {
            return layer, errors.New(fmt.Sprintf("class idx %d is higher than %d", thisItem.ClassID, len(context.StaticSchema.Instances)))
        }

        className := context.StaticSchema.Instances[thisItem.ClassID].Name
        thisService := &rbxfile.Instance{
            ClassName: className,
            Reference: string(referent),
            Properties: make(map[string]rbxfile.Value, 0),
			IsService: true,
        }
        context.DataModel.Instances[i] = thisService
        context.InstancesByReferent.AddInstance(referent, thisService)
        thisItem.Instance = thisService

        thisItem.Bool1, err = thisBitstream.readBool()
        if err != nil {
            return layer, err
        }
        thisItem.Bool2, err = thisBitstream.readBool()
        if err != nil {
            return layer, err
        }
        layer.Items[i] = thisItem
    }
    return layer, nil
}

func (layer *Packet81Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
    var err error
    err = stream.WriteByte(0x81)
    if err != nil {
        return err
    }

	err = stream.writeBool(layer.DistributedPhysicsEnabled)
    if err != nil {
        return err
    }
	err = stream.writeBool(layer.StreamJob)
    if err != nil {
        return err
    }
	err = stream.writeBool(layer.FilteringEnabled)
    if err != nil {
        return err
    }
	err = stream.writeBool(layer.AllowThirdPartySales)
    if err != nil {
        return err
    }
	err = stream.writeBool(layer.CharacterAutoSpawn)
    if err != nil {
        return err
    }

    err = stream.writeUint32BE(uint32(len(layer.ReferentString)))
    if err != nil {
        return err
    }
    err = stream.allBytes(layer.ReferentString)
    if err != nil {
        return err
    }

    // FIXME: assumes Studio

    err = stream.writeUintUTF8(uint32(len(layer.Items)))
    for _, item := range layer.Items {
        err = stream.writeObject(isClient, item.Instance, true, context)
        if err != nil {
            return err
        }
        err = stream.writeUint16BE(item.ClassID)
        if err != nil {
            return err
        }
        err = stream.writeBool(item.Bool1)
        if err != nil {
            return err
        }
        err = stream.writeBool(item.Bool2)
        if err != nil {
            return err
        }
    }
    return nil
}
