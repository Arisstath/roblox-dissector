package main

type pbool bool
type pint int32
type pfloat float32
type pdouble float64
type Axes int32
type Faces int32
type BrickColor uint64
type Object struct {
	Referent string
	ReferentInt uint32
}
type EnumValue int64
type pstring string
type ProtectedString []byte
type BinaryString []byte

type UDim struct {
	Scale float32
	Offset uint32
}

type UDim2 struct {
	X UDim
	Y UDim
}

type Vector2 struct {
	X float32
	Y float32
}
type Vector3 struct {
	X float32
	Y float32
	Z float32
}
type Vector2uint16 struct {
	X uint16
	Y uint16
}
type Vector3uint16 struct {
	X uint16
	Y uint16
	Z uint16
}

type Ray struct {
	Origin Vector3
	Direction Vector3
}

type Color3 struct {
	R float32
	G float32
	B float32
}
type Color3uint8 struct {
	R uint8
	G uint8
	B uint8
}

type CFrame struct {
	Position Vector3
	Matrix [4]float32
	SpecialRotMatrix uint64
}

type Content string

func (b *ExtendedReader) ReadUDim() (UDim, error) {
	var err error
	val := UDim{}
	val.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Offset, err = b.ReadUint32BE()
	return val, err
}

func (b *ExtendedReader) ReadUDim2() (UDim2, error) {
	var err error
	val := UDim2{UDim{}, UDim{}}
	val.X.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.X.Offset, err = b.ReadUint32BE()
	if err != nil {
		return val, err
	}
	val.Y.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Y.Offset, err = b.ReadUint32BE()
	return val, err
}

func (b *ExtendedReader) ReadRay() (Ray, error) {
	var err error
	val := Ray{Vector3{}, Vector3{}}
	val.Origin.X, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Origin.Y, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Origin.Z, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Direction.X, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Direction.Y, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Direction.Z, err = b.ReadFloat32BE()
	return val, err
}

func (b *ExtendedReader) ReadColor3() (Color3, error) {
	var err error
	val := Color3{}
	val.R, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.G, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.B, err = b.ReadFloat32BE()
	return val, err
}

func (b *ExtendedReader) ReadColor3uint8() (Color3uint8, error) {
	var err error
	val := Color3uint8{}
	val.R, err = b.ReadByte()
	if err != nil {
		return val, err
	}
	val.G, err = b.ReadByte()
	if err != nil {
		return val, err
	}
	val.B, err = b.ReadByte()
	return val, err
}

func (b *ExtendedReader) ReadVector2() (Vector2, error) {
	var err error
	val := Vector2{}
	val.X, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.ReadFloat32BE()
	return val, err
}

func (b *ExtendedReader) ReadVector3() (Vector3, error) {
	var err error
	val := Vector3{}
	isInteger, err := b.ReadBool()
	if !isInteger {
		val.X, err = b.ReadFloat32BE()
		if err != nil {
			return val, err
		}
		val.Y, err = b.ReadFloat32BE()
		if err != nil {
			return val, err
		}
		val.Z, err = b.ReadFloat32BE()
		return val, err
	} else {
		x, err := b.Bits(11)
		if err != nil {
			return val, err
		}
		x_short := uint16(((x & 0xFFF8) >> 3) | ((x & 7) << 8))
		if x_short & 0x400 != 0 {
			x_short |= 0xFC00
		}

		y, err := b.Bits(11)
		if err != nil {
			return val, err
		}

		z, err := b.Bits(11)
		if err != nil {
			return val, err
		}

		z_short := uint16(((z & 0xFFF8) >> 3) | ((z & 7) << 8))
		if z_short & 0x400 != 0 {
			z_short |= 0xFC00
		}

		val.X = float32(int16(x_short))
		val.Y = float32(y)
		val.Z = float32(int16(z_short))
		return val, err
	}
}

func (b *ExtendedReader) ReadVector2uint16() (Vector2uint16, error) {
	var err error
	val := Vector2uint16{}
	val.X, err = b.ReadUint16BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.ReadUint16BE()
	return val, err
}

