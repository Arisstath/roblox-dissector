package main
import "github.com/google/gopacket"

type Packet81LayerItem struct {
	Int1 uint16
	String1 []byte
	Int2 uint32
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
		thisItem.Int1 = uint16(len9Value)

		cacheIndex, err := thisBitstream.ReadUint8()
		if err != nil {
			return layer, err
		}
		if cacheIndex < 0x80 {
			thisItem.String1 = context.ReplicatorStringCache[cacheIndex]
		} else {
			stringLen, err := thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			thisItem.String1, err = thisBitstream.ReadString(int(stringLen))
			if err != nil {
				return layer, err
			}
			context.ReplicatorStringCache[cacheIndex - 0x80] = thisItem.String1
		}

		thisItem.Int2, err = thisBitstream.ReadUint32LE()
		if err != nil {
			return layer, err
		}

		layer.Items[j] = thisItem
	}

	return layer, err
}
