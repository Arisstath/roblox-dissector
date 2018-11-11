package bitstreams
import "encoding/binary"
import "errors"

func (b *BitstreamReader) ReadBool() (bool, error) {
	res, err := b.ReadBit()
	return bool(res), err
}

func (b *BitstreamReader) ReadUint16BE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.bytes(dest, 2)
	return binary.BigEndian.Uint16(dest), err
}

func (b *BitstreamReader) ReadUint16LE() (uint16, error) {
	dest := make([]byte, 2)
	err := b.bytes(dest, 2)
	return binary.LittleEndian.Uint16(dest), err
}

func (b *BitstreamReader) ReadBoolByte() (bool, error) {
	res, err := b.bits(8)
	return res == 1, err
}

func (b *BitstreamReader) ReadUint24LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 3)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *BitstreamReader) ReadUint32BE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 4)
	return binary.BigEndian.Uint32(dest), err
}
func (b *BitstreamReader) ReadUint32LE() (uint32, error) {
	dest := make([]byte, 4)
	err := b.bytes(dest, 4)
	return binary.LittleEndian.Uint32(dest), err
}

func (b *BitstreamReader) ReadUint64BE() (uint64, error) {
	dest := make([]byte, 8)
	err := b.bytes(dest, 8)
	return binary.BigEndian.Uint64(dest), err
}

func (b *BitstreamReader) ReadUint8() (uint8, error) {
	res, err := b.bits(8)
	return uint8(res), err
}

func (b *BitstreamReader) ReadString(length int) ([]byte, error) {
	var dest []byte
	if uint(length) > 0x1000000 {
		return dest, errors.New("Sanity check: string too long")
	}
	dest = make([]byte, length)
	err := b.bytes(dest, length)
	return dest, err
}

// Previously readLengthAndString
func (b *BitstreamReader) ReadUint16AndString() (string, error) {
	var ret []byte
	thisLen, err := b.ReadUint16BE()
	if err != nil {
		return "", err
	}
	b.Align()
	ret, err = b.ReadString(int(thisLen))
	return string(ret), err
}

func (b *BitstreamReader) ReadASCII(length int) (string, error) {
	res, err := b.ReadString(length)
	return string(res), err
}

func (b *BitstreamReader) ReadFloat32LE() (float32, error) {
	intf, err := b.ReadUint32LE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
}

func (b *BitstreamReader) ReadFloat32BE() (float32, error) {
	intf, err := b.ReadUint32BE()
	if err != nil {
		return 0.0, err
	}
	return math.Float32frombits(intf), err
}

func (b *BitstreamReader) ReadFloat64BE() (float64, error) {
	intf, err := b.bits(64)
	if err != nil {
		return 0.0, err
	}
	return math.Float64frombits(intf), err
}

func (b *BitstreamReader) ReadUint32AndString() (interface{}, error) {
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return "", err
	}
	return b.ReadASCII(int(stringLen))
}

func (b *BitstreamReader) ReadUintUTF8() (uint32, error) {
	var res uint32
	thisByte, err := b.ReadByte()
	var shiftIndex uint32
	for err == nil {
		res |= uint32(thisByte&0x7F) << shiftIndex
		shiftIndex += 7
		if thisByte&0x80 == 0 {
			break
		}
		thisByte, err = b.ReadByte()
	}
	return res, err
}
func (b *BitstreamReader) ReadSintUTF8() (int32, error) {
	res, err := b.ReadUintUTF8()
	return int32((res >> 1) ^ -(res & 1)), err
}
func (b *BitstreamReader) ReadVarint64() (uint64, error) {
	var res uint64
	thisByte, err := b.ReadByte()
	var shiftIndex uint32
	for err == nil {
		res |= uint64(thisByte&0x7F) << shiftIndex
		shiftIndex += 7
		if thisByte&0x80 == 0 {
			break
		}
		thisByte, err = b.ReadByte()
	}
	return res, err
}
func (b *BitstreamReader) ReadVarsint64() (int64, error) {
	res, err := b.ReadVarint64()
	return int64((res >> 1) ^ -(res & 1)), err
}

func (b *BitstreamReader) ReadVarLengthString() (string, error) {
	stringLen, err := b.ReadUintUTF8()
	if err != nil {
		return "", err
	}
	return b.ReadASCII(int(stringLen))
}

func (b *BitstreamWriter) WriteBool(value bool) error {
	return b.WriteBit(bitstream.Bit(value))
}

func (b *BitstreamWriter) WriteUint16BE(value uint16) error {
	dest := make([]byte, 2)
	binary.BigEndian.PutUint16(dest, value)
	return b.bytes(2, dest)
}
func (b *BitstreamWriter) WriteUint16LE(value uint16) error {
	dest := make([]byte, 2)
	binary.LittleEndian.PutUint16(dest, value)
	return b.bytes(2, dest)
}

func (b *BitstreamWriter) WriteUint32BE(value uint32) error {
	dest := make([]byte, 4)
	binary.BigEndian.PutUint32(dest, value)
	return b.bytes(4, dest)
}
func (b *BitstreamWriter) WriteUint32LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.bytes(4, dest)
}

func (b *BitstreamWriter) WriteUint64BE(value uint64) error {
	dest := make([]byte, 8)
	binary.BigEndian.PutUint64(dest, value)
	return b.bytes(8, dest)
}

func (b *BitstreamWriter) WriteUint24LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.bytes(3, dest)
}

func (b *BitstreamWriter) WriteFloat32BE(value float32) error {
	return b.WriteUint32BE(math.Float32bits(value))
}

func (b *BitstreamWriter) WriteFloat64BE(value float64) error {
	return b.bits(64, math.Float64bits(value))
}

func (b *BitstreamWriter) WriteFloat16BE(value float32, min float32, max float32) error {
	return b.WriteUint16BE(uint16(value / (max - min) * 65535.0))
}

func (b *BitstreamWriter) WriteBoolByte(value bool) error {
	if value {
		return b.WriteByte(1)
	} else {
		return b.WriteByte(0)
	}
}

func (b *BitstreamWriter) WriteASCII(value string) error {
	return b.allBytes([]byte(value))
}

func (b *BitstreamWriter) WriteUintUTF8(value uint32) error {
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

func (b *BitstreamWriter) WriteUint32AndString(val interface{}) error {
	str := val.(string)
	err := b.WriteUint32BE(uint32(len(str)))
	if err != nil {
		return err
	}
	return b.WriteASCII(str)
}

