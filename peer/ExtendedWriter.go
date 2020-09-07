package peer

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"

	"github.com/DataDog/zstd"
	"github.com/robloxapi/rbxfile"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

type extendedWriter struct {
	w io.Writer
}

func (b *extendedWriter) bytes(length int, value []byte) error {
	if length > len(value) {
		return errors.New("buffer overflow")
	}

	n, err := b.w.Write(value[:length])
	if n != length {
		return err
	}
	return nil
}

func (b *extendedWriter) allBytes(value []byte) error {
	return b.bytes(len(value), value)
}

func (b *extendedWriter) WriteByte(value byte) error {
	n, err := b.w.Write([]byte{value})
	if n != 1 {
		return err
	}
	return nil
}
func (b *extendedWriter) Write(value []byte) (int, error) {
	return b.w.Write(value)
}

func (b *extendedWriter) writeUint16BE(value uint16) error {
	dest := make([]byte, 2)
	binary.BigEndian.PutUint16(dest, value)
	return b.bytes(2, dest)
}
func (b *extendedWriter) writeUint32BE(value uint32) error {
	dest := make([]byte, 4)
	binary.BigEndian.PutUint32(dest, value)
	return b.bytes(4, dest)
}
func (b *extendedWriter) writeUint32LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.bytes(4, dest)
}

func (b *extendedWriter) writeUint64BE(value uint64) error {
	dest := make([]byte, 8)
	binary.BigEndian.PutUint64(dest, value)
	return b.bytes(8, dest)
}

func (b *extendedWriter) writeUint24LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.bytes(3, dest)
}

func (b *extendedWriter) writeFloat32BE(value float32) error {
	return b.writeUint32BE(math.Float32bits(value))
}

func (b *extendedWriter) writeFloat64BE(value float64) error {
	return b.writeUint64BE(math.Float64bits(value))
}

func (b *extendedWriter) writeBoolByte(value bool) error {
	if value {
		return b.WriteByte(1)
	}
	return b.WriteByte(0)
}

func (b *extendedWriter) writeAddress(value *net.UDPAddr) error {
	err := b.WriteByte(4)
	if err != nil {
		return err
	}
	for i := 0; i < len(value.IP); i++ {
		value.IP[i] = value.IP[i] ^ 0xFF // bitwise NOT
	}
	err = b.bytes(4, value.IP[len(value.IP)-4:])
	if err != nil {
		return err
	}

	// in case the value will be used again
	for i := 0; i < len(value.IP); i++ {
		value.IP[i] = value.IP[i] ^ 0xFF
	}

	err = b.writeUint16BE(uint16(value.Port))
	return err
}

func (b *extendedWriter) writeASCII(value string) error {
	return b.allBytes([]byte(value))
}

func (b *extendedWriter) writeUintUTF8(value uint32) error {
	if value == 0 {
		return b.WriteByte(0)
	}
	for value != 0 {
		nextValue := value >> 7
		if nextValue != 0 {
			err := b.WriteByte(byte(value&0x7F | 0x80))
			if err != nil {
				return err
			}
		} else {
			err := b.WriteByte(byte(value & 0x7F))
			if err != nil {
				return err
			}
		}
		value = nextValue
	}
	return nil
}

type cacheWriteCallback func(*extendedWriter, interface{}) error

func (b *extendedWriter) writeWithCache(value interface{}, cache Cache, writeCallback cacheWriteCallback) error {
	if value == nil {
		return b.WriteByte(0x00)
	}
	var matchedIndex byte
	var i byte
	for i = 0; i < 0x80; i++ {
		equal, existed := cache.Equal(i, value)
		if !existed {
			break
		}
		if equal {
			matchedIndex = i
			break
		}
	}
	if matchedIndex == 0 {
		cache.Put(value, cache.LastWrite()%0x7F+1)
		err := b.WriteByte(cache.LastWrite() | 0x80)
		if err != nil {
			return err
		}
		return writeCallback(b, value)
	}
	return b.WriteByte(matchedIndex)
}

func (b *extendedWriter) writeUint32AndString(val interface{}) error {
	str := val.(string)
	err := b.writeUint32BE(uint32(len(str)))
	if err != nil {
		return err
	}
	return b.writeASCII(str)
}

