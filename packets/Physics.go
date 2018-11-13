package packets

import (
    "github.com/gskartwii/roblox-dissector/util"
    "github.com/gskartwii/roblox-dissector/bitstreams"
	"github.com/gskartwii/rbxfile"
)

// PhysicsPacketSubpacket represents physics replication for one instance
type PhysicsPacketSubpacket struct {
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
	Instance           *rbxfile.Instance
	CFrame             rbxfile.ValueCFrame
	LinearVelocity     rbxfile.ValueVector3
	RotationalVelocity rbxfile.ValueVector3
	Motors             []bitstreams.PhysicsMotor
	Interval           float32
	PlatformChild      *rbxfile.Instance
}

// PhysicsPacket ID_PHYSICS - client <-> server
type PhysicsPacket struct {
	SubPackets []*PhysicsPacketSubpacket
}

// NewPhysicsPacket Initializes a new PhysicsPacket
func NewPhysicsPacket() *PhysicsPacket {
	return &PhysicsPacket{}
}

func (b *PacketReaderBitstream) readPhysicsData(data *PhysicsData, motors bool, reader util.PacketReader) error {
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
		reader.Context().InstancesByReferent.OnAddInstance(referent, func(inst *rbxfile.Instance) {
			data.PlatformChild = inst
		})
		return nil
	}
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodePhysicsPacket(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {

	context := reader.Context()
	layer := NewPhysicsPacket()
	for {
		referent, err := thisBitstream.ReadObject(reader.Caches())
		// unordered packets may have problems with caches
		if err != nil && err != CacheReadOOB {
			return layer, err
		}
		if referent.IsNull() {
			break
		}
		layers.Root.Logger.Println("reading physics for ref", referent.String())
		subpacket := &PhysicsPacketSubpacket{}
		// TODO: generic function for this
		if err != CacheReadOOB {
            // Do not try to add callbacks on instances that don't have scopes
			context.InstancesByReferent.OnAddInstance(referent, func(inst *rbxfile.Instance) {
				subpacket.Data.Instance = inst
			})
		}

		myFlags, err := thisBitstream.ReadUint8()
		if err != nil {
			return layer, err
		}
		subpacket.NetworkHumanoidState = myFlags & 0x1F

		if reader.IsClient() {
			err = thisBitstream.ReadPhysicsData(&subpacket.Data, true, reader)
			if err != nil {
				return layer, err
			}
		} else {
			subpacket.Data.Motors, err = thisBitstream.ReadMotors()
			if err != nil {
				return layer, err
			}
			numEntries, err := thisBitstream.ReadUint8()
			if err != nil {
				return layer, err
			}
			layers.Root.Logger.Println("reading movement history,", numEntries, "entries")
			subpacket.History = make([]*PhysicsData, numEntries)
			for i := 0; i < int(numEntries); i++ {
				subpacket.History[i] = new(PhysicsData)
				subpacket.History[i].Interval, err = thisBitstream.ReadFloat32BE()
				if err != nil {
					return layer, err
				}
				thisBitstream.ReadPhysicsData(subpacket.History[i], false, reader)
				if err != nil {
					return layer, err
				}
			}
		}

		if (myFlags>>5)&1 == 0 { // has children
			var object Referent
			for object, err = thisBitstream.ReadObject(reader.Caches()); (err == nil || err == CacheReadOOB) && !object.IsNull(); object, err = thisBitstream.ReadObject(reader.Caches()) {
				layers.Root.Logger.Println("reading physics child for ref", object.String())
				child := new(PhysicsData)
				if err != CacheReadOOB { // TODO: hack! unordered packets may have problems with caches
					context.InstancesByReferent.OnAddInstance(object, func(inst *rbxfile.Instance) {
						child.Instance = inst
					})
				}

				err = thisBitstream.ReadPhysicsData(child, true, reader)
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

func (b *PacketWriterBitstream) writePhysicsData(val *PhysicsData, motors bool, writer util.PacketWriter) error {
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

func (layer *PhysicsPacket) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
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
		err = stream.WriteObject(subpacket.Data.Instance, writer.Caches())
		if err != nil {
			return err
		}
		var header uint8
		header = subpacket.NetworkHumanoidState
		if len(subpacket.Children) != 0 {
			header |= 1 << 5
		}
		err = stream.WriteByte(header)
		if err != nil {
			return err
		}

		if writer.ToClient() {
			err = stream.WritePhysicsData(&subpacket.Data, true, writer)
			if err != nil {
				return err
			}
		} else {
			err = stream.WriteMotors(subpacket.Data.Motors)
			if err != nil {
				return err
			}
			err = stream.WriteByte(uint8(len(subpacket.History)))
			if err != nil {
				return err
			}
			for i := 0; i < int(len(subpacket.History)); i++ {
				err = stream.WriteFloat32BE(subpacket.History[i].Interval)
				if err != nil {
					return err
				}
				err = stream.WritePhysicsData(subpacket.History[i], false, writer)
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
			err = stream.WriteObject(child.Instance, writer.Caches())
			if err != nil {
				return err
			}

			err = stream.WritePhysicsData(child, true, writer)
			if err != nil {
				return err
			}
		}
		err = stream.WriteObject(nil, writer.Caches()) // Terminator for children
		if err != nil {
			return err
		}
	}
	return stream.WriteByte(0x00) // referent to NULL instance; terminator
}
