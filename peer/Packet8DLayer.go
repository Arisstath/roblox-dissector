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

// IsEmpty returns a value indicating whether the chunk is empty,
// i.e. doesn't contain any non-air cells.
func (chunk Chunk) IsEmpty() bool {
	if chunk.SideLength == 0 {
		return true
	}
	if len(chunk.CellCube) == 0 {
		return true
	}
	for _, cx := range chunk.CellCube {
		for _, cy := range cx {
			for _, cz := range cy {
				if cz.Material != 0 && cz.Occupancy != 0 {
					return false
				}
			}
		}
	}

	return true
}

// IsCube returns a value indicating whether all sides of the
// chunk match its SideLength.
// Note that IsCube() returns false if the chunk is empty.
func (chunk Chunk) IsCube() bool {
	sideLength := int(chunk.SideLength)
	if len(chunk.CellCube) != sideLength {
		return false
	}

	for _, cx := range chunk.CellCube {
		if len(cx) != sideLength {
			return false
		}
		for _, cy := range chunk.CellCube {
			if len(cy) != sideLength {
				return false
			}
		}
	}

	return true
}

func cellIndexToCoords(sideLength uint32, cubeIndex uint32) (uint32, uint32, uint32) {
	xCoord := cubeIndex % sideLength
	zCoord := (cubeIndex / sideLength) % sideLength
	yCoord := (cubeIndex / sideLength) / sideLength

	return xCoord, yCoord, zCoord
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
	var x, y, z int32
	for header, err = zstdStream.readUint8(); err == nil; header, err = zstdStream.readUint8() {
		subpacket := Chunk{}
		indexType := header & 0x60
		if indexType != 0 {
			if indexType == 0x20 {
				xDiff, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				yDiff, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				zDiff, err := zstdStream.readUint16BE()
				if err != nil {
					return layer, err
				}
				x += int32(int16(xDiff))
				y += int32(int16(yDiff))
				z += int32(int16(zDiff))
			} else if indexType == 0x40 {
				xDiff, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				yDiff, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				zDiff, err := zstdStream.readUint32BE()
				if err != nil {
					return layer, err
				}
				x += int32(xDiff)
				y += int32(yDiff)
				z += int32(zDiff)
			} else {
				return layer, errors.New("invalid chunk header")
			}
		} else {
			xDiff, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			yDiff, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			zDiff, err := zstdStream.readUint8()
			if err != nil {
				return layer, err
			}
			x += int32(int8(xDiff))
			y += int32(int8(yDiff))
			z += int32(int8(zDiff))
		}
		subpacket.ChunkIndex = datamodel.ValueVector3int32{
			X: x,
			Y: y,
			Z: z,
		}

		var cubeSideLength uint32 = 1 << (header & 0xF)
		subpacket.SideLength = uint32(cubeSideLength)

		isEmpty, err := zstdStream.readBoolByte()
		if err != nil {
			return layer, err
		}
		if isEmpty {
			layers.Root.Logger.Println("skipping empty cluster")
			layer.Chunks = append(layer.Chunks, subpacket)
			continue
		}
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
				xCoord, yCoord, zCoord := cellIndexToCoords(sideLength, i)
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

type chunkSerializer interface {
	WriteByte(byte) error
	writeBoolByte(bool) error
	writeUint16BE(uint16) error
	writeUint32BE(uint32) error
}

func areInBounds(min int32, max int32, vals ...int32) bool {
	for _, val := range vals {
		if val < min || val > max {
			return false
		}
	}
	return true
}

func (layer *Packet8DLayer) serializeChunks(stream chunkSerializer) error {
	var lastX, lastY, lastZ int32
	var err error
	for _, chunk := range layer.Chunks {
		var cubeSideLog2 uint8 = 0xFF
		for i := uint8(0); i < 0xF; i++ {
			if 1<<i == chunk.SideLength {
				cubeSideLog2 = i
			}
		}
		if cubeSideLog2 == 0xFF {
			return errors.New("invalid cube side length")
		}

		index := chunk.ChunkIndex
		thisX := index.X
		thisY := index.Y
		thisZ := index.Z
		diffX := thisX - lastX
		diffY := thisY - lastY
		diffZ := thisZ - lastZ
		lastX = thisX
		lastY = thisY
		lastZ = thisZ

		if areInBounds(-127, 127, diffX, diffY, diffZ) {
			err = stream.WriteByte(cubeSideLog2)
			if err != nil {
				return err
			}
			err = stream.WriteByte(byte(diffX))
			if err != nil {
				return err
			}
			err = stream.WriteByte(byte(diffY))
			if err != nil {
				return err
			}
			err = stream.WriteByte(byte(diffZ))
			if err != nil {
				return err
			}
		} else if areInBounds(-32767, 32767, diffX, diffY, diffZ) {
			err = stream.WriteByte(0x20 | cubeSideLog2)
			if err != nil {
				return err
			}
			err = stream.writeUint16BE(uint16(diffX))
			if err != nil {
				return err
			}
			err = stream.writeUint16BE(uint16(diffY))
			if err != nil {
				return err
			}
			err = stream.writeUint16BE(uint16(diffZ))
			if err != nil {
				return err
			}
		} else {
			err = stream.WriteByte(0x40 | cubeSideLog2)
			if err != nil {
				return err
			}
			err := stream.writeUint32BE(uint32(diffX))
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(uint32(diffY))
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(uint32(diffZ))
			if err != nil {
				return err
			}
		}

		isEmpty := chunk.IsEmpty()
		err = stream.writeBoolByte(isEmpty)
		if err != nil {
			return err
		}
		if isEmpty {
			continue
		}

		if !chunk.IsCube() {
			return errors.New("non-empty chunk isn't a cube")
		}

		cubeSize := chunk.SideLength * chunk.SideLength * chunk.SideLength
		for cellIndex := uint32(0); cellIndex < cubeSize; {
			rleCount := uint32(0)
			currCellX, currCellY, currCellZ := cellIndexToCoords(chunk.SideLength, cellIndex)
			currCell := chunk.CellCube[currCellX][currCellY][currCellZ]
			for isRunLength := true; isRunLength && rleCount < 256 && rleCount+cellIndex < cubeSize; {
				rleCount++
				nextCellX, nextCellY, nextCellZ := cellIndexToCoords(chunk.SideLength, rleCount+cellIndex)
				nextCell := chunk.CellCube[nextCellX][nextCellY][nextCellZ]

				isRunLength = currCell.Material == nextCell.Material && currCell.Occupancy == nextCell.Occupancy
			}

			// currCell won't be the cell being encoded here!
			material := prevCell.Material
			occupancy := prevCell.Occupancy
			count := uint8(rleCount - 1)

			header := material
			if count != 255 {
				header |= 0x80
			}
			if occupancy != 255 {
				header |= 0x40
			}

			err = stream.WriteByte(header)
			if err != nil {
				return err
			}

			if occupancy != 255 {
				err = stream.WriteByte(occupancy)
				if err != nil {
					return err
				}
			}

			if count != 255 {
				err = stream.WriteByte(count)
				if err != nil {
					return err
				}
			}

			cellIndex += rleCount
		}
	}
	return nil
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet8DLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteObject(layer.Instance, writer)
	if err != nil {
		return err
	}

	zstdStream := stream.wrapZstd()

	err = layer.serializeChunks(zstdStream)
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
