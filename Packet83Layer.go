package main
import "github.com/google/gopacket"
import "errors"
//import "encoding/hex"
//import "github.com/therecipe/qt/widgets"

type Packet83Subpacket interface {
	//Show() *widgets.QWidget_ITF
}

type Packet83_03 struct {
	MarkerId uint32
}
type Packet83_10 struct {
	TagId uint32
}
type Packet83_05 struct {
	Bool1 bool
	Int1 uint64
	Int2 uint32
	Int3 uint32
}
type Packet83_11 struct {
	SkipStats1 bool
	Stats_1_1 []byte
	Stats_1_2 float32
	Stats_1_3 float32
	Stats_1_4 float32
	Stats_1_5 bool

	SkipStats2 bool
	Stats_2_1 []byte
	Stats_2_2 float32
	Stats_2_3 uint32
	Stats_2_4 bool
	
	AvgPingMs float32
	AvgPhysicsSenderPktPS float32
	TotalDataKBPS float32
	TotalPhysicsKBPS float32
	DataThroughputRatio float32
}

type Packet83Layer struct {
	SubPackets []Packet83Subpacket
}

func NewPacket83Layer() Packet83Layer {
	return Packet83Layer{}
}

func extractPacketType(stream *ExtendedReader) (uint8, error) {
	ret, err := stream.Bits(2)
	if err != nil {
		return 0, err
	} else if ret != 0 {
		return uint8(ret), err
	}

	ret, err = stream.Bits(5)
	if err != nil {
		return 0, err
	}
	return uint8(ret), err
}

func DebugInfo(context *CommunicationContext, packet gopacket.Packet) string {
	if context.PacketFromClient(packet) {
		return "[C->S]"
	} else {
		return "[S->C]"
	}
}

func DecodePacket83Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket83Layer()

	packetType, err := extractPacketType(thisBitstream)
	if err != nil {
		return layer, err
	}
	context.WaitForSchema()
	context.WaitForDescriptors()
	defer context.FinishSchema()
	defer context.FinishDescriptors()
	classDescriptor := context.ClassDescriptor
	instanceSchema := context.InstanceSchema

	//if len(data) > 0x100 {
	//	println(DebugInfo(context, packet), hex.Dump(data[:0x100]))
	//}
	for packetType != 0 {
		if packetType == 3 {
			inner := &Packet83_03{}
			inner.MarkerId, err = thisBitstream.ReadUint32LE()
			if err != nil {
				return layer, err
			}
			println(DebugInfo(context, packet), "Receive marker", inner.MarkerId)
			layer.SubPackets = append(layer.SubPackets, inner)
		} else if packetType == 0x10 {
			inner := &Packet83_10{}
			inner.TagId, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			println(DebugInfo(context, packet), "Receive tag", inner.TagId)
			layer.SubPackets = append(layer.SubPackets, inner)
		} else if packetType == 0x05 {
			inner := &Packet83_05{}
			inner.Bool1, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}

			inner.Int1, err = thisBitstream.Bits(64)
			if err != nil {
				return layer, err
			}
			inner.Int2, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			inner.Int3, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			//println(DebugInfo(context, packet), "Receive 0x05", inner.Bool1, ",", inner.Int1, ",", inner.Int2, ",", inner.Int3)
		} else if packetType == 0x11 {
			inner := &Packet83_11{}
			inner.SkipStats1, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}
			println(DebugInfo(context, packet), "Skip stats 1:", inner.SkipStats1)
			if !inner.SkipStats1 {
				stringLen, err := thisBitstream.ReadUint32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_1_1, err = thisBitstream.ReadString(int(stringLen))
				if err != nil {
					return layer, err
				}

				inner.Stats_1_2, err = thisBitstream.ReadFloat32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_1_3, err = thisBitstream.ReadFloat32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_1_4, err = thisBitstream.ReadFloat32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_1_5, err = thisBitstream.ReadBool()
				if err != nil {
					return layer, err
				}
				print("Receive stats1", inner.Stats_1_1, ",", inner.Stats_1_2, ",", inner.Stats_1_3, ",", inner.Stats_1_4, ",", inner.Stats_1_5)
			}

			inner.SkipStats2, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}
			println(DebugInfo(context, packet), "Skip stats 2:", inner.SkipStats2)
			if !inner.SkipStats2 {
				stringLen, err := thisBitstream.ReadUint32LE()
				if err != nil {
					return layer, err
				}
				inner.Stats_2_1, err = thisBitstream.ReadString(int(stringLen))
				if err != nil {
					return layer, err
				}

				inner.Stats_2_2, err = thisBitstream.ReadFloat32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_2_3, err = thisBitstream.ReadUint32BE()
				if err != nil {
					return layer, err
				}
				inner.Stats_2_4, err = thisBitstream.ReadBool()
				if err != nil {
					return layer, err
				}
				print("Receive stats2", inner.Stats_2_1, ",", inner.Stats_2_2, ",", inner.Stats_2_3, ",", inner.Stats_2_4)
			}

			inner.AvgPingMs, err = thisBitstream.ReadFloat32BE()
			if err != nil {
				return layer, err
			}
			inner.AvgPhysicsSenderPktPS, err = thisBitstream.ReadFloat32BE()
			if err != nil {
				return layer, err
			}
			inner.TotalDataKBPS, err = thisBitstream.ReadFloat32BE()
			if err != nil {
				return layer, err
			}
			inner.TotalPhysicsKBPS, err = thisBitstream.ReadFloat32BE()
			if err != nil {
				return layer, err
			}
			inner.DataThroughputRatio, err = thisBitstream.ReadFloat32BE()
			if err != nil {
				return layer, err
			}
			println(DebugInfo(context, packet), "receive stats: %#v", inner)
		} else if packetType == 0x0B {
			_, err := DecodePacket83_0B(thisBitstream, context, packet, instanceSchema, classDescriptor)
			if err != nil {
				return layer, err
			}
		} else {
			return layer, errors.New("don't know how to parse replication subpacket")
		}

		packetType, err = extractPacketType(thisBitstream)
		if err != nil {
			return layer, err
		}
	}

	return layer, err
}
