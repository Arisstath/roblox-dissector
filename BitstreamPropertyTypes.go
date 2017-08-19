package main
import "net"
import "github.com/google/gopacket"
import "errors"
import "fmt"

type pbool bool
type psint int32
type pfloat float32
type pdouble float64
type Axes int32
type Faces int32
type BrickColor uint64
type Object struct {
	Referent string
	ReferentInt uint32
}
type RebindObject struct {
	Referent1 uint32
	Referent2 uint16
}
type EnumValue int64
type pstring string
type ProtectedString []byte
type BinaryString []byte

type UDim struct {
	Scale float32
	Offset int32
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

type CFrameSimple struct {
	Position Vector3
	Matrix [9]float32
	SpecialRotMatrix uint64
}

type Content string

type SystemAddress struct {
	net.UDPAddr
}

type Region3 struct {
	Start Vector3
	End Vector3
}

type Region3uint16 struct {
	Start Vector3uint16
	End Vector3uint16
}
type Instance Object

type TypeAndValue struct {
	Type string
	Value PropertyValue
}

type NumberSequenceKeypoint struct {
	Time float32
	Value float32
	Envelope float32
}
type NumberSequence []NumberSequenceKeypoint
type NumberRange struct {
	Min float32
	Max float32
}

type ColorSequenceKeypoint struct {
	Time float32
	Value Color3
	Envelope float32
}
type ColorSequence []ColorSequenceKeypoint

type Rect2D struct {
	MinX float32
	MinY float32
	MaxX float32
	MaxY float32
}

type PhysicalProperties struct {
	Density float32
	Friction float32
	Elasticity float32
	FrictionWeight float32
	ElasticityWeight float32
}

type Tuple []TypeAndValue
type Array []TypeAndValue
type Dictionary map[string]TypeAndValue
type Map map[string]TypeAndValue

func (b *ExtendedReader) ReadUDim() (UDim, error) {
	var err error
	val := UDim{}
	val.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	off, err := b.ReadUint32BE()
	val.Offset = int32(off)
	return val, err
}

func (b *ExtendedReader) ReadUDim2() (UDim2, error) {
	var err error
	val := UDim2{UDim{}, UDim{}}
	val.X.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	offx, err := b.ReadUint32BE()
	val.X.Offset = int32(offx)
	if err != nil {
		return val, err
	}
	val.Y.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	offy, err := b.ReadUint32BE()
	val.Y.Offset = int32(offy)
	return val, err
}

func (b *ExtendedReader) ReadRay() (Ray, error) {
	var err error
	val := Ray{}
	val.Origin, err = b.ReadVector3Simple()
	if err != nil {
		return val, err
	}
	val.Direction, err = b.ReadVector3Simple()
	return val, err
}

func (b *ExtendedReader) ReadRegion3() (Region3, error) {
	var err error
	val := Region3{}
	val.Start, err = b.ReadVector3Simple()
	if err != nil {
		return val, err
	}
	val.End, err = b.ReadVector3Simple()
	return val, err
}
func (b *ExtendedReader) ReadRegion3uint16() (Region3uint16, error) {
	var err error
	val := Region3uint16{}
	val.Start, err = b.ReadVector3uint16()
	if err != nil {
		return val, err
	}
	val.End, err = b.ReadVector3uint16()
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

func (b *ExtendedReader) ReadVector3Simple() (Vector3, error) {
	var err error
	val := Vector3{}
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
}

func (b *ExtendedReader) ReadVector3() (Vector3, error) {
	isInteger, err := b.ReadBool()
	if err != nil {
		return Vector3{}, err
	}
	if !isInteger {
		return b.ReadVector3Simple()
	} else {
		var err error
		val := Vector3{}
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
func (b *ExtendedReader) ReadPSInt() (psint, error) {
	val, err := b.ReadUint32BE()
	return psint(val), err
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

func formatBindable(obj Object) string {
	return fmt.Sprintf("%s_%d", obj.Referent, obj.ReferentInt)
}

func (b *ExtendedReader) ReadObject(isJoinData bool, context *CommunicationContext) (Object, error) {
	var err error
	Object := Object{}
	if isJoinData {
		Object.Referent, Object.ReferentInt, err = b.ReadJoinReferent()
	} else {
		Object.Referent, err = b.ReadCachedObject(context)
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
		val, err := b.ReadCachedProtectedString(context)
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

func (b *ExtendedReader) ReadCFrameSimple() (CFrame, error) {
	return CFrame{}, nil // nop for now, since nothing uses this
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

func (b *ExtendedReader) ReadContent(isJoinData bool, context *CommunicationContext) (Content, error) {
	if !isJoinData {
		val, err := b.ReadCachedContent(context)
		return Content(val), err
	}
	var result string
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return Content(result), err
	}
	result, err = b.ReadASCII(int(stringLen))
	return Content(result), err
}

func (b *ExtendedReader) ReadSystemAddress(isJoinData bool, context *CommunicationContext) (SystemAddress, error) {
	var thisAddress SystemAddress
	var err error
	var cacheIndex uint8
	if !isJoinData {
		cacheIndex, err = b.ReadUint8()
		if err != nil {
			return thisAddress, err
		}
		if cacheIndex == 0x00 {
			return thisAddress, err
		}

		if cacheIndex < 0x80 {
			result := context.ReplicatorSystemAddressCache[cacheIndex]
			if result == nil {
				return SystemAddress{net.UDPAddr{[]byte{0,0,0,0}, 0, "udp"}}, nil
			}
			return result.(SystemAddress), nil
		}
	}
	thisAddress.IP = make([]byte, 4)
	err = b.Bytes(thisAddress.IP, 4)
	if err != nil {
		return thisAddress, err
	}
	port, err := b.ReadUint16BE()
	thisAddress.Port = int(port)
	if err != nil {
		return thisAddress, err
	}

	if !isJoinData {
		context.ReplicatorSystemAddressCache[cacheIndex - 0x80] = thisAddress
	}

	return thisAddress, nil
}

func (b *ExtendedReader) ReadTypeAndValue(context *CommunicationContext, packet gopacket.Packet) (TypeAndValue, error) {
	var value TypeAndValue
	schemaIDx, err := b.Bits(0x8)
	if err != nil {
		return value, err
	}
	if int(schemaIDx) > int(len(context.TypeDescriptor)) {
		return value, errors.New(fmt.Sprintf("type idx %d is higher than %d", schemaIDx, len(context.TypeDescriptor)))
	}
	value.Type = context.TypeDescriptor[uint32(schemaIDx)]
	value.Value, err = decodeEventArgument(b, context, packet, value.Type)
	return value, err
}

func (b *ExtendedReader) ReadTuple(context *CommunicationContext, packet gopacket.Packet) (Tuple, error) {
	var tuple Tuple
	tupleLen, err := b.ReadUint32BE()
	if err != nil {
		return tuple, err
	}
	if tupleLen > 0x10000 {
		return tuple, errors.New("sanity check: exceeded maximum tuple len")
	}
	tuple = make(Tuple, tupleLen)
	for i := 0; i < int(tupleLen); i++ {
		tuple[i], err = b.ReadTypeAndValue(context, packet)
		if err != nil {
			return tuple, err
		}
	}

	return tuple, nil
}

func (b *ExtendedReader) ReadArray(context *CommunicationContext, packet gopacket.Packet) (Array, error) {
	array, err := b.ReadTuple(context, packet)
	return Array(array), err
}

func (b *ExtendedReader) ReadDictionary(context *CommunicationContext, packet gopacket.Packet) (Dictionary, error) {
	var dictionary Dictionary
	dictionaryLen, err := b.ReadUint32BE()
	if err != nil {
		return dictionary, err
	}
	if dictionaryLen > 0x10000 {
		return dictionary, errors.New("sanity check: exceeded maximum dictionary len")
	}
	dictionary = make(Dictionary, dictionaryLen)
	for i := 0; i < int(dictionaryLen); i++ {
		keyLen, err := b.ReadUint32BE()
		if err != nil {
			return dictionary, err
		}
		key, err := b.ReadASCII(int(keyLen))
		if err != nil {
			return dictionary, err
		}
		dictionary[key], err = b.ReadTypeAndValue(context, packet)
		if err != nil {
			return dictionary, err
		}
	}

	return dictionary, nil
}

func (b *ExtendedReader) ReadMap(context *CommunicationContext, packet gopacket.Packet) (Map, error) {
	thisMap, err := b.ReadDictionary(context, packet)
	return Map(thisMap), err
}

func (b *ExtendedReader) ReadUintUTF8() (uint32, error) {
	var res uint32
	thisByte, err := b.ReadByte()
	var shiftIndex uint32 = 0
	for err == nil {
		res |= uint32(thisByte & 0x7F) << shiftIndex
		shiftIndex += 7
		if thisByte & 0x80 == 0 {
			break
		}
		thisByte, err = b.ReadByte()
	}
	return res, err
}
func (b *ExtendedReader) ReadSintUTF8() (int32, error) {
	res, err := b.ReadUintUTF8()
	return int32((res >> 1) ^ -(res & 1)), err
}

func (b *ExtendedReader) ReadNewPString(isJoinData bool, context *CommunicationContext) (pstring, error) {
	if !isJoinData {
		val, err := b.ReadCached(context)
		return pstring(val), err
	}
	stringLen, err := b.ReadUintUTF8()
	if err != nil {
		return pstring(""), err
	}
	val, err := b.ReadASCII(int(stringLen))
	return pstring(val), err
}
func (b *ExtendedReader) ReadNewProtectedString(isJoinData bool, context *CommunicationContext) (ProtectedString, error) {
	if !isJoinData {
		res, err := b.ReadNewCachedProtectedString(context)
		return ProtectedString(res), err
	}
	res, err := b.ReadNewPString(true, context)
	return ProtectedString(res), err
}
func (b *ExtendedReader) ReadNewContent(isJoinData bool, context *CommunicationContext) (Content, error) {
	if !isJoinData {
		res, err := b.ReadCachedContent(context)
		return Content(res), err
	}
	res, err := b.ReadNewPString(true, context)
	return Content(res), err
}
func (b *ExtendedReader) ReadNewBinaryString() (BinaryString, error) {
	res, err := b.ReadNewPString(true, nil)
	return BinaryString(res), err
}

func (b *ExtendedReader) ReadNewEnumValue() (EnumValue, error) {
	val, err := b.ReadUintUTF8()
	return EnumValue(val), err
}

func (b *ExtendedReader) ReadNewPSint() (psint, error) {
	val, err := b.ReadSintUTF8()
	return psint(val), err
}

func (b *ExtendedReader) ReadNewTypeAndValue(isJoinData bool, context *CommunicationContext) (TypeAndValue, error) {
	val := TypeAndValue{}
	thisType, err := b.ReadUint8()
	val.Type = TypeNames[thisType]
	if thisType == 7 {
		_, err = b.ReadUint16BE()
		if err != nil {
			return val, err
		}
	}

	val.Value, err = readSerializedValue(isJoinData, thisType, b, context)
	return val, err
}

func (b *ExtendedReader) ReadNewTuple(isJoinData bool, context *CommunicationContext) (Tuple, error) {
	var tuple Tuple
	tupleLen, err := b.ReadUintUTF8()
	if err != nil {
		return tuple, err
	}
	if tupleLen > 0x10000 {
		return tuple, errors.New("sanity check: exceeded maximum tuple len")
	}
	tuple = make(Tuple, tupleLen)
	for i := 0; i < int(tupleLen); i++ {
		val, err := b.ReadNewTypeAndValue(isJoinData, context)
		if err != nil {
			return tuple, err
		}
		tuple[i] = val
	}

	return tuple, nil
}

func (b *ExtendedReader) ReadNewArray(isJoinData bool, context *CommunicationContext) (Array, error) {
	array, err := b.ReadNewTuple(isJoinData, context)
	return Array(array), err
}

func (b *ExtendedReader) ReadNewDictionary(isJoinData bool, context *CommunicationContext) (Dictionary, error) {
	var dictionary Dictionary
	dictionaryLen, err := b.ReadUintUTF8()
	if err != nil {
		return dictionary, err
	}
	if dictionaryLen > 0x10000 {
		return dictionary, errors.New("sanity check: exceeded maximum dictionary len")
	}
	dictionary = make(Dictionary, dictionaryLen)
	for i := 0; i < int(dictionaryLen); i++ {
		keyLen, err := b.ReadUintUTF8()
		if err != nil {
			return dictionary, err
		}
		key, err := b.ReadASCII(int(keyLen))
		if err != nil {
			return dictionary, err
		}
		dictionary[key], err = b.ReadNewTypeAndValue(isJoinData, context)
		if err != nil {
			return dictionary, err
		}
	}

	return dictionary, nil
}

func (b *ExtendedReader) ReadNewMap(isJoinData bool, context *CommunicationContext) (Map, error) {
	thisMap, err := b.ReadNewDictionary(isJoinData, context)
	return Map(thisMap), err
}

func (b *ExtendedReader) ReadNumberSequenceKeypoint() (NumberSequenceKeypoint, error) {
	var err error
	thisKeypoint := NumberSequenceKeypoint{}
	thisKeypoint.Time, err = b.ReadFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Value, err = b.ReadFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Envelope, err = b.ReadFloat32BE()
	return thisKeypoint, err
}

func (b *ExtendedReader) ReadNumberSequence() (NumberSequence, error) {
	var err error
	numKeypoints, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum numberseq len")
	}
	thisSequence := make(NumberSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.ReadNumberSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *ExtendedReader) ReadNumberRange() (NumberRange, error) {
	thisRange := NumberRange{}
	var err error
	thisRange.Min, err = b.ReadFloat32BE()
	if err != nil {
		return thisRange, err
	}
	thisRange.Max, err = b.ReadFloat32BE()
	return thisRange, err
}

func (b *ExtendedReader) ReadColorSequenceKeypoint() (ColorSequenceKeypoint, error) {
	var err error
	thisKeypoint := ColorSequenceKeypoint{}
	thisKeypoint.Time, err = b.ReadFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Value, err = b.ReadColor3()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Envelope, err = b.ReadFloat32BE()
	return thisKeypoint, err
}

func (b *ExtendedReader) ReadColorSequence() (ColorSequence, error) {
	var err error
	numKeypoints, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum colorseq len")
	}
	thisSequence := make(ColorSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.ReadColorSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *ExtendedReader) ReadRect2D() (Rect2D, error) {
	var err error
	thisRect := Rect2D{}

	thisRect.MinX, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.MinY, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.MaxX, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.MaxY, err = b.ReadFloat32BE()
	return thisRect, err
}

func (b *ExtendedReader) ReadPhysicalProperties() (PhysicalProperties, error) {
	var err error
	props := PhysicalProperties{}
	hasProperties, err := b.ReadBool()
	if hasProperties {
		if err != nil {
			return props, err
		}
		props.Density, err = b.ReadFloat32BE()
		if err != nil {
			return props, err
		}
		props.Friction, err = b.ReadFloat32BE()
		if err != nil {
			return props, err
		}
		props.Elasticity, err = b.ReadFloat32BE()
		if err != nil {
			return props, err
		}
		props.FrictionWeight, err = b.ReadFloat32BE()
		if err != nil {
			return props, err
		}
		props.ElasticityWeight, err = b.ReadFloat32BE()
	}

	return props, err
}
