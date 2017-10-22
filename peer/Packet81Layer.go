package peer
import "errors"
import "github.com/gskartwii/rbxfile"
import "fmt"

type Packet81LayerItem struct {
	ClassID uint16
	Instance *rbxfile.Instance
	Bool1 bool
	Bool2 bool
}

type Packet81Layer struct {
	Bools [5]bool
	ReferentString []byte
	Int1 uint32
	Int2 uint32
	Items []*Packet81LayerItem
}

func NewPacket81Layer() *Packet81Layer {
	return &Packet81Layer{}
}

func DecodePacket81Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket81Layer()
	thisBitstream := packet.Stream
	var err error

	for i := 0; i < 5; i++ {
		layer.Bools[i], err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}
	}
	stringLen, err := thisBitstream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	layer.ReferentString, err = thisBitstream.ReadString(int(stringLen))
	if err != nil {
		return layer, err
	}
	context.InstanceTopScope = string(layer.ReferentString)
	if !context.IsStudio {
		layer.Int1, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
	}

	context.WaitForSchema()
	defer context.FinishSchema()
    arrayLen, err := thisBitstream.ReadUintUTF8()
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
        referent, err := thisBitstream.ReadObject(context.IsClient(packet.Source), true, context)
        if err != nil {
            return layer, err
        }

        thisItem.ClassID, err = thisBitstream.ReadUint16BE()
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

        thisItem.Bool1, err = thisBitstream.ReadBool()
        if err != nil {
            return layer, err
        }
        thisItem.Bool2, err = thisBitstream.ReadBool()
        if err != nil {
            return layer, err
        }
        layer.Items[i] = thisItem
    }
    return layer, nil
}

func (layer *Packet81Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
    var err error
    err = stream.WriteByte(0x81)
    if err != nil {
        return err
    }

    for i := 0; i < 5; i++ {
        err = stream.WriteBool(layer.Bools[i])
        if err != nil {
            return err
        }
    }
    err = stream.WriteUint32BE(uint32(len(layer.ReferentString)))
    if err != nil {
        return err
    }
    err = stream.AllBytes(layer.ReferentString)
    if err != nil {
        return err
    }

    // FIXME: assumes Studio

    err = stream.WriteUintUTF8(len(layer.Items))
    for _, item := range layer.Items {
        err = stream.WriteObject(isClient, item.Instance, true, context)
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
