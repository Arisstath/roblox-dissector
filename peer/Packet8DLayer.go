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
	Int1 uint8
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

type chunkDeserializer interface {
	ReadByte() (byte, error)
	readUint8() (uint8, error)
	readBoolByte() (bool, error)
	readUint16BE() (uint16, error)
	readUint32BE() (uint32, error)
}

func deserializeChunks(stream chunkDeserializer) ([]Chunk, error) {
	var header uint8
	var x, y, z int32
	var chunks []Chunk
	var err error
	for header, err = stream.readUint8(); err == nil; header, err = stream.readUint8() {
		subpacket := Chunk{}
		indexType := header & 0x30
		if indexType != 0 {
			if indexType == 0x10 {
				xDiff, err := stream.readUint16BE()
				if err != nil {
					return chunks, err
				}
				yDiff, err := stream.readUint16BE()
				if err != nil {
					return chunks, err
				}
				zDiff, err := stream.readUint16BE()
				if err != nil {
					return chunks, err
				}
				x += int32(int16(xDiff))
				y += int32(int16(yDiff))
				z += int32(int16(zDiff))
			} else if indexType == 0x20 {
				xDiff, err := stream.readUint32BE()
				if err != nil {
					return chunks, err
				}
				yDiff, err := stream.readUint32BE()
				if err != nil {
					return chunks, err
				}
				zDiff, err := stream.readUint32BE()
				if err != nil {
					return chunks, err
				}
				x += int32(xDiff)
				y += int32(yDiff)
				z += int32(zDiff)
			} else {
				return chunks, errors.New("invalid chunk header")
			}
		} else {
			xDiff, err := stream.readUint8()
			if err != nil {
				return chunks, err
			}
			yDiff, err := stream.readUint8()
			if err != nil {
				return chunks, err
			}
			zDiff, err := stream.readUint8()
			if err != nil {
				return chunks, err
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
		subpacket.Int1 = header >> 6 // unknown value

		isEmpty, err := stream.readBoolByte()
		if err != nil {
			return chunks, err
		}
		if isEmpty {
			chunks = append(chunks, subpacket)
			continue
		}
		cubeSize := cubeSideLength * cubeSideLength * cubeSideLength
		if cubeSize > 0x100000 {
			return chunks, errors.New("cube size larger than max")
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
			cellHeader, err := stream.readUint8()
			if err != nil {
				return chunks, err
			}
			thisMaterial := cellHeader & 0x3F

			var occupancy uint8
			if cellHeader&0x40 != 0 {
				occupancy, err = stream.readUint8()
				if err != nil {
					return chunks, err
				}
			} else {
				occupancy = 0xFF
			}
			if thisMaterial == 0 {
				occupancy = 0
			}

			var count uint32
			if cellHeader&0x80 != 0 {
				countVal, err := stream.readUint8()
				if err != nil {
					return chunks, err
				}
				count = uint32(countVal)
			}
			// Implicit count of +1
			count++

			if i+count > cubeSize {
				return chunks, errors.New("chunk overflow")
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
				xCoord, yCoord, zCoord := cellIndexToCoords(sideLength, cubeIndex)
				subpacket.CellCube[xCoord][yCoord][zCoord] = cell
			}

			i += count
		}
		chunks = append(chunks, subpacket)
	}

	if err == io.EOF {
		return chunks, nil
	}

	return chunks, err
}

func (thisStream *extendedReader) DecodePacket8DLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet8DLayer{}

	context := reader.Context()

	reference, err := thisStream.ReadObject(reader)
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

	layer.Chunks, err = deserializeChunks(zstdStream)

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
		for i := uint8(0); i <= 0xF; i++ {
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
		extraInt := chunk.Int1 << 6

		if areInBounds(-127, 127, diffX, diffY, diffZ) {
			err = stream.WriteByte(cubeSideLog2 | extraInt)
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
			err = stream.WriteByte(0x10 | cubeSideLog2 | extraInt)
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
			err = stream.WriteByte(0x20 | cubeSideLog2 | extraInt)
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
			for isRunLength := true; isRunLength; {
				rleCount++
				if rleCount+cellIndex >= cubeSize || rleCount >= 256 {
					break
				}
				nextCellX, nextCellY, nextCellZ := cellIndexToCoords(chunk.SideLength, rleCount+cellIndex)
				nextCell := chunk.CellCube[nextCellX][nextCellY][nextCellZ]

				isRunLength = currCell.Material == nextCell.Material && currCell.Occupancy == nextCell.Occupancy
			}

			// currCell won't be the cell being encoded here!
			material := currCell.Material
			occupancy := currCell.Occupancy
			count := uint8(rleCount - 1)

			header := material
			customCount := count != 0
			customOccupancy := occupancy != 255 && material != 0
			if customCount {
				header |= 0x80
			}
			if customOccupancy {
				header |= 0x40
			}

			err = stream.WriteByte(header)
			if err != nil {
				return err
			}

			if customOccupancy {
				err = stream.WriteByte(occupancy)
				if err != nil {
					return err
				}
			}

			if customCount {
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
