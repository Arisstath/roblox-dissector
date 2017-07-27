package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"
//import "encoding/hex"
import "fmt"

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
	Int1 uint32
	ClassName string
	Bool1 bool
	Referent2 string
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

func DecodePacket83Layer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket83Layer()
	thisBitstream := &ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

	packetType, err := extractPacketType(thisBitstream)

	if err != nil {
		return layer, err
	}

	for packetType != 0 {
		if packetType == 3 {
			inner := &Packet83_03{}
			inner.MarkerId, err = thisBitstream.ReadUint32LE()
			if err != nil {
				return layer, err
			}
			println("Receive marker", inner.MarkerId)
			layer.SubPackets = append(layer.SubPackets, inner)
		} else if packetType == 0x10 {
			inner := &Packet83_10{}
			inner.TagId, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			println("Receive tag", inner.TagId)
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
			//println("Receive 0x05", inner.Bool1, ",", inner.Int1, ",", inner.Int2, ",", inner.Int3)
		} else if packetType == 0x11 {
			inner := &Packet83_11{}
			inner.SkipStats1, err = thisBitstream.ReadBool()
			if err != nil {
				return layer, err
			}
			print("Skip stats:", inner.SkipStats1)
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
			if !inner.SkipStats2 {
				stringLen, err := thisBitstream.ReadUint32BE()
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
			fmt.Println("receive stats: %#v", inner)
		} else if packetType == 0xB {
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
			//println(hex.Dump(dump[:0x100]))

			var i uint32
			for i = 0; i < arrayLen; i++ {
				thisInstance := &ReplicationInstance{}
				thisInstance.Referent, err = gzipStream.ReadJoinReferent()
				if err != nil {
					return layer, err
				}

				classIDx, err := gzipStream.Bits(9)
				if err != nil {
					return layer, err
				}
				realIDx := (classIDx & 1 << 8) | classIDx >> 1
				println("Our IDx: ", realIDx)
				thisInstance.ClassName = context.ClassDescriptor[uint32(realIDx)]

				thisPropertySchema := context.InstanceSchema[realIDx].PropertySchema

				thisInstance.Bool1, err = gzipStream.ReadBool()
				if err != nil {
					return layer, err
				}

				for _, property := range thisPropertySchema {
					if property.Type == "bool" {
						if err != nil {
							return layer, err
						}
						thisBool, err := gzipStream.ReadBool()
						if err != nil {
							return layer, err
						}
						println("Read", property.Name, "bool 1 bit", thisBool)
					} else {
						isDefault, err := gzipStream.ReadBool()
						if err != nil {
							return layer, err
						}
						if isDefault {
							println("Read", property.Name, "1 bit: default")
						} else {
							if property.Type == "string" {
								stringLen, err := gzipStream.ReadUint32BE()
								if err != nil {
									return layer, err
								}
								if stringLen > 0x200 {
									println("Sanity check: string len too high", stringLen)
									break
								}
								result, err := gzipStream.ReadASCII(int(stringLen))
								println("Read", property.Name, result)
							} else {
								println("Read", property.Name, "x bit")
								break
							}
						}
					}
				}
				thisInstance.Referent2, err = gzipStream.ReadJoinReferent()
				if err != nil {
					return layer, err
				}

				println("Read instance", thisInstance.Referent, ",", thisInstance.Int1, ",", thisInstance.ClassName, ",", thisInstance.Referent2)
				break

				gzipStream.Align()
			}
		} else {
			println("Decode packet of ID", packetType)
		}

		packetType, err = extractPacketType(thisBitstream)
		if err != nil {
			return layer, err
		}
	}

	return layer, err
}
