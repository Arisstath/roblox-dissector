package peer

import (
	"errors"
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
)

// Touch replication for a single touch
type Packet86LayerSubpacket struct {
	Instance1 *datamodel.Instance
	Instance2 *datamodel.Instance
	// Touch started? If false, ended.
	IsTouch bool
}

// ID_TOUCHES - client <-> server
type Packet86Layer struct {
	SubPackets []*Packet86LayerSubpacket
}

func NewPacket86Layer() *Packet86Layer {
	return &Packet86Layer{}
}

func (thisStream *extendedReader) DecodePacket86Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {

	layer := NewPacket86Layer()
	context := reader.Context()
	for {
		subpacket := &Packet86LayerSubpacket{}
		referent, err := thisStream.readObject(reader.Caches())
		// hopefully we don't need to check for CacheReadOOB here
		if err != nil {
			return layer, err
		}
		if referent.IsNull {
			break
		}
		context.InstancesByReferent.OnAddInstance(referent, func(inst *datamodel.Instance) {
			subpacket.Instance1 = inst
		})

		referent, err = thisStream.readObject(reader.Caches())
		if err != nil {
			return layer, err
		}
		if referent.IsNull {
			return layer, errors.New("NULL second touch referent!")
		}
		context.InstancesByReferent.OnAddInstance(referent, func(inst *datamodel.Instance) {
			subpacket.Instance2 = inst
		})

		subpacket.IsTouch, err = thisStream.readBoolByte()
		if err != nil {
			return layer, err
		}

		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (layer *Packet86Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
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

func (layer *Packet86Layer) String() string {
	return fmt.Sprintf("ID_TOUCH: %d items", len(layer.SubPackets))
}

func (Packet86Layer) TypeString() string {
	return "ID_TOUCH"
}
