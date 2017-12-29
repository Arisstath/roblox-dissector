package peer
import "github.com/gskartwii/go-bitstream"
import "encoding/binary"
import "errors"
import "net"
import "math"
import "bytes"
import "compress/gzip"
import "strconv"
import "io"
import "fmt"
import "github.com/DataDog/zstd"

var englishTree *HuffmanEncodingTree

func init() {
	englishTree = GenerateHuffmanFromFrequencyTable([]uint32{
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
		0,})
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
	println("max char len", maxCharLen)
	if err != nil {
		return name, err
	}
	sizeInBits, err := b.ReadUint32BECompressed(true)
	println("sizeinbits", sizeInBits)
	if err != nil {
		return name, err
	}

	if maxCharLen > 0x5000 || sizeInBits > 0x50000 {
		return name, errors.New("sanity check: exceeded maximum sizeinbits/maxcharlen of 0x5000")
	}
	name = make([]byte, maxCharLen)
	err = englishTree.DecodeArray(b, uint(sizeInBits), uint(maxCharLen), name)

	return name, err
}

func (b *ExtendedReader) ReadUint8() (uint8, error) {
	res, err := b.Bits(8)
	return uint8(res), err
}

func (b *ExtendedReader) ReadString(length int) ([]byte, error) {
	var dest []byte
	if uint(length) > 0x1000000 {
		return dest, errors.New("Sanity check: string too long")
	}
	dest = make([]byte, length)
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
	for i := 0; i < 4; i++ {
		address[i] = address[i] ^ 0xFF // bitwise NOT
	}
	port, err := b.ReadUint16BE()
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{address, int(port), ""}, nil
}

func (b *ExtendedReader) ReadFloat32LE() (float32, error) {
	intf, err := b.ReadUint32LE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
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
	println("compressedLen:", compressedLen)

	compressed := make([]byte, compressedLen)
	err = b.Bytes(compressed, int(compressedLen))
	if err != nil {
		return nil, err
	}
	fmt.Printf("compressed start: %v\n", compressed[:0x20])

	gzipStream, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	return &ExtendedReader{bitstream.NewReader(gzipStream)}, err
}

func (b *ExtendedReader) RegionToZStdStream() (*ExtendedReader, error) {
	compressedLen, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}

	_, err = b.ReadUint32BE()
	if err != nil {
		return nil, err
	}

	compressed := make([]byte, compressedLen)
	err = b.Bytes(compressed, int(compressedLen))
	if err != nil {
		return nil, err
	}

	zstdStream := zstd.NewReader(bytes.NewReader(compressed))
	return &ExtendedReader{bitstream.NewReader(zstdStream)}, err
}

func (b *ExtendedReader) ReadJoinReferent(context *CommunicationContext) (string, uint32, error) {
	stringLen, err := b.ReadUint8()
	if err != nil {
		return "", 0, err
	}
	if stringLen == 0x00 {
		return "NULL2", 0, err
	}
	var ref string
	if stringLen != 0xFF {
		ref, err = b.ReadASCII(int(stringLen))
		if err != nil {
			return "", 0, err
		}
	} else {
		ref = context.InstanceTopScope
	}

	intVal, err := b.ReadUint32LE()
	if err != nil && err != io.EOF {
		return "", 0, err
	}

	return ref, intVal, nil
}

func (b *ExtendedReader) ReadFloat16BE(floatMin float32, floatMax float32) (float32, error) {
	scale, err := b.ReadUint16BE()
	if err != nil {
		return 0.0, err
	}

	outFloat := float32(scale) / 65535.0 * (floatMax - floatMin) + floatMin
	if outFloat < floatMin {
		return floatMin, nil
	} else if outFloat > floatMax {
		return floatMax, nil
	}
	return outFloat, nil
}

type CacheReadCallback func(*ExtendedReader)(interface{}, error)
func (b *ExtendedReader) readWithCache(cache Cache, readCallback CacheReadCallback) (interface{}, error) {
	var result interface{}
	var err error
	cacheIndex, err := b.ReadUint8()
	if err != nil {
		return "", err
	}
	if cacheIndex == 0x00 {
		return "", err
	}

	if cacheIndex < 0x80 {
		result, _ = cache.Get(cacheIndex)
	} else {
		result, err = readCallback(b)
		if err != nil {
			return "", err
		}
		cache.Put(result, cacheIndex & 0x7F)
	}

	if result == nil {
		return "WARN_UNASSIGNED_" + strconv.Itoa(int(cacheIndex)), nil
	}

	return result, err
}

func (b *ExtendedReader) ReadUint32AndString() (interface{}, error) {
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return "", err
	}
	return b.ReadASCII(int(stringLen))
}

func (b *ExtendedReader) ReadCached(isClient bool, context *CommunicationContext) (string, error) {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.String
	} else {
		cache = &context.ServerCaches.String
	}

	thisString, err := b.readWithCache(cache, (*ExtendedReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *ExtendedReader) ReadCachedObject(isClient bool, context *CommunicationContext) (string, error) {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.Object
	} else {
		cache = &context.ServerCaches.Object
	}

	thisString, err := b.readWithCache(cache, (*ExtendedReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *ExtendedReader) ReadCachedContent(isClient bool, context *CommunicationContext) (string, error) {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.Content
	} else {
		cache = &context.ServerCaches.Content
	}

	thisString, err := b.readWithCache(cache, (*ExtendedReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *ExtendedReader) ReadNewCachedProtectedString(isClient bool, context *CommunicationContext) ([]byte, error) {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.ProtectedString
	} else {
		cache = &context.ServerCaches.ProtectedString
	}

	thisString, err := b.readWithCache(cache, func(b *ExtendedReader)(interface{}, error) {
		stringLen, err := b.ReadUint32BE()
		if err != nil {
			return []byte{}, err
		}
		thisString, err := b.ReadString(int(stringLen))
		return thisString, err
	})
	if _, ok := thisString.(string); ok {
		return nil, err
	}
	return thisString.([]byte), err
}
