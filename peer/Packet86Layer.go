package peer
import "github.com/gskartwii/rbxfile"

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

func decodePacket86Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	thisBitstream := packet.stream
	isClient := context.IsClient(packet.Source)

	layer := NewPacket86Layer()
	for {
		subpacket := &Packet86LayerSubpacket{}
		referent, err := thisBitstream.readObject(isClient, false, context)
		if err != nil {
			return layer, err
		}
		if referent == "null" {
			break
		}
		subpacket.Instance1 = context.InstancesByReferent.TryGetInstance(referent)
		referent, err = thisBitstream.readObject(isClient, false, context)
		if err != nil {
			return layer, err
		}
		subpacket.Instance2 = context.InstancesByReferent.TryGetInstance(referent)
		subpacket.IsTouch, err = thisBitstream.readBool()
		if err != nil {
			return layer, err
		}

		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (layer *Packet86Layer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		err := stream.writeObject(isClient, subpacket.Instance1, false, context)
		if err != nil {
			return err
		}
		err = stream.writeObject(isClient, subpacket.Instance2, false, context)
		if err != nil {
			return err
		}
		err = stream.writeBool(subpacket.IsTouch)
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
