package bitstreams
import "github.com/gskartwii/rbxfile"

func (b *BitstreamReader) readNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.readSintUTF8()
	return rbxfile.ValueInt(val), err
}
func (b *BitstreamReader) readPBool() (rbxfile.ValueBool, error) {
	val, err := b.readBoolByte()
	return rbxfile.ValueBool(val), err
}

// reads a signed integer
func (b *BitstreamReader) readPSInt() (rbxfile.ValueInt, error) {
	val, err := b.readUint32BE()
	return rbxfile.ValueInt(val), err
}

// reads a single-precision float
func (b *BitstreamReader) readPFloat() (rbxfile.ValueFloat, error) {
	val, err := b.readFloat32BE()
	return rbxfile.ValueFloat(val), err
}

// reads a double-precision float
func (b *BitstreamReader) readPDouble() (rbxfile.ValueDouble, error) {
	val, err := b.readFloat64BE()
	return rbxfile.ValueDouble(val), err
}
func (b *BitstreamReader) readNewPString(caches *Caches) (rbxfile.ValueString, error) {
	val, err := b.readCached(caches)
	return rbxfile.ValueString(val), err
}

func (b *BitstreamReader) readNewProtectedString(caches *Caches) (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewCachedProtectedString(caches)
	return rbxfile.ValueProtectedString(res), err
}

func (b *BitstreamReader) readNewContent(caches *Caches) (rbxfile.ValueContent, error) {
	res, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(res), err
}
func (b *BitstreamReader) readNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.readVarLengthString()
	return rbxfile.ValueBinaryString(res), err
}
func (b *BitstreamReader) readInt64() (rbxfile.ValueInt64, error) {
	val, err := b.readVarsint64()
	return rbxfile.ValueInt64(val), err
}
func (b *BitstreamReader) readContent(caches *Caches) (rbxfile.ValueContent, error) {
	val, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(val), err
}


func (b *BitstreamWriter) WritePBool(val rbxfile.ValueBool) error {
	return b.WriteBoolByte(bool(val))
}
func (b *BitstreamWriter) WritePSint(val rbxfile.ValueInt) error {
	return b.WriteUint32BE(uint32(val))
}
func (b *BitstreamWriter) WritePFloat(val rbxfile.ValueFloat) error {
	return b.WriteFloat32BE(float32(val))
}
func (b *BitstreamWriter) WritePDouble(val rbxfile.ValueDouble) error {
	return b.WriteFloat64BE(float64(val))
}
func (b *BitstreamWriter) WriteNewPString(val rbxfile.ValueString, caches *Caches) error {
	return b.WriteCached(string(val), caches)
}
func (b *BitstreamWriter) WritePStringNoCache(val rbxfile.ValueString) error {
	return b.WriteVarLengthString(string(val))
}
func (b *BitstreamWriter) WriteNewProtectedString(val rbxfile.ValueProtectedString, caches *Caches) error {
	return b.WriteNewCachedProtectedString([]byte(val), caches)
}
func (b *BitstreamWriter) WriteNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.WriteVarLengthString(string(val))
}
func (b *BitstreamWriter) WriteNewContent(val rbxfile.ValueContent, caches *Caches) error {
	return b.WriteCachedContent(string(val), caches)
}
func (b *BitstreamWriter) WriteCFrameSimple(val rbxfile.ValueCFrame) error {
	return errors.New("not implemented!")
}
