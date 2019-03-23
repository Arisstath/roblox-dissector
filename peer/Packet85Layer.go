package peer

import (
	"fmt"

	"github.com/gskartwii/roblox-dissector/datamodel"
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

// NewPacket85Layer Initializes a new Packet85Layer
func NewPacket85Layer() *Packet85Layer {
	return &Packet85Layer{}
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
	referent, err := b.readObject(reader.Caches())
	if err != CacheReadOOB {
		reader.Context().InstancesByReferent.OnAddInstance(referent, func(inst *datamodel.Instance) {
			data.PlatformChild = inst
		})
		return nil
	}
	return err
}

func (thisBitstream *extendedReader) DecodePacket85Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {

	context := reader.Context()
	layer := NewPacket85Layer()
	for {
		referent, err := thisBitstream.readObject(reader.Caches())
		// unordered packets may have problems with caches
		if err != nil && err != CacheReadOOB {
			return layer, err
		}
		if referent.IsNull {
			break
		}
		layers.Root.Logger.Println("reading physics for ref", referent.String())
		subpacket := &Packet85LayerSubpacket{}
		// TODO: generic function for this
		if err != CacheReadOOB {
			context.InstancesByReferent.OnAddInstance(referent, func(inst *datamodel.Instance) {
				subpacket.Data.Instance = inst
			})
		}

		myFlags, err := thisBitstream.readUint8()
		if err != nil {
			return layer, err
		}
		subpacket.NetworkHumanoidState = myFlags & 0x1F

		if reader.IsClient() {
			err = thisBitstream.readPhysicsData(&subpacket.Data, true, reader)
			if err != nil {
				return layer, err
			}
		} else {
			subpacket.Data.Motors, err = thisBitstream.readMotors()
			if err != nil {
				return layer, err
			}
			numEntries, err := thisBitstream.readUint8()
			if err != nil {
				return layer, err
			}
			layers.Root.Logger.Println("reading movement history,", numEntries, "entries")
			subpacket.History = make([]*PhysicsData, numEntries)
			for i := 0; i < int(numEntries); i++ {
				subpacket.History[i] = new(PhysicsData)
				subpacket.History[i].Interval, err = thisBitstream.readFloat32BE()
				if err != nil {
					return layer, err
				}
				thisBitstream.readPhysicsData(subpacket.History[i], false, reader)
				if err != nil {
					return layer, err
				}
			}
		}

		if (myFlags>>5)&1 == 0 { // has children
			var object datamodel.Reference
			for object, err = thisBitstream.readObject(reader.Caches()); (err == nil || err == CacheReadOOB) && !object.IsNull; object, err = thisBitstream.readObject(reader.Caches()) {
				layers.Root.Logger.Println("reading physics child for ref", object.String())
				child := new(PhysicsData)
				if err != CacheReadOOB { // TODO: hack! unordered packets may have problems with caches
					context.InstancesByReferent.OnAddInstance(object, func(inst *datamodel.Instance) {
						child.Instance = inst
					})
				}

				err = thisBitstream.readPhysicsData(child, true, reader)
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

	err = b.writeObject(val.PlatformChild, writer.Caches())
	return err
}

func (layer *Packet85Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x85)
	if err != nil {
		return err
	}
	for i := 0; i < len(layer.SubPackets); i++ {
		subpacket := layer.SubPackets[i]
		if subpacket.Data.Instance == nil {
			println("WARNING: skipping 0x85 serialize because instance doesn't exist yet")
			continue
		}
		err = stream.writeObject(subpacket.Data.Instance, writer.Caches())
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
			err = stream.writeObject(child.Instance, writer.Caches())
			if err != nil {
				return err
			}

			err = stream.writePhysicsData(child, true, writer)
			if err != nil {
				return err
			}
		}
		err = stream.writeObject(nil, writer.Caches()) // Terminator for children
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}

func (layer *Packet85Layer) String() string {
	return fmt.Sprintf("ID_PHYSICS: %d items", len(layer.SubPackets))
}

func (Packet85Layer) TypeString() string {
	return "ID_PHYSICS"
}
