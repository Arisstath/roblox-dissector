package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

// Packet85LayerSubpacket represents physics replication for one instance
type Packet85LayerSubpacket struct {
	Data PhysicsData
	// See http://wiki.roblox.com/index.php?title=API:Enum/HumanoidStateType
	NetworkHumanoidState uint8
	// CFrames for any motors attached
	// Any other parts attached to this mechanism
	Children []*PhysicsData
	History  []*PhysicsData
}

// String() implements fmt.Stringer
func (packet *Packet85LayerSubpacket) String() string {
	return "Physics: " + packet.Data.Instance.GetFullName()
}

// PhysicsData represents generic physics data
type PhysicsData struct {
	Instance           *datamodel.Instance
	CFrame             rbxfile.ValueCFrame
	LinearVelocity     rbxfile.ValueVector3
	RotationalVelocity rbxfile.ValueVector3
	Motors             []PhysicsMotor
	Interval           float32
	PlatformChild      *datamodel.Instance
}

// Packet85Layer ID_PHYSICS - client <-> server
type Packet85Layer struct {
	SubPackets []*Packet85LayerSubpacket
}

func (b *extendedReader) readPhysicsData(data *PhysicsData, motors bool, reader PacketReader) error {
	var err error
	if motors {
		data.Motors, err = b.readMotors()
		if err != nil {
			return err
		}
	}

	data.CFrame, err = b.readPhysicsCFrame()
	if err != nil {
		return err
	}
	data.LinearVelocity, err = b.readPhysicsVelocity()
	if err != nil {
		return err
	}
	data.RotationalVelocity, err = b.readPhysicsVelocity()
	if err != nil {
		return err
	}
	hasPlatformChild, err := b.readBoolByte()
	if err != nil || !hasPlatformChild {
		return err
	}
	reference, err := b.readObject(reader.Context())
	if err != nil {
		return err
	}
	data.PlatformChild, _ = reader.Context().InstancesByReference.TryGetInstance(reference)
	return nil
}

func (b *extendedReader) DecodePacket85Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	context := reader.Context()
	layer := &Packet85Layer{}
	for {
		reference, err := b.readObject(reader.Context())
		if err != nil {
			return layer, err
		}
		if reference.IsNull {
			break
		}
		layers.Root.Logger.Println("reading physics for ref", reference.String())
		subpacket := &Packet85LayerSubpacket{}
		// ignore errors
		subpacket.Data.Instance, _ = context.InstancesByReference.TryGetInstance(reference)

		myFlags, err := b.readUint8()
		if err != nil {
			return layer, err
		}
		subpacket.NetworkHumanoidState = myFlags & 0x1F

		if reader.IsClient() {
			err = b.readPhysicsData(&subpacket.Data, true, reader)
			if err != nil {
				return layer, err
			}
		} else {
			subpacket.Data.Motors, err = b.readMotors()
			if err != nil {
				return layer, err
			}
			numEntries, err := b.readUint8()
			if err != nil {
				return layer, err
			}
			layers.Root.Logger.Println("reading movement history,", numEntries, "entries")
			subpacket.History = make([]*PhysicsData, numEntries)
			for i := 0; i < int(numEntries); i++ {
				subpacket.History[i] = new(PhysicsData)
				subpacket.History[i].Interval, err = b.readFloat32BE()
				if err != nil {
					return layer, err
				}
				b.readPhysicsData(subpacket.History[i], false, reader)
				if err != nil {
					return layer, err
				}
			}
		}

		if (myFlags>>5)&1 == 0 { // has children
			var object datamodel.Reference
			// peerID system shouldn't have caching problems anymore
			// TODO: remove cache hack
			for object, err = b.readObject(reader.Context()); (err == nil || err == ErrCacheReadOOB) && !object.IsNull; object, err = b.readObject(reader.Context()) {
				layers.Root.Logger.Println("reading physics child for ref", object.String())
				child := new(PhysicsData)
				// ignore errors
				child.Instance, _ = context.InstancesByReference.TryGetInstance(object)

				err = b.readPhysicsData(child, true, reader)
				if err != nil {
					return layer, err
				}

				subpacket.Children = append(subpacket.Children, child)
			}
			if err != nil {
				return layer, err
			}
		}

		layer.SubPackets = append(layer.SubPackets, subpacket)
	}
	return layer, nil
}

func (b *extendedWriter) writePhysicsData(val *PhysicsData, motors bool, writer PacketWriter) error {
	var err error
	if motors {
		err = b.writeMotors(val.Motors)
		if err != nil {
			return err
		}
	}

	err = b.writePhysicsCFrame(val.CFrame)
	if err != nil {
		return err
	}

	err = b.writePhysicsVelocity(val.LinearVelocity)
	if err != nil {
		return err
	}

	err = b.writePhysicsVelocity(val.RotationalVelocity)
	if err != nil {
		return err
	}

	err = b.writeBoolByte(val.PlatformChild != nil)
	if err != nil {
		return err
	}

	if val.PlatformChild == nil {
		return nil
	}

	err = b.writeObject(val.PlatformChild, writer.Context())
	return err
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet85Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		if subpacket.Data.Instance == nil {
			println("WARNING: skipping 0x85 serialize because instance doesn't exist yet; to client: ", writer.ToClient())
			continue
		}
		err = stream.writeObject(subpacket.Data.Instance, writer.Context())
		if err != nil {
			return err
		}
		var header uint8
		header = subpacket.NetworkHumanoidState
		if len(subpacket.Children) == 0 { // if bit is OFF, then children are ON -- what kind of logic is this?
			header |= 1 << 5
		}
		err = stream.WriteByte(header)
		if err != nil {
			return err
		}

		if !writer.ToClient() { // Writing to server, don't include history
			err = stream.writePhysicsData(&subpacket.Data, true, writer)
			if err != nil {
				return err
			}
		} else {
			err = stream.writeMotors(subpacket.Data.Motors)
			if err != nil {
				return err
			}
			err = stream.WriteByte(uint8(len(subpacket.History)))
			if err != nil {
				return err
			}
			for i := 0; i < int(len(subpacket.History)); i++ {
				err = stream.writeFloat32BE(subpacket.History[i].Interval)
				if err != nil {
					return err
				}
				err = stream.writePhysicsData(subpacket.History[i], false, writer)
				if err != nil {
					return err
				}
			}
		}

		for j := 0; j < len(subpacket.Children); j++ {
			child := subpacket.Children[j]
			if child.Instance == nil {
				println("WARNING: 0x85 skipping serialize because child doesn't exist yet!")
				continue
			}
			err = stream.writeObject(child.Instance, writer.Context())
			if err != nil {
				return err
			}

			err = stream.writePhysicsData(child, true, writer)
			if err != nil {
				return err
			}
		}
		if len(subpacket.Children) > 0 {
			err = stream.writeObject(nil, writer.Context()) // Terminator for children
			if err != nil {
				return err
			}
		}
	}
	return stream.WriteByte(0x00) // reference to NULL instance; terminator
}

func (layer *Packet85Layer) String() string {
	return fmt.Sprintf("ID_PHYSICS: %d items", len(layer.SubPackets))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet85Layer) TypeString() string {
	return "ID_PHYSICS"
}

// Type implements RakNetPacket.Type()
func (Packet85Layer) Type() byte {
	return 0x85
}
