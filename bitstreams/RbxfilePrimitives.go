package bitstreams
import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/util"
import "errors"

func (b *BitstreamReader) ReadNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.ReadSintUTF8()
	return rbxfile.ValueInt(val), err
}
func (b *BitstreamReader) ReadPBool() (rbxfile.ValueBool, error) {
	val, err := b.ReadBoolByte()
	return rbxfile.ValueBool(val), err
}

// reads a signed integer
func (b *BitstreamReader) ReadPSInt() (rbxfile.ValueInt, error) {
	val, err := b.ReadUint32BE()
	return rbxfile.ValueInt(val), err
}

// reads a single-precision float
func (b *BitstreamReader) ReadPFloat() (rbxfile.ValueFloat, error) {
	val, err := b.ReadFloat32BE()
	return rbxfile.ValueFloat(val), err
}

// reads a double-precision float
func (b *BitstreamReader) ReadPDouble() (rbxfile.ValueDouble, error) {
	val, err := b.ReadFloat64BE()
	return rbxfile.ValueDouble(val), err
}
func (b *BitstreamReader) ReadNewPString(caches *util.Caches) (rbxfile.ValueString, error) {
	val, err := b.ReadCached(caches)
	return rbxfile.ValueString(val), err
}

func (b *BitstreamReader) ReadNewProtectedString(caches *util.Caches) (rbxfile.ValueProtectedString, error) {
	res, err := b.ReadNewCachedProtectedString(caches)
	return rbxfile.ValueProtectedString(res), err
}

func (b *BitstreamReader) ReadNewContent(caches *util.Caches) (rbxfile.ValueContent, error) {
	res, err := b.ReadCachedContent(caches)
	return rbxfile.ValueContent(res), err
}
func (b *BitstreamReader) ReadNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.ReadVarLengthString()
	return rbxfile.ValueBinaryString(res), err
}
func (b *BitstreamReader) ReadInt64() (rbxfile.ValueInt64, error) {
	val, err := b.ReadVarsint64()
	return rbxfile.ValueInt64(val), err
}
func (b *BitstreamReader) ReadContent(caches *util.Caches) (rbxfile.ValueContent, error) {
	val, err := b.ReadCachedContent(caches)
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
func (b *BitstreamWriter) WriteNewPString(val rbxfile.ValueString, caches *util.Caches) error {
	return b.WriteCached(string(val), caches)
}
func (b *BitstreamWriter) WritePStringNoCache(val rbxfile.ValueString) error {
	return b.WriteVarLengthString(string(val))
}
func (b *BitstreamWriter) WriteNewProtectedString(val rbxfile.ValueProtectedString, caches *util.Caches) error {
	return b.WriteNewCachedProtectedString([]byte(val), caches)
}
func (b *BitstreamWriter) WriteNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.WriteVarLengthString(string(val))
}
func (b *BitstreamWriter) WriteNewContent(val rbxfile.ValueContent, caches *util.Caches) error {
	return b.WriteCachedContent(string(val), caches)
}
func (b *BitstreamWriter) WriteCFrameSimple(val rbxfile.ValueCFrame) error {
	return errors.New("not implemented!")
}
