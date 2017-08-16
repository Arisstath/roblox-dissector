package main
import "github.com/google/gopacket"

type Packet81LayerItem struct {
	ClassID uint16
	Object1 Object
	Bool1 bool
	Bool2 bool
}

type Packet81Layer struct {
	Bools [5]bool
	String1 []byte
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

	context.WaitForSchema()
	defer context.FinishSchema()

	if !context.UseStaticSchema {
		len, err := thisBitstream.ReadUint8()
		if err != nil {
			return layer, err
		}
		var j uint8
		layer.Items = make([]*Packet81LayerItem, len)
		for j = 0; j < len; j++ {
			thisItem := &Packet81LayerItem{}
			len9Value, err := thisBitstream.Bits(9)
			if err != nil {
				return layer, err
			}
			thisItem.ClassID = uint16(len9Value)
			thisItem.Object1, err = thisBitstream.ReadObject(false, context)
			if err != nil {
				return layer, err
			}
			serialized := formatBindable(thisItem.Object1)
			context.Rebindables[serialized] = struct{}{}
			println("REGISTERED REBIND: ", serialized)

			layer.Items[j] = thisItem
		}
		return layer, err
	} else {
		arrayLen, err := thisBitstream.ReadUintUTF8()
		if err != nil {
			return layer, err
		}
		println("Will read array of", arrayLen)

		layer.Items = make([]*Packet81LayerItem, arrayLen)
		for i := 0; i < int(arrayLen); i++ {
			thisItem := &Packet81LayerItem{}
			thisItem.Object1, err = thisBitstream.ReadObject(false, context)
			if err != nil {
				return layer, err
			}
			serialized := formatBindable(thisItem.Object1)
			context.Rebindables[serialized] = struct{}{}
			println("REGISTERED REBIND: ", serialized)

			thisItem.ClassID, err = thisBitstream.ReadUint16BE()
			if err != nil {
				return layer, err
			}

			thisItem.Bool1, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}
			thisItem.Bool2, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}
		}
		return layer, nil
	}
}
