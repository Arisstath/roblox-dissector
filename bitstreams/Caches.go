package bitstreams
import "errors"
import "github.com/gskartwii/roblox-dissector/util"
import "github.com/gskartwii/rbxfile"

type cacheReadCallback func(*BitstreamReader) (interface{}, error)

var CacheReadOOB = errors.New("Cache read is out of bounds")

func (b *BitstreamReader) readWithCache(cache util.Cache, readCallback cacheReadCallback) (interface{}, error) {
	var result interface{}
	var err error
	cacheIndex, err := b.ReadUint8()
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

// TODO: Perhaps make readWithCache() operate with a member function of the cache instead?
func (b *BitstreamReader) ReadCached(caches *util.Caches) (string, error) {
	cache := &caches.String

	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadCachedScope(caches *util.Caches) (string, error) {
	cache := &caches.Object
	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadCachedContent(caches *util.Caches) (string, error) {
	cache := &caches.Content

	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadNewCachedProtectedString(caches *util.Caches) ([]byte, error) {
	cache := &caches.ProtectedString

	thisString, err := b.readWithCache(cache, func(b *BitstreamReader) (interface{}, error) {
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

type cacheWriteCallback func(*BitstreamWriter, interface{}) error

func (b *BitstreamWriter) WriteWithCache(value interface{}, cache util.Cache, writeCallback cacheWriteCallback) error {
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
		cache.Put(value, cache.LastWrite()%0x7F+1)
		err := b.WriteByte(cache.LastWrite() | 0x80)
		if err != nil {
			return err
		}
		return writeCallback(b, value)
	} else {
		return b.WriteByte(matchedIndex)
	}
}

func (b *BitstreamWriter) WriteCached(val string, caches *util.Caches) error {
	cache := &caches.String

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteCachedObject(val string, caches *util.Caches) error {
	cache := &caches.Object

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteCachedContent(val string, caches *util.Caches) error {
	cache := &caches.Content

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteNewCachedProtectedString(val []byte, caches *util.Caches) error {
	cache := &caches.ProtectedString

	return b.WriteWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		str := val.([]byte)
		err := b.WriteUint32BE(uint32(len(str)))
		if err != nil {
			return err
		}
        n, err := b.Write(val.([]byte))
        if err != nil {
            return err
        } else if n < len(val.([]byte)) {
            return errors.New("couldn't write whole protstring")
        }
        return nil
	})
}
func (b *BitstreamWriter) WriteCachedSystemAddress(val rbxfile.ValueSystemAddress, caches *util.Caches) error {
	cache := &caches.SystemAddress

	return b.WriteWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		return b.WriteSystemAddressRaw(val.(rbxfile.ValueSystemAddress))
	})
}
