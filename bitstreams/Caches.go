package bitstreams

// Cache represents a network cache that stores repeatable objects such as strings.
type Cache interface {
	// Get fetches the object matching a cache index.
	Get(uint8) (interface{}, bool)
	// Put creates a new object in the caches at the specific index.
	Put(interface{}, uint8)
	// Equal compares a cached object and another object and returns if
	// they are equal, and if the indexed object exists.
	Equal(uint8, interface{}) (bool, bool)
	// LastWrite returns the cache index that was last written to.
	LastWrite() uint8
}

// StringCache represents a cache that stores strings.
type StringCache struct {
	// Values contains the stored strings.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get()
func (c *StringCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put()
func (c *StringCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(string)
	c.lastWrite = index
}

// Equal implements Cache.Equal()
func (c *StringCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(string) == val.(string), val1 != nil
}

// LastWrite implements Cache.LastWrite()
func (c *StringCache) LastWrite() uint8 {
	return c.lastWrite
}

// SysAddrCache is a Cache that stores SystemAddress values.
type SysAddrCache struct {
	// Values contains the stores SystemAddresses.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *SysAddrCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *SysAddrCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(rbxfile.ValueSystemAddress)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *SysAddrCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(rbxfile.ValueSystemAddress).String() == val.(rbxfile.ValueSystemAddress).String(), val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *SysAddrCache) LastWrite() uint8 {
	return c.lastWrite
}

// ByteSliceCache is a Cache that stores []byte objects.
type ByteSliceCache struct {
	// Values contains the []byte objects.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *ByteSliceCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *ByteSliceCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.([]byte)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *ByteSliceCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return bytes.Compare(val1.([]byte), val.([]byte)) == 0, val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *ByteSliceCache) LastWrite() uint8 {
	return c.lastWrite
}

type Caches struct {
	String          StringCache
	Object          StringCache
	Content         StringCache
	SystemAddress   SysAddrCache
	ProtectedString ByteSliceCache
}

type cacheReadCallback func(*BitstreamReader) (interface{}, error)

var CacheReadOOB = errors.New("Cache read is out of bounds")

func (b *BitstreamReader) readWithCache(cache Cache, readCallback cacheReadCallback) (interface{}, error) {
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

// TODO: Perhaps make readWithCache() operate with a member function of the cache instead?
func (b *BitstreamReader) readCached(caches *Caches) (string, error) {
	cache := &caches.String

	thisString, err := b.readWithCache(cache, (*BitstreamReader).readUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) readCachedScope(caches *Caches) (string, error) {
	cache := &caches.Object
	thisString, err := b.readWithCache(cache, (*BitstreamReader).readUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) readCachedContent(caches *Caches) (string, error) {
	cache := &caches.Content

	thisString, err := b.readWithCache(cache, (*BitstreamReader).readUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) readNewCachedProtectedString(caches *Caches) ([]byte, error) {
	cache := &caches.ProtectedString

	thisString, err := b.readWithCache(cache, func(b *BitstreamReader) (interface{}, error) {
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

type cacheWriteCallback func(*BitstreamWriter, interface{}) error

func (b *BitstreamWriter) WriteWithCache(value interface{}, cache Cache, writeCallback cacheWriteCallback) error {
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

func (b *BitstreamWriter) WriteCached(val string, caches *Caches) error {
	cache := &caches.String

	return b.writeWithCache(val, cache, (*BitstreamWriter).writeUint32AndString)
}
func (b *BitstreamWriter) WriteCachedObject(val string, caches *Caches) error {
	cache := &caches.Object

	return b.writeWithCache(val, cache, (*BitstreamWriter).writeUint32AndString)
}
func (b *BitstreamWriter) WriteCachedContent(val string, caches *Caches) error {
	cache := &caches.Content

	return b.writeWithCache(val, cache, (*BitstreamWriter).writeUint32AndString)
}
func (b *BitstreamWriter) WriteNewCachedProtectedString(val []byte, caches *Caches) error {
	cache := &caches.ProtectedString

	return b.writeWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		str := val.([]byte)
		err := b.writeUint32BE(uint32(len(str)))
		if err != nil {
			return err
		}
		return b.allBytes(val.([]byte))
	})
}
func (b *BitstreamWriter) WriteCachedSystemAddress(val rbxfile.ValueSystemAddress, caches *Caches) error {
	cache := &caches.SystemAddress

	return b.writeWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		return b.writeSystemAddressRaw(val.(rbxfile.ValueSystemAddress))
	})
}
