package main
import "github.com/google/gopacket"
//import "encoding/hex"
import "fmt"
import "errors"
import "github.com/davecgh/go-spew/spew"

type UDim struct {
	Scale float32
	Offset uint32
}

type UDim2 struct {
	X UDim
	Y UDim
}

type Vector2 struct {
	X float32
	Y float32
}
type Vector3 struct {
	X float32
	Y float32
	Z float32
}
type Vector2uint16 struct {
	X uint16
	Y uint16
}
type Vector3uint16 struct {
	X uint16
	Y uint16
	Z uint16
}

type Ray struct {
	Origin Vector3
	Direction Vector3
}

type Color3 struct {
	R float32
	G float32
	B float32
}
type Color3uint8 struct {
	R uint8
	G uint8
	B uint8
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

type ReplicationInstance struct {
	Referent string
	ReferentInt1 uint32
	Int1 uint32
	ClassName string
	Bool1 bool
	Referent2 string
	ReferentInt2 uint32
}

type Packet83_0B struct {
	Instances []*ReplicationInstance
}

type Packet83Layer struct {
	SubPackets []interface{}
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
			thisBitstream.Align()
			arrayLen, err := thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			gzipStream, err := thisBitstream.RegionToGZipStream()
			if err != nil {
				return layer, err
			}
			//dump, _ := gzipStream.DumpAll()
			//println(DebugInfo(context, packet), hex.Dump(dump[:0x100]))

			var i uint32
			for i = 0; i < arrayLen; i++ {
				thisInstance := &ReplicationInstance{}
				thisInstance.Referent, thisInstance.ReferentInt1, err = gzipStream.ReadJoinReferent()
				if err != nil {
					return layer, err
				}

				classIDx, err := gzipStream.Bits(9)
				if err != nil {
					return layer, err
				}
				realIDx := (classIDx & 1 << 8) | classIDx >> 1
				println(DebugInfo(context, packet), "Our IDx: ", realIDx)
				thisInstance.ClassName = classDescriptor[uint32(realIDx)]
				if int(realIDx) > int(len(instanceSchema)) {
					return layer, errors.New(fmt.Sprintf("idx %d is higher than %d", realIDx, len(context.InstanceSchema)))
				}

				thisPropertySchema := instanceSchema[realIDx].PropertySchema

				thisInstance.Bool1, err = gzipStream.ReadBool()
				if err != nil {
					return layer, err
				}

				for _, property := range thisPropertySchema {
					if !property.Bool1 {
						continue
					}
					if property.Type == "bool" {
						if err != nil {
							return layer, err
						}
						thisBool, err := gzipStream.ReadBool()
						if err != nil {
							return layer, err
						}
						println(DebugInfo(context, packet), "Read", property.Name, "bool 1 bit", thisBool)
					} else {
						isDefault, err := gzipStream.ReadBool()
						if err != nil {
							return layer, err
						}
						if isDefault {
							println(DebugInfo(context, packet), "Read", property.Name, "1 bit: default")
						} else {
							if property.Type == "string" || property.Type == "BinaryString" || property.Type == "ProtectedString" {
								stringLen, err := gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								if stringLen > 0x1E84800 {
									println(DebugInfo(context, packet), "Sanity check: string len too high", stringLen)
									break
								}
								result, err := gzipStream.ReadASCII(int(stringLen))
								println(DebugInfo(context, packet), "Read", property.Name, result)
							} else if property.Type == "int" {
								val, err := gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, val)
							} else if property.Type == "float" {
								val, err := gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, val)
							} else if property.Type == "double" {
								val, err := gzipStream.ReadFloat64BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, val)
							} else if property.Type == "UDim" {
								val := UDim{}
								val.Scale, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Offset, err = gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "UDim2" {
								val := UDim2{UDim{}, UDim{}}
								val.X.Scale, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.X.Offset, err = gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								val.Y.Scale, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Y.Offset, err = gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Ray" {
								val := Ray{Vector3{}, Vector3{}}
								val.Origin.X, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Origin.Y, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Origin.Z, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Direction.X, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Direction.Y, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Direction.Z, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Axes" || property.Type == "Faces" {
								val, err := gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}

								println(DebugInfo(context, packet), "Read", property.Name, val)
							} else if property.Type == "BrickColor" {
								val, err := gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}

								println(DebugInfo(context, packet), "Read", property.Name, val, "(look it up yourself i'm lazy)")
							} else if property.Type == "Color3" {
								val := Color3{}
								val.R, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.G, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.B, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Color3uint8" {
								val := Color3uint8{}
								val.R, err = gzipStream.ReadByte()
								if err != nil {
									return layer, err
								}
								val.G, err = gzipStream.ReadByte()
								if err != nil {
									return layer, err
								}
								val.B, err = gzipStream.ReadByte()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Vector2" {
								val := Vector2{}
								val.X, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								val.Y, err = gzipStream.ReadFloat32BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Vector3" {
								val := Vector3{}
								isInteger, err := gzipStream.ReadBool()
								if !isInteger {
									val.X, err = gzipStream.ReadFloat32BE()
									if err != nil {
										return layer, err
									}
									val.Y, err = gzipStream.ReadFloat32BE()
									if err != nil {
										return layer, err
									}
									val.Z, err = gzipStream.ReadFloat32BE()
									if err != nil {
										return layer, err
									}
								} else {
									x, err := gzipStream.Bits(11)
									if err != nil {
										return layer, err
									}
									x_short := uint16(((x & 0xFFF8) >> 3) | ((x & 7) << 8))
									if x_short & 0x400 != 0 {
										x_short |= 0xFC00
									}

									y, err := gzipStream.Bits(11)
									if err != nil {
										return layer, err
									}

									z, err := gzipStream.Bits(11)
									if err != nil {
										return layer, err
									}

									z_short := uint16(((z & 0xFFF8) >> 3) | ((z & 7) << 8))
									if z_short & 0x400 != 0 {
										z_short |= 0xFC00
									}

									val.X = float32(int16(x_short))
									val.Y = float32(y)
									val.Z = float32(int16(z_short))
								}

								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Vector2uint16" {
								val := Vector2uint16{}
								val.X, err = gzipStream.ReadUint16BE()
								if err != nil {
									return layer, err
								}
								val.Y, err = gzipStream.ReadUint16BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Vector3uint16" {
								val := Vector3uint16{}
								val.X, err = gzipStream.ReadUint16BE()
								if err != nil {
									return layer, err
								}
								val.Y, err = gzipStream.ReadUint16BE()
								if err != nil {
									return layer, err
								}
								val.Z, err = gzipStream.ReadUint16BE()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
							} else if property.Type == "Object" {
								val, valint, err := gzipStream.ReadJoinReferent()
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, val, valint)
							} else if property.IsEnum {
								val, err := gzipStream.Bits(int(property.BitSize))
								if err != nil {
									return layer, err
								}
								println(DebugInfo(context, packet), "Read", property.Name, val)
							} else {
								println(DebugInfo(context, packet), "Read", property.Name, "x bit")
								break
							}
						}
					}
				}
				thisInstance.Referent2, thisInstance.ReferentInt2, err = gzipStream.ReadJoinReferent()
				if err != nil {
					return layer, err
				}

				println(DebugInfo(context, packet), "Read instance", thisInstance.Referent, ",", thisInstance.ReferentInt1, ",",thisInstance.Int1, ",", thisInstance.ClassName, ",", thisInstance.Referent2, ",", thisInstance.ReferentInt2)
				break

				gzipStream.Align()
			}
		} else {
			println(DebugInfo(context, packet), "Decode packet of ID", packetType)
		}

		packetType, err = extractPacketType(thisBitstream)
		if err != nil {
			return layer, err
		}
	}

	return layer, err
}