func (b *ExtendedReader) ReadVector3uint16() (Vector3uint16, error) {
	var err error
	val := Vector3uint16{}
	val.X, err = b.ReadUint16BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.ReadUint16BE()
	if err != nil {
		return val, err
	}
	val.Z, err = b.ReadUint16BE()
	return val, err
}

func (b *ExtendedReader) ReadPBool() (pbool, error) {
	val, err := b.ReadBool()
	return pbool(val), err
}
func (b *ExtendedReader) ReadPInt() (pint, error) {
	val, err := b.ReadUint32BE()
	return pint(val), err
}
func (b *ExtendedReader) ReadPFloat() (pfloat, error) {
	val, err := b.ReadFloat32BE()
	return pfloat(val), err
}
func (b *ExtendedReader) ReadPDouble() (pdouble, error) {
	val, err := b.ReadFloat64BE()
	return pdouble(val), err
}
func (b *ExtendedReader) ReadAxes() (Axes, error) {
	val, err := b.ReadUint32BE()
	return Axes(val), err
}
func (b *ExtendedReader) ReadFaces() (Faces, error) {
	val, err := b.ReadUint32BE()
	return Faces(val), err
}
func (b *ExtendedReader) ReadBrickColor() (BrickColor, error) {
	val, err := b.Bits(7)
	return BrickColor(val), err
}
func (b *ExtendedReader) ReadObject(isJoinData bool, context *CommunicationContext) (Object, error) {
	var err error
	Object := Object{}
	if isJoinData {
		Object.Referent, Object.ReferentInt, err = b.ReadJoinReferent()
	} else {
		Object.Referent, err = b.ReadCached(context)
		if err != nil {
			return Object, err
		}
		Object.ReferentInt, err = b.ReadUint32LE()
	}
	return Object, err
}
func (b *ExtendedReader) ReadEnumValue(bitSize uint32) (EnumValue, error) {
	val, err := b.Bits(int(bitSize + 1))
	return EnumValue(val), err
}
func (b *ExtendedReader) ReadPString(isJoinData bool, context *CommunicationContext) (pstring, error) {
	if !isJoinData {
		val, err := b.ReadCached(context)
		return pstring(val), err
	}
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return pstring(""), err
	}
	val, err := b.ReadASCII(int(stringLen))
	return pstring(val), err
}
func (b *ExtendedReader) ReadProtectedString(isJoinData bool, context *CommunicationContext) (ProtectedString, error) {
	if !isJoinData {
		val, err := b.ReadCached(context)
		return ProtectedString(val), err
	}
	b.Align() // BitStream::operator>>(BinaryString) does implicit alignment. why?
	var result []byte
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return result, err
	}
	val, err := b.ReadString(int(stringLen))
	return ProtectedString(val), err
}
func (b *ExtendedReader) ReadBinaryString() (BinaryString, error) {
	b.Align() // BitStream::operator>>(BinaryString) does implicit alignment. why?
	var result []byte
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return result, err
	}
	val, err := b.ReadString(int(stringLen))
	return BinaryString(val), err
}

func (b *ExtendedReader) ReadCFrame() (CFrame, error) {
	var err error
	val := CFrame{}
	val.Position, err = b.ReadVector3()
	if err != nil {
		return val, err
	}

	isLookup, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	if isLookup {
		val.SpecialRotMatrix, err = b.Bits(6)
	} else {
		val.Matrix[3], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		val.Matrix[0], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		val.Matrix[1], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		val.Matrix[2], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
	}

	return val, nil
}

func (b *ExtendedReader) ReadContent() (Content, error) {
	hasContent, err := b.ReadBoolByte()
	if err != nil {
		return Content(""), err
	}
	if hasContent {
		contentLen, err := b.ReadUint32BE()
		if err != nil {
			return Content(""), err
		}

		content, err := b.ReadASCII(int(contentLen))
		return Content(content), err
	}
	return Content(""), nil
}
