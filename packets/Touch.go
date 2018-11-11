package packets

import (
	"errors"

	"github.com/gskartwii/rbxfile"
)

// Touch replication for a single touch
type TouchSubpacket struct {
	Instance1 *rbxfile.Instance
	Instance2 *rbxfile.Instance
	// Touch started? If false, ended.
	IsTouch bool
}

// ID_TOUCHES - client <-> server
type Touch struct {
	SubPackets []*TouchSubpacket
}

func NewTouch() *Touch {
	return &Touch{}
}

func (thisBitstream *PacketReaderBitstream) DecodeTouch(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	

	layer := NewTouch()
	context := reader.Context()
	for {
		subpacket := &TouchSubpacket{}
		referent, err := thisBitstream.readObject(reader.Caches())
		// hopefully we don't need to check for CacheReadOOB here
		if err != nil {
			return layer, err
		}
		if referent.IsNull() {
			break
		}
		context.InstancesByReferent.OnAddInstance(referent, func(inst *rbxfile.Instance) {
			subpacket.Instance1 = inst
		})

		referent, err = thisBitstream.readObject(reader.Caches())
		if err != nil {
			return layer, err
		}
		if referent.IsNull() {
			return layer, errors.New("NULL second touch referent!")
		}
		context.InstancesByReferent.OnAddInstance(referent, func(inst *rbxfile.Instance) {
			subpacket.Instance2 = inst
		})

		subpacket.IsTouch, err = thisBitstream.readBoolByte()
		if err != nil {
			return layer, err
		}

		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (layer *Touch) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	err := stream.WriteByte(0x86)
	if err != nil {
		return err
	}
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		if subpacket.Instance1 == nil || subpacket.Instance2 == nil {
			println("WARNING: 0x86 skipping serialize because instances don't exist yet!")
			continue
		}
		err = stream.writeObject(subpacket.Instance1, writer.Caches())
		if err != nil {
			return err
		}
		err = stream.writeObject(subpacket.Instance2, writer.Caches())
		if err != nil {
			return err
		}
		err = stream.writeBoolByte(subpacket.IsTouch)
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