func (b *extendedWriter) writeCached(val string, caches *Caches) error {
	cache := &caches.String

	return b.writeWithCache(val, cache, (*extendedWriter).writeUint32AndString)
}
func (b *extendedWriter) writeCachedObject(val string, caches *Caches) error {
	cache := &caches.Object

	return b.writeWithCache(val, cache, (*extendedWriter).writeUint32AndString)
}

func (b *extendedWriter) writeNewCachedProtectedString(val []byte, caches *Caches) error {
	cache := &caches.ProtectedString

	return b.writeWithCache(val, cache, func(b *extendedWriter, val interface{}) error {
		str := val.([]byte)
		err := b.writeUint32BE(uint32(len(str)))
		if err != nil {
			return err
		}
		return b.allBytes(val.([]byte))
	})
}

func (b *extendedWriter) writeJoinRef(ref datamodel.Reference, context *CommunicationContext) error {
	var err error
	if ref.IsNull {
		err = b.WriteByte(0)
		return err
	}
	if ref.Scope == context.InstanceTopScope {
		err = b.WriteByte(0xFF)
	} else {
		err = b.WriteByte(uint8(len(ref.Scope)))
		if err != nil {
			return err
		}
		err = b.writeASCII(ref.Scope)
	}
	if err != nil {
		return err
	}

	return b.writeUint32LE(ref.Id)
}
func (b *extendedWriter) writeJoinObject(object *datamodel.Instance, context *CommunicationContext) error {
	if context.ServerPeerID != 0 {
		if object == nil {
			// Yes, I know it's equivalent to WriteByte(0)
			// I think this is more clear though
			return b.writeVarint64(0)
		}
		return b.writeRefPeerID(object.Ref, context)
	}
	if object == nil {
		return b.WriteByte(0)
	}
	return b.writeJoinRef(object.Ref, context)
}

func (b *extendedWriter) writeRef(ref datamodel.Reference, caches *Caches) error {
	if ref.IsNull {
		return b.WriteByte(0)
	}
	err := b.writeCachedObject(ref.Scope, caches)
	if err != nil {
		return err
	}
	return b.writeUint32LE(ref.Id)
}

// TODO: Implement a similar system for readers, where it simply returns an instance
func (b *extendedWriter) writeObject(object *datamodel.Instance, context *CommunicationContext, caches *Caches) error {
	if context.ServerPeerID != 0 {
		if object == nil {
			// Yes, I know it's equivalent to WriteByte(0)
			// I think this is more clear though
			return b.writeVarint64(0)
		}
		return b.writeRefPeerID(object.Ref, context)
	}
	if object == nil {
		return b.WriteByte(0)
	}
	return b.writeRef(object.Ref, caches)
}
func (b *extendedWriter) writeRefPeerID(ref datamodel.Reference, context *CommunicationContext) error {
	if ref.IsNull {
		return b.writeVarint64(0)
	}
	err := b.writeVarint64(uint64(ref.PeerId))
	if err != nil {
		return err
	}
	return b.writeUint32LE(ref.Id)
}

type aesExtendedWriter struct {
	*extendedWriter
	buffer       *bytes.Buffer
	targetStream io.Writer
	key          [0x10]byte
}

func (b *aesExtendedWriter) Close() error {
	rawBuffer := b.buffer

	// create slice with 6-byte header and padding to align to 0x10-byte blocks
	length := rawBuffer.Len()
	paddingSize := 0xF - (length+5)%0x10
	rawCopy := make([]byte, length+6+paddingSize)
	rawCopy[5] = byte(paddingSize & 0xF)
	copy(rawCopy[6+paddingSize:], rawBuffer.Bytes())

	checkSum := calculateChecksum(rawCopy[4:])
	rawCopy[3] = byte(checkSum >> 24 & 0xFF)
	rawCopy[2] = byte(checkSum >> 16 & 0xFF)
	rawCopy[1] = byte(checkSum >> 8 & 0xFF)
	rawCopy[0] = byte(checkSum & 0xFF)

	// CBC blocks are encrypted in a weird order
	dest := make([]byte, len(rawCopy))
	shuffledEncryptable := shuffleSlice(rawCopy)
	block, err := aes.NewCipher(b.key[:])
	if err != nil {
		return err
	}
	c := cipher.NewCBCEncrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	c.CryptBlocks(dest, shuffledEncryptable)
	dest = shuffleSlice(dest) // shuffle back to correct order

	_, err = b.targetStream.Write(dest)
	return err
}

