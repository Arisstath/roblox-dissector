package peer
import "github.com/gskartwii/rbxfile"

// History waypoint in the movement of a mechanism
type PhysicsHistoryWaypoint struct {
	// Position at that point
	Position rbxfile.ValueVector3
	// Level of precision. Smaller = higher precision
	PrecisionLevel uint8
	// Interval to previous waypoint in ms
	Interval uint8
}

// Physics replication for one instance
type Packet85LayerSubpacket struct {
	Instance *rbxfile.Instance
	// See http://wiki.roblox.com/index.php?title=API:Enum/HumanoidStateType
	NetworkHumanoidState uint8
	CFrame rbxfile.ValueCFrame
	LinearVelocity rbxfile.ValueVector3
	RotationalVelocity rbxfile.ValueVector3
	// CFrames for any motors attached
	Motors []PhysicsMotor
	// Any other parts attached to this mechanism
	Children []Packet85LayerSubpacket
	// Movement history
	HistoryWaypoints []PhysicsHistoryWaypoint
}

// ID_PHYSICS - client <-> server
type Packet85Layer struct {
	SubPackets []*Packet85LayerSubpacket
}

func NewPacket85Layer() *Packet85Layer {
	return &Packet85Layer{}
}

func decodePacket85Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	thisBitstream := packet.stream

	layer := NewPacket85Layer()
	for {
		subpacket := &Packet85LayerSubpacket{}
		referent, err := thisBitstream.readObject(context.IsClient(packet.Source), false, context)
		if err != nil {
			return layer, err
		}
		if referent == "null" {
			break
		}
		subpacket.Instance = context.InstancesByReferent.TryGetInstance(referent)

		hasState, err := thisBitstream.readBool()
		if err != nil {
			return layer, err
		}
		if hasState {
			subpacket.NetworkHumanoidState, err = thisBitstream.ReadByte()
			if err != nil {
				return layer, err
			}
		}
		
		if !context.IsClient(packet.Source) { // packet came from server: must read movement history
			println("from server")
			hasHistory, err := thisBitstream.readBool()
			if err != nil {
				return layer, err
			}
			if !hasHistory {
				println("no history")
				subpacket.Motors, err = thisBitstream.readMotors()
				if err != nil {
					return layer, err
				}
			} else {
				println("has history")
				xPacketCompression, err := thisBitstream.readBool()
				if err != nil {
					return layer, err
				}

				if !xPacketCompression {
					println("use compr")
					subpacket.CFrame, err = thisBitstream.readPhysicsCFrame()
					if err != nil {
						return layer, err
					}
					println("cf", subpacket.CFrame.String())
					subpacket.LinearVelocity, err = thisBitstream.readCoordsMode1()
					if err != nil {
						return layer, err
					}
					println("lv", subpacket.LinearVelocity.String())
					subpacket.RotationalVelocity, err = thisBitstream.readCoordsMode1()
					if err != nil {
						return layer, err
					}
					println("rv", subpacket.RotationalVelocity.String())
				} else {
					matrix, err := thisBitstream.readPhysicsMatrix()
					if err != nil {
						return layer, err
					}
					subpacket.CFrame = rbxfile.ValueCFrame{rbxfile.ValueVector3{}, matrix}
					println("cf", subpacket.CFrame.String())
				}
				subpacket.Motors, err = thisBitstream.readMotors()
				if err != nil {
					return layer, err
				}
				println("motor succ")

				numNodes, err := thisBitstream.readUintUTF8()
				if err != nil {
					return layer, err
				}
				println("numnod", numNodes)
				
				currentPosition := subpacket.CFrame.Position

				subpacket.HistoryWaypoints = make([]PhysicsHistoryWaypoint, numNodes)
				for i := 0; i < int(numNodes); i++ {
					subpacket.HistoryWaypoints[i].PrecisionLevel, err = thisBitstream.ReadByte()
					if err != nil {
						return layer, err
					}
					deltaX, err := thisBitstream.ReadByte()
					if err != nil {
						return layer, err
					}
					deltaY, err := thisBitstream.ReadByte()
					if err != nil {
						return layer, err
					}
					deltaZ, err := thisBitstream.ReadByte()
					if err != nil {
						return layer, err
					}
					subpacket.HistoryWaypoints[i].Interval, err = thisBitstream.ReadByte()
					if err != nil {
						return layer, err
					}
					currentPosition.X -= float32(int8(deltaX) * int8(subpacket.HistoryWaypoints[i].PrecisionLevel + 1)) * 0.01
					currentPosition.Y -= float32(int8(deltaY) * int8(subpacket.HistoryWaypoints[i].PrecisionLevel + 1)) * 0.01
					currentPosition.Z -= float32(int8(deltaZ) * int8(subpacket.HistoryWaypoints[i].PrecisionLevel + 1)) * 0.01
					subpacket.HistoryWaypoints[i].Position = currentPosition
				}
			} // hasHistory
		} else {
			// packet came from client, just read the assembly
			subpacket.CFrame, err = thisBitstream.readPhysicsCFrame()
			if err != nil {
				return layer, err
			}
			subpacket.LinearVelocity, err = thisBitstream.readCoordsMode1()
			if err != nil {
				return layer, err
			}
			subpacket.RotationalVelocity, err = thisBitstream.readCoordsMode1()
			if err != nil {
				return layer, err
			}
			subpacket.Motors, err = thisBitstream.readMotors()
			if err != nil {
				return layer, err
			}
		}

		isSolo, err := thisBitstream.readBool()
		if err != nil {
			return layer, err
		}
		if !isSolo {
			for {
				child := Packet85LayerSubpacket{}
				referent, err := thisBitstream.readObject(context.IsClient(packet.Source), false, context)
				if err != nil {
					return layer, err
				}
				child.Instance = context.InstancesByReferent.TryGetInstance(referent)
				child.CFrame, err = thisBitstream.readPhysicsCFrame()
				if err != nil {
					return layer, err
				}
				child.LinearVelocity, err = thisBitstream.readCoordsMode1()
				if err != nil {
					return layer, err
				}
				child.RotationalVelocity, err = thisBitstream.readCoordsMode1()
				if err != nil {
					return layer, err
				}
				child.Motors, err = thisBitstream.readMotors()
				if err != nil {
					return layer, err
				}
				isEOF, err := thisBitstream.readBool()
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

func (layer *Packet85Layer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		err := stream.writeObject(isClient, subpacket.Instance, false, context)
		if err != nil {
			return err
		}
		err = stream.writeBool(subpacket.NetworkHumanoidState != 0)
		if err != nil {
			return err
		}
		if subpacket.NetworkHumanoidState != 0 {
			err = stream.WriteByte(subpacket.NetworkHumanoidState)
			if err != nil {
				return err
			}
		}

		err = stream.writePhysicsCFrame(subpacket.CFrame)
		if err != nil {
			return err
		}
		err = stream.writeCoordsMode1(subpacket.LinearVelocity)
		if err != nil {
			return err
		}
		err = stream.writeCoordsMode1(subpacket.RotationalVelocity)
		if err != nil {
			return err
		}
		err = stream.writeMotors(subpacket.Motors)
		if err != nil {
			return err
		}

		for j := 0; j < len(subpacket.Children); j++ {
			child := subpacket.Children[j]
			err = stream.writeBool(true)
			if err != nil {
				return err
			}
			err = stream.writeObject(isClient, child.Instance, false, context)
			if err != nil {
				return err
			}

			err = stream.writePhysicsCFrame(child.CFrame)
			if err != nil {
				return err
			}
			err = stream.writeCoordsMode1(child.LinearVelocity)
			if err != nil {
				return err
			}
			err = stream.writeCoordsMode1(child.RotationalVelocity)
			if err != nil {
				return err
			}
			err = stream.writeMotors(child.Motors)
			if err != nil {
				return err
			}
		}
		err = stream.writeBool(false) // Terminator for children
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
