package peer

import (
	"errors"
	"fmt"
	"io"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Cell represents a terrain cell package
type Cell struct {
	Material  uint8
	Occupancy uint8
}

// Chunk represents a cubic terrain chunk package
type Chunk struct {
	ChunkIndex datamodel.ValueVector3int32
	SideLength uint32
	CellCube   [][][]Cell
}

// Packet8DLayer represents ID_CLUSTER: server -> client
type Packet8DLayer struct {
	Instance *datamodel.Instance
	Chunks   []Chunk
}

func (thisStream *extendedReader) DecodePacket8DLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet8DLayer{}

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

	var header uint8
	for header, err = zstdStream.readUint8(); err == nil; header, err = zstdStream.readUint8() {
		subpacket := Chunk{}
		indexType := header & 0x60
		if indexType != 0 {
			if indexType == 0x20 {
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
				subpacket.ChunkIndex = datamodel.ValueVector3int32{
					X: int32(int16(x)),
					Y: int32(int16(y)),
					Z: int32(int16(z)),
				}
			} else if indexType == 0x40 {
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
				subpacket.ChunkIndex = datamodel.ValueVector3int32{
					X: int32(x),
					Y: int32(y),
					Z: int32(z),
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
			subpacket.ChunkIndex = datamodel.ValueVector3int32{
				X: int32(int8(x)),
				Y: int32(int8(y)),
				Z: int32(int8(z)),
			}
		}

		isEmpty, err := zstdStream.readBoolByte()
		if err != nil {
			return layer, err
		}
		if isEmpty {
			layers.Root.Logger.Println("skipping empty cluster")
			layer.Chunks = append(layer.Chunks, subpacket)
			continue
		}

		var cubeSideLength uint32 = 1 << (header & 0xF)
		subpacket.SideLength = uint32(cubeSideLength)
		cubeSize := cubeSideLength * cubeSideLength * cubeSideLength
		if cubeSize > 0x100000 {
			return layer, errors.New("cube size larger than max")
		}
		subpacket.CellCube = make([][][]Cell, subpacket.SideLength)
		for i := uint32(0); i < subpacket.SideLength; i++ {
			subpacket.CellCube[i] = make([][]Cell, subpacket.SideLength)
			for j := uint32(0); j < subpacket.SideLength; j++ {
				subpacket.CellCube[i][j] = make([]Cell, subpacket.SideLength)
			}
		}

		// Don't increment i here; we will do it inside the loop
		for i := uint32(0); i < cubeSize; {
			cellHeader, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			thisMaterial := cellHeader & 0x3F

			var occupancy uint8
			if cellHeader&0x40 != 0 {
				occupancy, err = zstdStream.readUint8()
				if err != nil {
					return layer, err
				}
			} else {
				occupancy = 0xFF
			}
			if thisMaterial == 0 {
				occupancy = 0
			}

			var count uint32
			if cellHeader&0x80 != 0 {
				countVal, err := zstdStream.readUint8()
				if err != nil {
					return layer, err
				}
				count = uint32(countVal)
			}
			// Implicit count of +1
			count++

			if i+count > cubeSize {
				return layer, errors.New("chunk overflow")
			}

			// copy the cell to the `count` next cells in the cube
			cell := Cell{
				Occupancy: occupancy,
				Material:  thisMaterial,
			}

			// Terrain coordinate system:
			// 0 => x=0 y=0 z=0
			// 1 => x=1 y=0 z=0
			// maxX => x=0 y=0 z=1
			// maxX*maxZ => x=0 y=1 z=0
			// yeah...
			sideLength := subpacket.SideLength
			for cubeIndex := uint32(i); cubeIndex < i+count; cubeIndex++ {
				xCoord := cubeIndex % sideLength
				zCoord := (cubeIndex / sideLength) % sideLength
				yCoord := (cubeIndex / sideLength) / sideLength
				subpacket.CellCube[xCoord][yCoord][zCoord] = cell
			}

			i += count
		}
		layer.Chunks = append(layer.Chunks, subpacket)
	}
	if err == io.EOF { // eof is normal, marks end of packet here
		layers.Root.Logger.Println("Normal EOF when parsing")
		return layer, nil
	}

	return layer, err
}

func (layer *Packet8DLayer) serializeChunks(writer PacketWriter, stream *extendedWriter) error {
	return errors.New("serialize chunks not implemented")
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet8DLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteObject(layer.Instance, writer)
	if err != nil {
		return err
	}

	zstdStream := stream.wrapZstd()

	err = layer.serializeChunks(writer, stream)
	if err != nil {
		zstdStream.Close()
		return err
	}

	return zstdStream.Close()
}

func (layer *Packet8DLayer) String() string {
	return fmt.Sprintf("ID_CLUSTER: %d terrain chunks", len(layer.Chunks))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet8DLayer) TypeString() string {
	return "ID_CLUSTER"
}

// Type implements RakNetPacket.Type()
func (Packet8DLayer) Type() byte {
	return 0x8D
}
