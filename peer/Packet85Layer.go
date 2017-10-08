package peer
import "github.com/gskartwii/rbxfile"

type Packet85LayerSubpacket struct {
	Instance *rbxfile.Instance
	UnknownInt uint8
	CFrame rbxfile.ValueCFrame
	Pos1 rbxfile.ValueVector3
	Pos2 rbxfile.ValueVector3
	Motors []PhysicsMotor
	Children []Packet85LayerSubpacket
}

type Packet85Layer struct {
	SubPackets []*Packet85LayerSubpacket
}

func NewPacket85Layer() *Packet85Layer {
	return &Packet85Layer{}
}

func DecodePacket85Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	thisBitstream := packet.Stream

	layer := NewPacket85Layer()
	for {
		subpacket := &Packet85LayerSubpacket{}
		referent, err := thisBitstream.ReadObject(false, context)
		if err != nil {
			return layer, err
		}
		if referent == "null" {
			break
		}
		subpacket.Instance = context.InstancesByReferent.TryGetInstance(referent)

		hasInt, err := thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		if hasInt {
			subpacket.UnknownInt, err = thisBitstream.ReadByte()
			if err != nil {
				return layer, err
			}
		}
		subpacket.CFrame, err = thisBitstream.ReadPhysicsCFrame()
		if err != nil {
			return layer, err
		}
		subpacket.Pos1, err = thisBitstream.ReadCoordsMode1()
		if err != nil {
			return layer, err
		}
		subpacket.Pos2, err = thisBitstream.ReadCoordsMode1()
		if err != nil {
			return layer, err
		}
		subpacket.Motors, err = thisBitstream.ReadMotors()
		if err != nil {
			return layer, err
		}

		isSolo, err := thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		if !isSolo {
			for {
				child := Packet85LayerSubpacket{}
				referent, err := thisBitstream.ReadObject(false, context)
				if err != nil {
					return layer, err
				}
				child.Instance = context.InstancesByReferent.TryGetInstance(referent)
				child.CFrame, err = thisBitstream.ReadPhysicsCFrame()
				if err != nil {
					return layer, err
				}
				child.Pos1, err = thisBitstream.ReadCoordsMode1()
				if err != nil {
					return layer, err
				}
				child.Pos2, err = thisBitstream.ReadCoordsMode1()
				if err != nil {
					return layer, err
				}
				child.Motors, err = thisBitstream.ReadMotors()
				if err != nil {
					return layer, err
				}
				isEOF, err := thisBitstream.ReadBool()
				if isEOF {
					break
				}
				subpacket.Children = append(subpacket.Children, child)
			}
		}
		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (layer *Packet85Layer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		err := stream.WriteObject(subpacket.Instance, false, context)
		if err != nil {
			return err
		}
		err = stream.WriteBool(subpacket.UnknownInt != 0)
		if err != nil {
			return err
		}
		if subpacket.UnknownInt != 0 {
			err = stream.WriteByte(subpacket.UnknownInt)
			if err != nil {
				return err
			}
		}

		err = stream.WritePhysicsCFrame(subpacket.CFrame)
		if err != nil {
			return err
		}
		err = stream.WriteCoordsMode1(subpacket.Pos1)
		if err != nil {
			return err
		}
		err = stream.WriteCoordsMode1(subpacket.Pos2)
		if err != nil {
			return err
		}
		err = stream.WriteMotors(subpacket.Motors)
		if err != nil {
			return err
		}

		for j := 0; j < len(subpacket.Children); j++ {
			child := subpacket.Children[j]
			err = stream.WriteBool(true)
			if err != nil {
				return err
			}
			err = stream.WriteObject(child.Instance, false, context)
			if err != nil {
				return err
			}

			err = stream.WritePhysicsCFrame(child.CFrame)
			if err != nil {
				return err
			}
			err = stream.WriteCoordsMode1(child.Pos1)
			if err != nil {
				return err
			}
			err = stream.WriteCoordsMode1(child.Pos2)
			if err != nil {
				return err
			}
			err = stream.WriteMotors(child.Motors)
			if err != nil {
				return err
			}
		}
		err = stream.WriteBool(false) // Terminator for children
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
