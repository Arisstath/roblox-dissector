package bitstreams

// Cache represents a network cache that stores repeatable objects such as strings.
type cache interface {
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

// stringCache represents a cache that stores strings.
type stringCache struct {
	// Values contains the stored strings.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get()
func (c *stringCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put()
func (c *stringCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(string)
	c.lastWrite = index
}

// Equal implements Cache.Equal()
func (c *stringCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(string) == val.(string), val1 != nil
}

// LastWrite implements Cache.LastWrite()
func (c *stringCache) LastWrite() uint8 {
	return c.lastWrite
}

// sysAddrCache is a Cache that stores SystemAddress values.
type sysAddrCache struct {
	// Values contains the stores SystemAddresses.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *sysAddrCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *sysAddrCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.(rbxfile.ValueSystemAddress)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *sysAddrCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return val1.(rbxfile.ValueSystemAddress).String() == val.(rbxfile.ValueSystemAddress).String(), val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *sysAddrCache) LastWrite() uint8 {
	return c.lastWrite
}

// byteSliceCache is a Cache that stores []byte objects.
type byteSliceCache struct {
	// Values contains the []byte objects.
	Values    [0x80]interface{}
	lastWrite uint8
}

// Get implements Cache.Get().
func (c *byteSliceCache) Get(index uint8) (interface{}, bool) {
	a := c.Values[index]
	return a, a != nil
}

// Put implements Cache.Put().
func (c *byteSliceCache) Put(val interface{}, index uint8) {
	c.Values[index] = val.([]byte)
	c.lastWrite = index
}

// Equal implements Cache.Equal().
func (c *byteSliceCache) Equal(index uint8, val interface{}) (bool, bool) {
	val1 := c.Values[index]
	if val1 == nil || val == nil {
		return val1 == val, val1 == nil
	}
	return bytes.Compare(val1.([]byte), val.([]byte)) == 0, val1 != nil
}

// LastWrite implements Cache.LastWrite().
func (c *byteSliceCache) LastWrite() uint8 {
	return c.lastWrite
}

type Caches struct {
	String          stringCache
	Object          stringCache
	Content         stringCache
	SystemAddress   sysAddrCache
	ProtectedString byteSliceCache
}

type cacheReadCallback func(*BitstreamReader) (interface{}, error)

var CacheReadOOB = errors.New("Cache read is out of bounds")

func (b *BitstreamReader) readWithCache(cache Cache, readCallback cacheReadCallback) (interface{}, error) {
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
func (b *BitstreamReader) ReadCached(caches *Caches) (string, error) {
	cache := &caches.String

	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadCachedScope(caches *Caches) (string, error) {
	cache := &caches.Object
	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadCachedContent(caches *Caches) (string, error) {
	cache := &caches.Content

	thisString, err := b.readWithCache(cache, (*BitstreamReader).ReadUint32AndString)
	return thisString.(string), err
}

func (b *BitstreamReader) ReadNewCachedProtectedString(caches *Caches) ([]byte, error) {
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

func (b *BitstreamWriter) WriteWithCache(value interface{}, cache cache, writeCallback cacheWriteCallback) error {
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

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteCachedObject(val string, caches *Caches) error {
	cache := &caches.Object

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteCachedContent(val string, caches *Caches) error {
	cache := &caches.Content

	return b.WriteWithCache(val, cache, (*BitstreamWriter).WriteUint32AndString)
}
func (b *BitstreamWriter) WriteNewCachedProtectedString(val []byte, caches *Caches) error {
	cache := &caches.ProtectedString

	return b.WriteWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		str := val.([]byte)
		err := b.WriteUint32BE(uint32(len(str)))
		if err != nil {
			return err
		}
		return b.allBytes(val.([]byte))
	})
}
func (b *BitstreamWriter) WriteCachedSystemAddress(val rbxfile.ValueSystemAddress, caches *Caches) error {
	cache := &caches.SystemAddress

	return b.WriteWithCache(val, cache, func(b *BitstreamWriter, val interface{}) error {
		return b.WriteSystemAddressRaw(val.(rbxfile.ValueSystemAddress))
	})
}
