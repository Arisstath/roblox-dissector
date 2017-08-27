package main
import "github.com/google/gopacket"
import "errors"
import "github.com/gskartwii/rbxfile"

type Packet81LayerItem struct {
	ClassID uint16
	Instance *rbxfile.Instance
	Bool1 bool
	Bool2 bool
}

type Packet81Layer struct {
	Bools [5]bool
	String1 []byte
	Int1 uint32
	Int2 uint32
	Items []*Packet81LayerItem
}

func NewPacket81Layer() Packet81Layer {
	return Packet81Layer{}
}

func DecodePacket81Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket81Layer()
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
	layer.String1, err = thisBitstream.ReadString(int(stringLen))
	if err != nil {
		return layer, err
	}
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
        referent, err := thisBitstream.ReadObject(true, context)
        if err != nil {
            return layer, err
        }

        thisItem.ClassID, err = thisBitstream.ReadUint16BE()
        if err != nil {
            return layer, err
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
