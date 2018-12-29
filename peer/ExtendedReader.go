package peer

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/DataDog/zstd"
	bitstream "github.com/gskartwii/go-bitstream"
	"github.com/gskartwii/roblox-dissector/datamodel"
)

var englishTree *huffmanEncodingTree

func init() {
	englishTree = generateHuffmanFromFrequencyTable([]uint32{
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		722,
		0,
		0,
		2,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		11084,
		58,
		63,
		1,
		0,
		31,
		0,
		317,
		64,
		64,
		44,
		0,
		695,
		62,
		980,
		266,
		69,
		67,
		56,
		7,
		73,
		3,
		14,
		2,
		69,
		1,
		167,
		9,
		1,
		2,
		25,
		94,
		0,
		195,
		139,
		34,
		96,
		48,
		103,
		56,
		125,
		653,
		21,
		5,
		23,
		64,
		85,
		44,
		34,
		7,
		92,
		76,
		147,
		12,
		14,
		57,
		15,
		39,
		15,
		1,
		1,
		1,
		2,
		3,
		0,
		3611,
		845,
		1077,
		1884,
		5870,
		841,
		1057,
		2501,
		3212,
		164,
		531,
		2019,
		1330,
		3056,
		4037,
		848,
		47,
		2586,
		2919,
		4771,
		1707,
		535,
		1106,
		152,
		1243,
		100,
		0,
		2,
		0,
		10,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0})
}

// TODO: Move extendedReader to its own package (roblox-dissector/bitstreams)?
type extendedReader struct {
	*bitstream.BitReader
}

func (b *extendedReader) bits(len int) (uint64, error) {
	return b.ReadBits(len)
}

func (b *extendedReader) bytes(dest []byte, len int) error {
	var Byte byte
	for i := 0; i < len; i++ {
		res, err := b.bits(8)
		if err != nil {
			return err
		}
		Byte = byte(res)
		dest[i] = Byte
	}
	return nil
}

func (b *extendedReader) readBool() (bool, error) {
	res, err := b.ReadBit()
	return bool(res), err
}

func (b *extendedReader) readUint16BE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.bytes(dest, 2)
	return binary.BigEndian.Uint16(dest), err
}

func (b *extendedReader) readUint16LE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.bytes(dest, 2)
	return binary.LittleEndian.Uint16(dest), err
}

func (b *extendedReader) readBoolByte() (bool, error) {
	res, err := b.bits(8)
	return res == 1, err
}

func (b *extendedReader) readUint24LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 3)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *extendedReader) readUint32BE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 4)
	return binary.BigEndian.Uint32(dest), err
}
func (b *extendedReader) readUint32LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 4)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *extendedReader) readUint64BE() (uint64, error) {
	dest := make([]byte, 8)
	err := b.bytes(dest, 8)
	return binary.BigEndian.Uint64(dest), err
}

func (b *extendedReader) readCompressed(dest []byte, size uint32, unsignedData bool) error {
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
		res, err := b.readBool()
		if err != nil {
			return err
		}
		if res {
			dest[currentByte] = byteMatch
			currentByte--
		} else {
			err = b.bytes(dest, int(currentByte+1))
			return err
		}
	}

	res, err := b.readBool()
	if err != nil {
		return err
	}

	if res {
		res, err := b.bits(4)
		if err != nil {
			return err
		}
		dest[currentByte] = byte(res) | halfByteMatch
	} else {
		err := b.bytes(dest[currentByte:], 1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedReader) readUint32BECompressed(unsignedData bool) (uint32, error) {
	dest := make([]byte, 4)
	err := b.readCompressed(dest, 32, unsignedData)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(dest), nil
}

func (b *extendedReader) readHuffman() ([]byte, error) {
	var name []byte
	maxCharLen, err := b.readUint32BE()
	if err != nil {
		return name, err
	}
	sizeInBits, err := b.readUint32BECompressed(true)
	if err != nil {
		return name, err
	}

	if maxCharLen > 0x5000 || sizeInBits > 0x50000 {
		return name, errors.New("sanity check: exceeded maximum sizeinbits/maxcharlen of 0x5000")
	}
	name = make([]byte, maxCharLen)
	err = englishTree.decodeArray(b, uint(sizeInBits), uint(maxCharLen), name)

	return name, err
}

func (b *extendedReader) readUint8() (uint8, error) {
	res, err := b.bits(8)
	return uint8(res), err
}

func (b *extendedReader) readString(length int) ([]byte, error) {
	var dest []byte
	if uint(length) > 0x1000000 {
		return dest, errors.New("Sanity check: string too long")
	}
	dest = make([]byte, length)
	err := b.bytes(dest, length)
	return dest, err
}

func (b *extendedReader) readLengthAndString() (string, error) {
	var ret []byte
	thisLen, err := b.readUint16BE()
	if err != nil {
		return "", err
	}
	b.Align()
	ret, err = b.readString(int(thisLen))
	return string(ret), err
}

func (b *extendedReader) readASCII(length int) (string, error) {
	res, err := b.readString(length)
	return string(res), err
}

func (b *extendedReader) readAddress() (*net.UDPAddr, error) {
	version, err := b.readUint8()
	if err != nil {
		return nil, err
	}
	if version != 4 {
		return nil, errors.New("Unsupported version")
	}
	var address net.IP = make([]byte, 4)
	err = b.bytes(address, 4)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 4; i++ {
		address[i] = address[i] ^ 0xFF // bitwise NOT
	}
	port, err := b.readUint16BE()
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{IP: address, Port: int(port)}, nil
}

func (b *extendedReader) readFloat32LE() (float32, error) {
	intf, err := b.readUint32LE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
}

func (b *extendedReader) readFloat32BE() (float32, error) {
	intf, err := b.readUint32BE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
}

func (b *extendedReader) readFloat64BE() (float64, error) {
	intf, err := b.bits(64)
	if err != nil {
		return 0.0, err
	}
	return math.Float64frombits(intf), err
}

func (b *extendedReader) RegionToGZipStream() (*extendedReader, error) {
	compressedLen, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}
	println("compressedLen:", compressedLen)

	compressed := make([]byte, compressedLen)
	err = b.bytes(compressed, int(compressedLen))
	if err != nil {
		return nil, err
	}
	fmt.Printf("compressed start: %v\n", compressed[:0x20])

	gzipStream, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	return &extendedReader{bitstream.NewReader(gzipStream)}, err
}

