package peer

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"

	"github.com/DataDog/zstd"
	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

// TODO: Move extendedReader to its own package (roblox-dissector/parser)?
type extendedReader struct {
	r io.Reader
}

func (b *extendedReader) ReadByte() (byte, error) {
	var byt [1]byte
	n, err := b.r.Read(byt[:])
	if n != 1 {
		return 0, err
	}
	return byt[0], nil
}
func (b *extendedReader) Read(dest []byte) (int, error) {
	return b.r.Read(dest)
}

func (b *extendedReader) bytes(dest []byte, length int) error {
	n, err := b.Read(dest[:length])
	if n != length {
		return err
	}
	return nil
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
	res, err := b.ReadByte()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return false, err
	case 1:
		return true, err
	default:
		return false, errors.New("invalid bool byte")
	}
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

func (b *extendedReader) readUint8() (uint8, error) {
	res, err := b.ReadByte()
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
	intf, err := b.readUint64BE()
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

	return &extendedReader{gzipStream}, err
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
	return &extendedReader{zstdStream}, nil
}

func (b *extendedReader) readObjectPeerID(context *CommunicationContext) (datamodel.Reference, error) {
	ref := datamodel.Reference{}
	peerID, err := b.readVarint64()
	if err != nil {
		return ref, err
	}
	ref.PeerId = uint32(peerID)
	if peerID == 0 {
		ref.IsNull = true
		ref.Scope = "null"
		return ref, err
		// Approximately reflects handling in client?
	} else if uint32(peerID) == context.ServerPeerID {
		ref.Scope = "RBXServer"
	} else {
		ref.Scope = fmt.Sprintf("RBXPID%d", peerID)
	}
	ref.Id, err = b.readUint32LE()

	return ref, nil
}

func (b *extendedReader) readJoinObject(context *CommunicationContext) (datamodel.Reference, error) {
	ref := datamodel.Reference{}
	if context.ServerPeerID == 0 {
		// read scope using old system
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
			if err != nil {
				return ref, err
			}
			if len(refString) != 0x23 {
				return ref, errors.New("wrong scope len")
			}
			ref.Scope = refString
		} else {
			ref.Scope = context.InstanceTopScope
		}
		ref.Id, err = b.readUint32LE()
		return ref, err
	}
	return b.readObjectPeerID(context)
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

// ErrCacheReadOOB is an error specifying that there was a cache
// miss, i.e. a cached used before it was initialized
var ErrCacheReadOOB = errors.New("cache read is out of bounds")

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
		return "", ErrCacheReadOOB
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

