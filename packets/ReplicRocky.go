package packets

import "errors"
import "github.com/gskartwii/roblox-dissector/util"

type ReplicRockySubpacket interface{}

type ReplicRocky struct {
	Subpacket     ReplicRockySubpacket
	SubpacketType uint8
}

type ReplicRocky_00 struct {
	Values [5]uint32
}

type ReplicRocky_01 struct {
	Int1 uint8
	Int2 uint32
	Int3 uint32
	Int4 uint32
	Int5 uint64
}

type ReplicRocky_05 struct {
	Int uint32
}

type ReplicRocky_06 struct {
	Int1 uint32
	Int2 uint32
}

type ReplicRocky_07 struct{}

func (thisBitstream *PacketReaderBitstream) DecodeReplicRocky(reader util.PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	inner := &ReplicRocky{}

	inner.SubpacketType, err = thisBitstream.ReadUint8()
	if err != nil {
		return inner, err
	}
	var subpacket interface{}
	switch inner.SubpacketType {
	case 0: // Rocky
		thisSubpacket := &ReplicRocky_00{}
		for i := 0; i < 5; i++ {
			thisSubpacket.Values[i], err = thisBitstream.ReadUint32BE()
			if err != nil {
				return inner, err
			}
		}
		subpacket = thisSubpacket
	case 1:
		thisSubpacket := &ReplicRocky_01{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int3, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int4, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int5, err = thisBitstream.ReadUint64BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 2: // net pmc response

	case 5:
		thisSubpacket := &ReplicRocky_05{}
		thisSubpacket.Int, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 6: // id response
		thisSubpacket := &ReplicRocky_06{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 7:
		thisSubpacket := &ReplicRocky_07{}
		subpacket = thisSubpacket
	default:
		layers.Root.Logger.Println("don't know rocky subpacket", inner.Type)
		return inner, errors.New("unimplemented subpacket type")
	}
	inner.Subpacket = subpacket

	return inner, err
}

func (layer *ReplicRocky) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteByte(layer.SubpacketType)
	if err != nil {
		return err
	}

	switch layer.SubpacketType {
	case 6:
		subpacket := layer.Subpacket.(*ReplicRocky_06)
		err = stream.WriteUint32BE(subpacket.Int1)
		if err != nil {
			return err
		}
		err = stream.WriteUint32BE(subpacket.Int2)
		break
	default:
		println("Tried to write rocky packet", layer.Type)
		return errors.New("rocky packet not implemented!")
	}

	return err
}

func (ReplicRocky) Type() uint8 {
	return 9
}
func (ReplicRocky) TypeString() string {
	return "ID_REPLIC_ROCKY"
}