func (b *extendedWriter) aesEncrypt(key [0x10]byte) *aesExtendedWriter {
	rawBuffer := new(bytes.Buffer)
	rawStream := &extendedWriter{rawBuffer}

	return &aesExtendedWriter{
		extendedWriter: rawStream,
		targetStream:   b,
		buffer:         rawBuffer,
		key:            key,
	}
}

func (b *extendedWriter) writeRakNetFlags(val RakNetFlags) error {
	var flags byte
	if val.IsValid {
		flags |= 1 << 7
	}
	if val.IsACK {
		flags |= 1 << 6
		if val.HasBAndAS {
			flags |= 1 << 5
		}
		return b.WriteByte(flags)
	}
	if val.IsNAK {
		flags |= 1 << 5
		if val.HasBAndAS {
			flags |= 1 << 4
		}
		return b.WriteByte(flags)
	}
	if val.IsPacketPair {
		flags |= 1 << 4
	}
	if val.IsContinuousSend {
		flags |= 1 << 3
	}
	if val.NeedsBAndAS {
		flags |= 1 << 2
	}
	return b.WriteByte(flags)
}

func (b *extendedWriter) writeReliabilityFlags(rel uint8, hasSplit bool) error {
	flags := rel << 5
	if hasSplit {
		flags |= 1 << 4
	}
	return b.WriteByte(flags)
}

type zstdExtendedWriter struct {
	*extendedWriter
	compressedBuffer *bytes.Buffer
	compressor       *zstd.Writer
	counter          *countWriter
	targetStream     *extendedWriter
	closed           bool
}

// Close flushes and closes the compression stream.
// Trying to call this method on a closed stream will do nothing.
func (b *zstdExtendedWriter) Close() error {
	// double closing a zstd.Writer will result in flaky memory management
	if b.closed {
		return nil
	}
	b.closed = true

	err := b.compressor.Close()
	if err != nil {
		return err
	}

	err = b.targetStream.writeUint32BE(uint32(b.compressedBuffer.Len()))
	if err != nil {
		return err
	}
	err = b.targetStream.writeUint32BE(uint32(b.counter.numBytes))
	if err != nil {
		return err
	}

	return b.targetStream.allBytes(b.compressedBuffer.Bytes())
}

func (b *extendedWriter) wrapZstd() *zstdExtendedWriter {
	compressedBuffer := new(bytes.Buffer)
	compressor := zstd.NewWriter(compressedBuffer)
	counter := newCountWriter()
	writeMux := io.MultiWriter(compressor, counter)
	return &zstdExtendedWriter{
		extendedWriter:   &extendedWriter{writeMux},
		targetStream:     b,
		compressedBuffer: compressedBuffer,
		counter:          counter,
		compressor:       compressor,
	}
}

func newWriteDeferredStrings(writer PacketWriter) writeDeferredStrings {
	return writeDeferredStrings{
		m:                    make(map[string]rbxfile.ValueSharedString),
		underlyingDictionary: writer.SharedStrings(),
	}
}

func (m writeDeferredStrings) Defer(value *datamodel.ValueDeferredString) {
	// if previously sent in *another packet*, we're not expected to defer it
	if _, ok := m.underlyingDictionary[value.Hash]; ok {
		return
	}
	// if duplicated within the same packet, overwritten -- doesn't matter
	// otherwise just add it to the defers
	m.m[value.Hash] = value.Value
}

func (b *extendedWriter) resolveDeferredStrings(defers writeDeferredStrings) error {
	var err error
	for md5, value := range defers.m {
		if len(md5) != 0x10 {
			return errors.New("invalid md5")
		}
		err = b.writeASCII(md5)
		if err != nil {
			return err
		}
		if _, ok := defers.underlyingDictionary[md5]; ok {
			// the same sharedstring has been sent in the same packet before
			err = b.WriteByte(1)
			if err != nil {
				return err
			}
			continue
		}

		err = b.WriteByte(0)
		if err != nil {
			return err
		}
		err = b.writeUintUTF8(uint32(len(value)))
		if err != nil {
			return err
		}
		err = b.allBytes(value)
		if err != nil {
			return err
		}

		// Deduplicate shared string on subsequent sends
		defers.underlyingDictionary[md5] = value
	}
	return nil
}
