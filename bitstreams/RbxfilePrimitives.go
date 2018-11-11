package bitstreams
import "github.com/gskartwii/rbxfile"

func (b *extendedReader) readNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.readSintUTF8()
	return rbxfile.ValueInt(val), err
}
func (b *extendedReader) readPBool() (rbxfile.ValueBool, error) {
	val, err := b.readBoolByte()
	return rbxfile.ValueBool(val), err
}

// reads a signed integer
func (b *extendedReader) readPSInt() (rbxfile.ValueInt, error) {
	val, err := b.readUint32BE()
	return rbxfile.ValueInt(val), err
}

// reads a single-precision float
func (b *extendedReader) readPFloat() (rbxfile.ValueFloat, error) {
	val, err := b.readFloat32BE()
	return rbxfile.ValueFloat(val), err
}

// reads a double-precision float
func (b *extendedReader) readPDouble() (rbxfile.ValueDouble, error) {
	val, err := b.readFloat64BE()
	return rbxfile.ValueDouble(val), err
}
func (b *extendedReader) readNewPString(caches *Caches) (rbxfile.ValueString, error) {
	val, err := b.readCached(caches)
	return rbxfile.ValueString(val), err
}

func (b *extendedReader) readNewProtectedString(caches *Caches) (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewCachedProtectedString(caches)
	return rbxfile.ValueProtectedString(res), err
}

func (b *extendedReader) readNewContent(caches *Caches) (rbxfile.ValueContent, error) {
	res, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(res), err
}
func (b *extendedReader) readNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.readVarLengthString()
	return rbxfile.ValueBinaryString(res), err
}
func (b *extendedReader) readInt64() (rbxfile.ValueInt64, error) {
	val, err := b.readVarsint64()
	return rbxfile.ValueInt64(val), err
}
func (b *extendedReader) readContent(caches *Caches) (rbxfile.ValueContent, error) {
	val, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(val), err
}


func (b *extendedWriter) writePBool(val rbxfile.ValueBool) error {
	return b.writeBoolByte(bool(val))
}
func (b *extendedWriter) writePSint(val rbxfile.ValueInt) error {
	return b.writeUint32BE(uint32(val))
}
func (b *extendedWriter) writePFloat(val rbxfile.ValueFloat) error {
	return b.writeFloat32BE(float32(val))
}
func (b *extendedWriter) writePDouble(val rbxfile.ValueDouble) error {
	return b.writeFloat64BE(float64(val))
}
func (b *extendedWriter) writeNewPString(val rbxfile.ValueString, caches *Caches) error {
	return b.writeCached(string(val), caches)
}
func (b *extendedWriter) writePStringNoCache(val rbxfile.ValueString) error {
	return b.writeVarLengthString(string(val))
}
func (b *extendedWriter) writeNewProtectedString(val rbxfile.ValueProtectedString, caches *Caches) error {
	return b.writeNewCachedProtectedString([]byte(val), caches)
}
func (b *extendedWriter) writeNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.writeVarLengthString(string(val))
}
func (b *extendedWriter) writeNewContent(val rbxfile.ValueContent, caches *Caches) error {
	return b.writeCachedContent(string(val), caches)
}
func (b *extendedWriter) writeCFrameSimple(val rbxfile.ValueCFrame) error {
	return errors.New("not implemented!")
}
