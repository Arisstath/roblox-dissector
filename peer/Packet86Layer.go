package peer

import (
	"errors"

	"github.com/gskartwii/rbxfile"
)

// Touch replication for a single touch
type Packet86LayerSubpacket struct {
	Instance1 *rbxfile.Instance
	Instance2 *rbxfile.Instance
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

func DecodePacket86Layer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	thisBitstream := packet.stream

	layer := NewPacket86Layer()
	context := reader.Context()
	for {
		subpacket := &Packet86LayerSubpacket{}
		referent, err := thisBitstream.readObject(reader.Caches())
		if err != nil {
			return layer, err
		}
		if referent.IsNull() {
			break
		}
		subpacket.Instance1, err = context.InstancesByReferent.TryGetInstance(referent)
		if err != nil {
			return layer, err
		}

		referent, err = thisBitstream.readObject(reader.Caches())
		if err != nil {
			return layer, err
		}
		if referent.IsNull() {
			return layer, errors.New("NULL second touch referent!")
		}
		subpacket.Instance2, err = context.InstancesByReferent.TryGetInstance(referent)
		if err != nil {
			return layer, err
		}

		subpacket.IsTouch, err = thisBitstream.readBoolByte()
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
