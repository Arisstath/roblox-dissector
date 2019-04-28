package peer

import (
	"errors"
	"fmt"
	"io"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

type Cell struct {
	Material  uint8
	Occupancy uint8
	Count     uint32
}
type Chunk struct {
	Header    uint8
	ChunkSize rbxfile.ValueVector3
	Contents  []Cell
}

// ID_ROBLOX_CLUSTER: server -> client
type Packet8DLayer struct {
	Instance *datamodel.Instance
	Chunks   []Chunk
}

func NewPacket8DLayer() *Packet8DLayer {
	return &Packet8DLayer{}
}

func (thisStream *extendedReader) DecodePacket8DLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket8DLayer()

	context := reader.Context()

	reference, err := thisStream.readObject(reader.Caches())
	if err != nil {
		return nil, err
	}
	if reference.IsNull {
		return nil, errors.New("cluster instance is null")
	}
	layer.Instance, err = context.InstancesByReference.TryGetInstance(reference)
	if err != nil {
		return layer, err
	}

	layers.Root.Logger.Printf("Reading cluster for terrain: %s\n", layer.Instance.Name())
	zstdStream, err := thisStream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	header, err := zstdStream.readUint8()
	for err == nil {
		subpacket := Chunk{}
		subpacket.Header = header & 0xF

		sizeception := header & 0x60
		if sizeception != 0 {
			if sizeception == 0x20 {
				x, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				y, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				z, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				subpacket.ChunkSize = rbxfile.ValueVector3{
					X: float32(int16(x)),
					Y: float32(int16(y)),
					Z: float32(int16(z)),
				}
			} else if sizeception == 0x40 {
				x, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				y, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				z, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				subpacket.ChunkSize = rbxfile.ValueVector3{
					X: float32(int32(x)),
					Y: float32(int32(y)),
					Z: float32(int32(z)),
				}
			} else {
				return layer, errors.New("invalid chunk header")
			}
		} else {
			x, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			y, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			z, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			subpacket.ChunkSize = rbxfile.ValueVector3{
				X: float32(int8(x)),
				Y: float32(int8(y)),
				Z: float32(int8(z)),
			}
		}

		validCheck, err := zstdStream.readUint8()
		if err != nil {
			return layer, err
		}
		if validCheck != 0 {
			layers.Root.Logger.Println("valid check failed! trying to continue")
			continue
		}

		cubeSideLength := 1 << subpacket.Header
		cubeSize := cubeSideLength * cubeSideLength * cubeSideLength
		if cubeSize > 0x100000 {
			return layer, errors.New("cube size larger than max")
		}
		subpacket.Contents = make([]Cell, 0)

		for remainingCount := cubeSize; remainingCount > 0; {
			cellHeader, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			var thisMaterial uint8 = cellHeader & 0x3F
			var occupancy uint8
			var count int
			if cellHeader&0x40 != 0 {
				occupancy, err = zstdStream.readUint8()
				if err != nil {
					return layer, err
				}
			} else {
				occupancy = 0xFF
			}

			if cellHeader&0x80 != 0 {
				myCount, err := zstdStream.readUint8()
				if err != nil {
					return layer, err
				}
				count = int(myCount) + 1
			} else {
				count = 1
			}

			remainingCount -= int(count)
			if thisMaterial == 0 {
				occupancy = 0
			}
			subpacket.Contents = append(subpacket.Contents, Cell{
				Occupancy: occupancy,
				Material:  thisMaterial,
			})
			layers.Root.Logger.Printf("Read cell with head:%d, material:%d, occ:%d, count:%d\n", cellHeader, thisMaterial, occupancy, count)
		}
		header, err = zstdStream.readUint8()
		layer.Chunks = append(layer.Chunks, subpacket)
	}
	if err == io.EOF { // eof is normal, marks end of packet here
		layers.Root.Logger.Println("Normal EOF when parsing")
		return layer, nil
	}

	return layer, err
}

func (layer *Packet8DLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return errors.New("not implemented!")
}

func (layer *Packet8DLayer) String() string {
	return fmt.Sprintf("ID_CLUSTER: WIP")
}

func (Packet8DLayer) TypeString() string {
	return "ID_CLUSTER"
}

func (Packet8DLayer) Type() byte {
	return 0x8D
}