func (b *extendedReader) RegionToZStdStream() (*extendedReader, error) {
	compressedLen, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}

	_, err = b.readUint32BE()
	if err != nil {
		return nil, err
	}

	compressed := make([]byte, compressedLen)
	err = b.bytes(compressed, int(compressedLen))
	if err != nil {
		return nil, err
	}
	/*decompressed, err := zstd.Decompress(nil, compressed)
	println("decomp len", len(decompressed))
	if len(decompressed) > 0x20 {
		fmt.Printf("first bytes %#X\n", decompressed[:0x20])
	} else {
		fmt.Printf("first bytes %#X\n", decompressed)
	}*/

	zstdStream := zstd.NewReader(bytes.NewReader(compressed))
	return &extendedReader{bitstream.NewReader(zstdStream)}, nil
}

func (b *extendedReader) readJoinObject(context *CommunicationContext) (datamodel.Reference, error) {
	ref := datamodel.Reference{}
	stringLen, err := b.readUint8()
	if err != nil {
		return ref, err
	}
	if stringLen == 0x00 {
		ref.IsNull = true
		ref.Scope = "null"
		return ref, err
	}
	var refString string
	if stringLen != 0xFF {
		refString, err = b.readASCII(int(stringLen))
		if len(refString) != 0x23 {
			println("WARN: wrong ref len!! this should never happen, unless you are communicating with a non-standard peer")
		}
		if err != nil {
			return ref, err
		}
		ref.Scope = refString
	} else {
		ref.Scope = context.InstanceTopScope
	}

	ref.Id, err = b.readUint32LE()
	return ref, err
}

func (b *extendedReader) readFloat16BE(floatMin float32, floatMax float32) (float32, error) {
	scale, err := b.readUint16BE()
	if err != nil {
		return 0.0, err
	}

	outFloat := float32(scale)/65535.0*(floatMax-floatMin) + floatMin
	if outFloat < floatMin {
		return floatMin, nil
	} else if outFloat > floatMax {
		return floatMax, nil
	}
	return outFloat, nil
}

type cacheReadCallback func(*extendedReader) (interface{}, error)

var CacheReadOOB = errors.New("Cache read is out of bounds")

func (b *extendedReader) readWithCache(cache Cache, readCallback cacheReadCallback) (interface{}, error) {
	var result interface{}
	var err error
	cacheIndex, err := b.readUint8()
	if err != nil {
		return "", err
	}
	if cacheIndex == 0x00 {
		return "NULL", nil
	}

	if cacheIndex < 0x80 {
		result, _ = cache.Get(cacheIndex)
	} else {
		result, err = readCallback(b)
		if err != nil {
			return "", err
		}
		cache.Put(result, cacheIndex&0x7F)
	}

	if result == nil {
		return "", CacheReadOOB
	}

	return result, err
}

func (b *extendedReader) readUint32AndString() (interface{}, error) {
	stringLen, err := b.readUint32BE()
	if err != nil {
		return "", err
	}
	return b.readASCII(int(stringLen))
}

// TODO: Perhaps make readWithCache() operate with a member function of the cache instead?
func (b *extendedReader) readCached(caches *Caches) (string, error) {
	cache := &caches.String

	thisString, err := b.readWithCache(cache, (*extendedReader).readUint32AndString)
	return thisString.(string), err
}

func (b *extendedReader) readCachedScope(caches *Caches) (string, error) {
	cache := &caches.Object
	thisString, err := b.readWithCache(cache, (*extendedReader).readUint32AndString)
	return thisString.(string), err
}

func (b *extendedReader) readCachedContent(caches *Caches) (string, error) {
	cache := &caches.Content

	thisString, err := b.readWithCache(cache, (*extendedReader).readUint32AndString)
	return thisString.(string), err
}

func (b *extendedReader) readNewCachedProtectedString(caches *Caches) ([]byte, error) {
	cache := &caches.ProtectedString

	thisString, err := b.readWithCache(cache, func(b *extendedReader) (interface{}, error) {
		stringLen, err := b.readUint32BE()
		if err != nil {
			return []byte{}, err
		}
		thisString, err := b.readString(int(stringLen))
		return thisString, err
	})
	if _, ok := thisString.(string); ok {
		return nil, err
	}
	return thisString.([]byte), err
}
