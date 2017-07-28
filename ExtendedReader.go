package main
import "github.com/dgryski/go-bitstream"
import "encoding/binary"
import "errors"
import "net"
import "math"
import "bytes"
import "compress/gzip"
//import "fmt"

var englishTree *HuffmanEncodingTree

func init() {
	englishTree = GenerateHuffmanFromFrequencyTable([]uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 722, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11084, 58, 63, 1, 0, 31, 0, 317, 64, 64, 44, 0, 695, 62, 980, 266, 69, 67, 56, 7, 73, 3, 14, 2, 69, 1, 167, 9, 1, 2, 25, 94, 0, 195, 139, 34, 96, 48, 103, 56, 125, 653, 21, 5, 23, 64, 85, 44, 34, 7, 92, 76, 147, 12, 14, 57, 15, 39, 15, 1, 1, 1, 2, 3, 0, 3611, 845, 1077, 1884, 5870, 841, 1057, 2501, 3212, 164, 531, 2019, 1330, 3056, 4037, 848, 47, 2586, 2919, 4771, 1707, 535, 1106, 152, 1243, 100, 0, 2, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

type ExtendedReader struct {
	*bitstream.BitReader
}

func (b *ExtendedReader) Bits(len int) (uint64, error) {
	return b.ReadBits(len)
}

func (b *ExtendedReader) Bytes(dest []byte, len int) (error) {
	var Byte byte
	for i := 0; i < len; i++ {
		res, err := b.Bits(8)
		if err != nil {
			return err
		}
		Byte = byte(res)
		dest[i] = Byte
	}
	return nil
}

func (b *ExtendedReader) ReadBool() (bool, error) {
	res, err := b.ReadBit()
	return bool(res), err
}

func (b *ExtendedReader) ReadUint16BE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.Bytes(dest, 2)
	return binary.BigEndian.Uint16(dest), err
}

func (b *ExtendedReader) ReadUint16LE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.Bytes(dest, 2)
	return binary.LittleEndian.Uint16(dest), err
}

func (b *ExtendedReader) ReadBoolByte() (bool, error) {
	res, err := b.Bits(8)
	return res == 1, err
}

func (b *ExtendedReader) ReadUint24LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.Bytes(dest, 3)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *ExtendedReader) ReadUint32BE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.Bytes(dest, 4)
	return binary.BigEndian.Uint32(dest), err
}
func (b *ExtendedReader) ReadUint32LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.Bytes(dest, 4)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *ExtendedReader) ReadUint64BE() (uint64, error) {
	dest := make([]byte, 8)
	err := b.Bytes(dest, 8)
	return binary.BigEndian.Uint64(dest), err
}

func (b *ExtendedReader) ReadCompressed(dest []byte, size uint32, unsignedData bool) error {
	var currentByte uint32 = (size >> 3) - 1
	var byteMatch, halfByteMatch byte

	if unsignedData {
		byteMatch = 0
		halfByteMatch = 0
	} else {
		byteMatch = 0xFF
		halfByteMatch = 0xF0
	}

	for currentByte > 0 {
		res, err := b.ReadBool()
		if err != nil {
			return err
		}
		if res {
			dest[currentByte] = byteMatch
			currentByte--
		} else {
			err = b.Bytes(dest, int(currentByte + 1))
			return err
		}
	}

	res, err := b.ReadBool()
	if err != nil {
		return err
	}

	if res {

		res, err := b.Bits(4)
		if err != nil {
			return err
		}
		dest[currentByte] = byte(res) | halfByteMatch
	} else {
		err := b.Bytes(dest[currentByte:], 1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *ExtendedReader) ReadUint32BECompressed(unsignedData bool) (uint32, error) {
	dest := make([]byte, 4)
	err := b.ReadCompressed(dest, 32, unsignedData)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(dest), nil
}

func (b *ExtendedReader) ReadHuffman() ([]byte, error) {
	var name []byte
	maxCharLen, err := b.ReadUint32BE()
	if err != nil {
		return name, err
	}
	sizeInBits, err := b.ReadUint32BECompressed(true)
	if err != nil {
		return name, err
	}

	name = make([]byte, maxCharLen)
//	fmt.Printf("Going for huffman: %X bits, %X chars\n", sizeInBits, maxCharLen)
	if maxCharLen > 0x5000 || sizeInBits > 0x5000 {
		return name, errors.New("sanity check: exceeded maximum sizeinbits/maxcharlen of 0x1000")
	}
	englishTree.DecodeArray(b, uint(sizeInBits), uint(maxCharLen), name)
	//fmt.Printf("Done with huffman: %s\n", name)

	return name, nil
}

func (b *ExtendedReader) ReadUint8() (uint8, error) {
	res, err := b.Bits(8)
	return uint8(res), err
}

func (b *ExtendedReader) ReadString(length int) ([]byte, error) {
	dest := make([]byte, length)
	err := b.Bytes(dest, length)
	return dest, err
}

func (b *ExtendedReader) ReadLengthAndString() (string, error) {
	var ret []byte
	thisLen, err := b.ReadUint16BE()
	if err != nil {
		return "", err
	}
	b.Align()
	ret, err = b.ReadString(int(thisLen))
	return string(ret), err
}

func (b *ExtendedReader) ReadASCII(length int) (string, error) {
	res, err := b.ReadString(length)
	return string(res), err
}

func (b *ExtendedReader) ReadAddress() (*net.UDPAddr, error) {
	version, err := b.ReadUint8()
	if err != nil {
		return nil, err
	}
	if version != 4 {
		return nil, errors.New("Unsupported version")
	}
	var address net.IP = make([]byte, 4)
	err = b.Bytes(address, 4)
	if err != nil {
		return nil, err
	}
	port, err := b.ReadUint16BE()
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{address, int(port), ""}, nil
}

func (b *ExtendedReader) ReadFloat32BE() (float32, error) {
	intf, err := b.ReadUint32BE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
}

func (b *ExtendedReader) ReadFloat64BE() (float64, error) {
	intf, err := b.Bits(64)
	if err != nil {
		return 0.0, err
	}
	return math.Float64frombits(intf), err
}

func (b *ExtendedReader) RegionToGZipStream() (*ExtendedReader, error) {
	compressedLen, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}

	compressed := make([]byte, compressedLen)
	err = b.Bytes(compressed, int(compressedLen))
	if err != nil {
		return nil, err
	}

	gzipStream, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	return &ExtendedReader{bitstream.NewReader(gzipStream)}, err
}

func (b *ExtendedReader) ReadJoinReferent() (string, uint32, error) {
	stringLen, err := b.ReadUint8()
	if err != nil {
		return "", 0, err
	}
	ref := "NULL"
	if stringLen != 0xFF {
		ref, err = b.ReadASCII(int(stringLen))
		if err != nil {
			return "", 0, err
		}
	}

	intVal, err := b.ReadUint32LE()
	if err != nil {
		return "", 0, err
	}

	return ref, intVal, nil
}
