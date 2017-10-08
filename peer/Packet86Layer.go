package peer
import "github.com/gskartwii/rbxfile"

type Packet86LayerSubpacket struct {
	Instance1 *rbxfile.Instance
	Instance2 *rbxfile.Instance
	IsTouch bool
}

type Packet86Layer struct {
	SubPackets []*Packet86LayerSubpacket
}

func NewPacket86Layer() *Packet86Layer {
	return &Packet86Layer{}
}

func DecodePacket86Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	thisBitstream := packet.Stream

	layer := NewPacket86Layer()
	for {
		subpacket := &Packet86LayerSubpacket{}
		referent, err := thisBitstream.ReadObject(false, context)
		if err != nil {
			return layer, err
		}
		if referent == "null" {
			break
		}
		subpacket.Instance1 = context.InstancesByReferent.TryGetInstance(referent)
		referent, err = thisBitstream.ReadObject(false, context)
		if err != nil {
			return layer, err
		}
		subpacket.Instance2 = context.InstancesByReferent.TryGetInstance(referent)
		subpacket.IsTouch, err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}

		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (layer *Packet86Layer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		err := stream.WriteObject(subpacket.Instance1, false, context)
		if err != nil {
			return err
		}
		err = stream.WriteObject(subpacket.Instance2, false, context)
		if err != nil {
			return err
		}
		err = stream.WriteBool(subpacket.IsTouch)
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
