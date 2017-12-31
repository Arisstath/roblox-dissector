package peer
import "github.com/gskartwii/go-bitstream"
import "encoding/binary"
import "net"
import "github.com/gskartwii/rbxfile"
import "math"
import "bytes"

type extendedWriter struct {
	*bitstream.BitWriter
}

func (b *extendedWriter) bits(len int, value uint64) error {
	return b.WriteBits(value, len)
}

func (b *extendedWriter) bytes(len int, value []byte) error {
	for i := 0; i < len; i++ {
		err := b.WriteByte(value[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedWriter) allBytes(value []byte) error {
	return b.bytes(len(value), value)
}

func (b *extendedWriter) writeBool(value bool) error {
	return b.WriteBit(bitstream.Bit(value))
}

func (b *extendedWriter) writeUint16BE(value uint16) error {
	dest := make([]byte, 2)
	binary.BigEndian.PutUint16(dest, value)
	return b.bytes(2, dest)
}
func (b *extendedWriter) writeUint16LE(value uint16) error {
	dest := make([]byte, 2)
	binary.LittleEndian.PutUint16(dest, value)
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
	return b.bits(64, math.Float64bits(value))
}

func (b *extendedWriter) writeFloat16BE(value float32, min float32, max float32) error {
	return b.writeUint16BE(uint16(value / (max - min) * 65535.0))
}

func (b *extendedWriter) writeBoolByte(value bool) error {
	if value {
		return b.WriteByte(1)
	} else {
		return b.WriteByte(0)
	}
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
		print("new round because", value)
        nextValue := value >> 7
        if nextValue != 0 {
            err := b.WriteByte(byte(value&0x7F|0x80))
            if err != nil {
                return err
            }
        } else {
            err := b.WriteByte(byte(value&0x7F))
            if err != nil {
                return err
            }
        }
        value = nextValue
    }
    return nil
}

type CacheWriteCallback func(*extendedWriter, interface{})(error)
func (b *extendedWriter) writeWithCache(value interface{}, cache Cache, writeCallback CacheWriteCallback) error {
	if value == nil {
		return b.WriteByte(0x00)
	}
	var matchedIndex byte = 0
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
		cache.Put(value, cache.LastWrite() % 0x7F + 1)
		println("Writing new value to cache: ", cache.LastWrite())
		err := b.WriteByte(cache.LastWrite() | 0x80)
		if err != nil {
			return err
		}
		return writeCallback(b, value)
	} else {
		return b.WriteByte(matchedIndex)
	}
}

func (b *extendedWriter) writeUint32AndString(val interface{}) error {
	str := val.(string)
	err := b.writeUint32BE(uint32(len(str)))
	if err != nil {
		return err
	}
	return b.writeASCII(str)
}

func (b *extendedWriter) writeCached(isClient bool, val string, context *CommunicationContext) error {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.String
	} else {
		cache = &context.ServerCaches.String
	}

	return b.writeWithCache(val, cache, (*extendedWriter).writeUint32AndString)
}
func (b *extendedWriter) writeCachedObject(isClient bool, val string, context *CommunicationContext) error {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.Object
	} else {
		cache = &context.ServerCaches.Object
	}

	return b.writeWithCache(val, cache, (*extendedWriter).writeUint32AndString)
}
func (b *extendedWriter) writeCachedContent(isClient bool, val string, context *CommunicationContext) error {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.Content
	} else {
		cache = &context.ServerCaches.Content
	}

	return b.writeWithCache(val, cache, (*extendedWriter).writeUint32AndString)
}
func (b *extendedWriter) writeNewCachedProtectedString(isClient bool, val string, context *CommunicationContext) error {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.ProtectedString
	} else {
		cache = &context.ServerCaches.ProtectedString
	}

	return b.writeWithCache(val, cache, func(b *extendedWriter, val interface{})(error) {
		str := val.([]byte)
		err := b.writeUint32BE(uint32(len(str)))
		if err != nil {
			return err
		}
		return b.allBytes(val.([]byte))
	})
}
func (b *extendedWriter) writeCachedSystemAddress(isClient bool, val rbxfile.ValueSystemAddress, context *CommunicationContext) error {
	var cache Cache
	if isClient {
		cache = &context.ClientCaches.SystemAddress
	} else {
		cache = &context.ServerCaches.SystemAddress
	}

	return b.writeWithCache(val, cache, func(b *extendedWriter, val interface{})(error) {
		return b.writeSystemAddress(isClient, val.(rbxfile.ValueSystemAddress), true, nil)
	})
}

func (b *extendedWriter) writeObject(isClient bool, object *rbxfile.Instance, isJoinData bool, context *CommunicationContext) error {
    var err error
    if object == nil {
        return b.WriteByte(0)
    }

	referentString, referent := refToObject(object.Reference)
    if isJoinData {
        if referentString == "NULL2" {
            err = b.WriteByte(0)
            return err
		} else if referentString == context.InstanceTopScope {
            err = b.WriteByte(0xFF)
        } else {
            err = b.WriteByte(uint8(len(referentString)))
            if err != nil {
                return err
            }
            err = b.writeASCII(referentString)
		}
		if err != nil {
            return err
        }

        return b.writeUint32LE(referent)
    } else {
		if referentString == "NULL2" || referentString == "null" {
			return b.WriteByte(0x00)
		}
		err = b.writeCachedObject(isClient, referentString, context)
		if err != nil {
			return err
		}
		return b.writeUint32LE(referent)
	}
    return nil
}

func (b *extendedWriter) writeHuffman(value []byte) error {
	encodedBuffer := new(bytes.Buffer)
	encodedStream := &extendedWriter{bitstream.NewWriter(encodedBuffer)}

	bitLen, err := englishTree.encodeArray(encodedStream, value)
	if err != nil {
		return err
	}
	err = encodedStream.Flush(bitstream.Bit(false))
	if err != nil {
		return err
	}

	err = b.writeUint32BE(uint32(len(value)))
	if err != nil {
		return err
	}
	err = b.writeUint32BECompressed(uint32(bitLen))
	if err != nil {
		return err
	}
	return b.allBytes(encodedBuffer.Bytes())
}

func (b *extendedWriter) writeCompressed(value []byte, length uint32, isUnsigned bool) error {
	var byteMatch, halfByteMatch byte
	var err error
	if !isUnsigned {
		byteMatch = 0xFF
		halfByteMatch = 0xF0
	}
	var currentByte uint32
	for currentByte = length >> 3 - 1; currentByte > 0; currentByte-- {
		isMatch := value[currentByte] == byteMatch
		err = b.writeBool(isMatch)
		if err != nil {
			return err
		}
		if !isMatch {
			return b.allBytes(value[:currentByte+1])
		}
	}
	lastByte := value[0]
	if lastByte & 0xF0 == halfByteMatch {
		err = b.writeBool(true)
		if err != nil {
			return err
		}
		return b.bits(4, uint64(lastByte))
	}
	err = b.writeBool(false)
	if err != nil {
		return err
	}
	return b.WriteByte(lastByte)
}
func (b *extendedWriter) writeUint32BECompressed(value uint32) error {
	println("writing compressed val", value)
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val, value)
	return b.writeCompressed(val, 32, true)
}