func (b *extendedReader) readScope() (interface{}, error) {
	stringLen, err := b.readUint32BE()
	if err != nil {
		return "", err
	}
	if stringLen != 0x23 {
		return "", errors.New("invalid scope len")
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
	thisString, err := b.readWithCache(cache, (*extendedReader).readScope)
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

func shuffleSlice(src []byte) []byte {
	ShuffledSrc := make([]byte, 0, len(src))
	ShuffledSrc = append(ShuffledSrc, src[:0x10]...)
	for j := len(src) - 0x10; j >= 0x10; j -= 0x10 {
		ShuffledSrc = append(ShuffledSrc, src[j:j+0x10]...)
	}
	return ShuffledSrc
}

func calculateChecksum(data []byte) uint32 {
	var sum uint32
	var r uint16 = 55665
	var c1 uint16 = 52845
	var c2 uint16 = 22719
	for i := 0; i < len(data); i++ {
		char := data[i]
		cipher := (char ^ byte(r>>8)) & 0xFF
		r = (uint16(cipher)+r)*c1 + c2
		sum += uint32(cipher)
	}
	return sum
}

var packet90AESKey = [...]byte{0xFE, 0xF9, 0xF0, 0xEB, 0xE2, 0xDD, 0xD4, 0xCF, 0xC6, 0xC1, 0xB8, 0xB3, 0xAA, 0xA5, 0x9C, 0x97}

func (b *extendedReader) aesDecrypt(lenBytes int, key [0x10]byte) (*extendedReader, error) {
	data, err := b.readString(lenBytes)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	dest := make([]byte, len(data))
	// "Invalid" initialization for debug
	for i := range dest {
		dest[i] = 0xBA
	}
	c := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	ShuffledSrc := shuffleSlice(data)

	c.CryptBlocks(dest, ShuffledSrc)
	dest = shuffleSlice(dest)

	checkSum := calculateChecksum(dest[4:])
	thisStream := &extendedReader{bytes.NewReader(dest)}
	storedChecksum, err := thisStream.readUint32LE()
	if err != nil {
		return thisStream, err
	}
	if storedChecksum != checkSum {
		println("checksum check failed!", storedChecksum, checkSum)
		return thisStream, errors.New("checksum check fail")
	}

	_, err = thisStream.ReadByte()
	if err != nil {
		return thisStream, err
	}
	paddingSizeByte, err := thisStream.ReadByte()
	if err != nil {
		return thisStream, err
	}
	PaddingSize := paddingSizeByte & 0xF

	void := make([]byte, PaddingSize)
	err = thisStream.bytes(void, int(PaddingSize))
	if err != nil {
		return thisStream, err
	}

	return thisStream, nil
}

// RakNetFlags contains a set of flags which outline basic
// information about a RakNet layer packet
type RakNetFlags struct {
	// IsValid specifies whether the packet can be considered valid.
	IsValid bool
	// IsACK specifies whether the packet is an acknowledgement packet
	IsACK bool
	// IsNAK specifies whether the packet is a not-acknowledged packet
	IsNAK            bool
	IsPacketPair     bool
	IsContinuousSend bool
	NeedsBAndAS      bool
	HasBAndAS        bool
}

func (b *extendedReader) readRakNetFlags() (RakNetFlags, error) {
	val := RakNetFlags{}
	flags, err := b.ReadByte()
	if err != nil {
		return val, err
	}
	val.IsValid = flags>>7&1 == 1
	val.IsACK = flags>>6&1 == 1
	if val.IsACK {
		val.HasBAndAS = flags>>5&1 == 1
		return val, nil
	}
	val.IsNAK = flags>>5&1 == 1
	if val.IsNAK {
		val.HasBAndAS = flags>>4&1 == 1
		return val, nil
	}
	val.IsPacketPair = flags>>4&1 == 1
	val.IsContinuousSend = flags>>3&1 == 1
	val.NeedsBAndAS = flags>>2&1 == 1
	return val, nil
}

func (b *extendedReader) readReliabilityFlags() (uint8, bool, error) {
	flags, err := b.ReadByte()
	return flags >> 5, flags>>4&1 == 1, err
}

type deferredStrings struct {
	m                    map[string][]*datamodel.ValueDeferredString
	underlyingDictionary map[string]rbxfile.ValueSharedString
}

func newDeferredStrings(reader PacketReader) deferredStrings {
	return deferredStrings{
		m:                    make(map[string][]*datamodel.ValueDeferredString),
		underlyingDictionary: reader.SharedStrings(),
	}
}

func (m deferredStrings) NewValue(md5 string) *datamodel.ValueDeferredString {
	val := &datamodel.ValueDeferredString{Hash: md5}
	if resolved, ok := m.underlyingDictionary[md5]; ok {
		val.Value = resolved
		return val
	}

	// doesn't matter if it's nil, we need to update it in the map anyway
	m.m[md5] = append(m.m[md5], val)
	return val
}

func (m deferredStrings) Resolve(md5 string, value rbxfile.ValueSharedString) {
	for _, resolvable := range m.m[md5] {
		resolvable.Value = value
	}
}

type writeDeferredStrings struct {
	m                    map[string]rbxfile.ValueSharedString
	underlyingDictionary map[string]rbxfile.ValueSharedString
}

func (b *extendedReader) resolveDeferredStrings(defers deferredStrings) error {
	for len(defers.m) > 0 {
		md5, err := b.readASCII(0x10)
		if err != nil {
			return err
		}
		format, err := b.ReadByte()
		if err != nil {
			return err
		}

		_, isValid := defers.m[md5]
		if !isValid {
			return fmt.Errorf("didn't defer with md5 = %X", md5)
		}

		cachedValue, cachedExists := defers.underlyingDictionary[md5]
		switch format {
		case 0:
			if cachedExists {
				return fmt.Errorf("duplicated deferred resolve for %X", md5)
			}
			resolvedValue, err := b.readVarLengthString()
			if err != nil {
				return err
			}
			defers.Resolve(md5, rbxfile.ValueSharedString(resolvedValue))
			delete(defers.m, md5)

			defers.underlyingDictionary[md5] = rbxfile.ValueSharedString(resolvedValue)
		case 1:
			if !cachedExists {
				return fmt.Errorf("couldn't resolve deferred %X", md5)
			}
			defers.Resolve(md5, cachedValue)
			delete(defers.m, md5)
		default:
			return errors.New("invalid deferred string dictionary format")
		}
	}
	return nil
}
