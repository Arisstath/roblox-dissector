package peer

import "net"
import "errors"
import "github.com/gskartwii/rbxfile"
import "fmt"
import "math"
import "strings"

// Referent is a type that is used to refer to rbxfile.Instances.
// A Referent to a NULL instance is "NULL2"
// Other Referents are of the the form "RBX123456789ABCDEF_1234", consisting of
// a scope and an index number
type Referent string

// IsNull checks if a a Referent refers to a NULL/nil instance.
func (ref Referent) IsNull() bool {
	return ref == "null" || ref == "NULL2"
}
func (ref Referent) String() string {
	return string(ref)
}

// PhysicsMotor is an alias type for rbxfile.ValueCFrames. They are used to
// describe motors in physics packets
type PhysicsMotor rbxfile.ValueCFrame

// Returns the stringified version of the motor
func (m PhysicsMotor) String() string {
	return rbxfile.ValueCFrame(m).String()
}

func (b *extendedReader) readUDim() (rbxfile.ValueUDim, error) {
	var err error
	val := rbxfile.ValueUDim{}
	val.Scale, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	off, err := b.readUint32BE()
	val.Offset = int16(off)
	return val, err
}

func (b *extendedReader) readUDim2() (rbxfile.ValueUDim2, error) {
	var err error
	val := rbxfile.ValueUDim2{rbxfile.ValueUDim{}, rbxfile.ValueUDim{}}
	val.X.Scale, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	offx, err := b.readUint32BE()
	val.X.Offset = int16(offx)
	if err != nil {
		return val, err
	}
	val.Y.Scale, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	offy, err := b.readUint32BE()
	val.Y.Offset = int16(offy)
	return val, err
}

func (b *extendedReader) readRay() (rbxfile.ValueRay, error) {
	var err error
	val := rbxfile.ValueRay{}
	val.Origin, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}
	val.Direction, err = b.readVector3Simple()
	return val, err
}

func (b *extendedReader) readRegion3() (rbxfile.ValueRegion3, error) {
	var err error
	val := rbxfile.ValueRegion3{}
	val.Start, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}
	val.End, err = b.readVector3Simple()
	return val, err
}
func (b *extendedReader) readRegion3int16() (rbxfile.ValueRegion3int16, error) {
	var err error
	val := rbxfile.ValueRegion3int16{}
	val.Start, err = b.readVector3int16()
	if err != nil {
		return val, err
	}
	val.End, err = b.readVector3int16()
	return val, err
}

func (b *extendedReader) readColor3() (rbxfile.ValueColor3, error) {
	var err error
	val := rbxfile.ValueColor3{}
	val.R, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	val.G, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	val.B, err = b.readFloat32BE()
	return val, err
}

func (b *extendedReader) readColor3uint8() (rbxfile.ValueColor3uint8, error) {
	var err error
	val := rbxfile.ValueColor3uint8{}
	val.R, err = b.readUint8()
	if err != nil {
		return val, err
	}
	val.G, err = b.readUint8()
	if err != nil {
		return val, err
	}
	val.B, err = b.readUint8()
	return val, err
}

func (b *extendedReader) readVector2() (rbxfile.ValueVector2, error) {
	var err error
	val := rbxfile.ValueVector2{}
	val.X, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.readFloat32BE()
	return val, err
}

// reads a simple Vector3 value (f32 X, f32 Y, f32 Z)
func (b *extendedReader) readVector3Simple() (rbxfile.ValueVector3, error) {
	var err error
	val := rbxfile.ValueVector3{}
	val.X, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	val.Z, err = b.readFloat32BE()
	return val, err
}

// reads a complicated Vector3 value
func (b *extendedReader) readVector3() (rbxfile.ValueVector3, error) {
	isInteger, err := b.readBool()
	if err != nil {
		return rbxfile.ValueVector3{}, err
	}
	if !isInteger {
		return b.readVector3Simple()
	}
	val := rbxfile.ValueVector3{}
	x, err := b.bits(11)
	if err != nil {
		return val, err
	}
	xShort := uint16((x >> 3) | ((x & 7) << 8))
	if xShort&0x400 != 0 {
		xShort |= 0xFC00
	}

	y, err := b.bits(11)
	if err != nil {
		return val, err
	}
	yShort := uint16((y >> 3) | ((y & 7) << 8))

	z, err := b.bits(11)
	if err != nil {
		return val, err
	}

	zShort := uint16((z >> 3) | ((z & 7) << 8))
	if zShort&0x400 != 0 {
		zShort |= 0xFC00
	}

	val.X = float32(int16(xShort)) * 0.5
	val.Y = float32(yShort) * 0.1
	val.Z = float32(int16(zShort)) * 0.5
	return val, err
}

func (b *extendedReader) readVector2int16() (rbxfile.ValueVector2int16, error) {
	var err error
	val := rbxfile.ValueVector2int16{}
	valX, err := b.readUint16BE()
	if err != nil {
		return val, err
	}
	valY, err := b.readUint16BE()
	val.X = int16(valX)
	val.Y = int16(valY)
	return val, err
}

func (b *extendedReader) readVector3int16() (rbxfile.ValueVector3int16, error) {
	var err error
	var val rbxfile.ValueVector3int16
	valX, err := b.readUint16BE()
	if err != nil {
		return val, err
	}
	valY, err := b.readUint16BE()
	if err != nil {
		return val, err
	}
	valZ, err := b.readUint16BE()
	val = rbxfile.ValueVector3int16{int16(valX), int16(valY), int16(valZ)}
	return val, err
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
func (b *extendedReader) readAxes() (rbxfile.ValueAxes, error) {
	val, err := b.readUint32BE()
	axesVal := rbxfile.ValueAxes{
		X: val&4 != 0,
		Y: val&2 != 0,
		Z: val&1 != 0,
	}
	return axesVal, err
}
func (b *extendedReader) readFaces() (rbxfile.ValueFaces, error) {
	val, err := b.readUint32BE()
	facesVal := rbxfile.ValueFaces{
		Right:  val&32 != 0,
		Top:    val&16 != 0,
		Back:   val&8 != 0,
		Left:   val&4 != 0,
		Bottom: val&2 != 0,
		Front:  val&1 != 0,
	}
	return facesVal, err
}
func (b *extendedReader) readBrickColor() (rbxfile.ValueBrickColor, error) {
	val, err := b.readUint16BE()
	return rbxfile.ValueBrickColor(val), err
}

func formatBindable(obj rbxfile.ValueReference) string {
	return obj.String()
}
func objectToRef(referent string, referentInt uint32) Referent {
	if referentInt == 0 {
		return "null"
	}
	return Referent(fmt.Sprintf("%s_%d", referent, referentInt))
}
func refToObject(refString Referent) (string, uint32) {
	if refString.IsNull() {
		return "NULL2", 0
	}
	components := strings.Split(string(refString), "_")
	return components[0], uint32(mustAtoi(components[1]))
}

func (b *extendedReader) readJoinObject(context *CommunicationContext) (Referent, error) {
	referent, referentInt, err := b.readJoinReferent(context)
	serialized := objectToRef(referent, referentInt)

	return Referent(serialized), err
}
func (b *extendedReader) readObject(caches *Caches) (Referent, error) {
	var referentInt uint32
	referent, err := b.readCachedScope(caches)
	if err != nil {
		return "", err
	}
	if referent != "" {
		referentInt, err = b.readUint32LE()
	}

	serialized := objectToRef(referent, referentInt)

	return Referent(serialized), err
}

func (b *extendedReader) readCFrameSimple() (rbxfile.ValueCFrame, error) {
	return rbxfile.ValueCFrame{}, nil // nop for now, since nothing uses this
}

func quaternionToRotMatrix(q [4]float32) [9]float32 {
	X := q[0]
	Y := q[1]
	Z := q[2]
	W := q[3]
	XS := X * 2
	YS := Y * 2
	ZS := Z * 2
	WX := W * XS
	WY := W * YS
	WZ := W * ZS

	XX := XS * X
	XY := YS * X
	XZ := ZS * X
	YY := YS * Y
	YZ := ZS * Y
	ZZ := ZS * Z

	return [9]float32{
		1 - (YY + ZZ),
		XY - WZ,
		XZ + WY,
		XY + WZ,
		1 - (XX + ZZ),
		YZ - WX,
		XZ - WY,
		YZ + WX,
		1 - (XX + YY),
	}
}

func transformQuaternionToMatrix(q [4]float32) [9]float32 {
	midresult := quaternionToRotMatrix(q)

	xScaleFactor := float32(1.0 / math.Sqrt(float64(midresult[0]*midresult[0]+midresult[3]*midresult[3]+midresult[6]*midresult[6])))

	result := [9]float32{
		xScaleFactor * midresult[0], 0, 0,
		xScaleFactor * midresult[3], 0, 0,
		xScaleFactor * midresult[6], 0, 0,
	} // X has been normalized

	sXYSum := result[0]*midresult[1] + result[3]*midresult[4] + result[6]*midresult[7]
	trueR01 := midresult[1] - result[0]*sXYSum
	trueR11 := midresult[4] - result[3]*sXYSum
	trueR21 := midresult[7] - result[6]*sXYSum

	yScaleFactor := float32(1.0 / math.Sqrt(float64(trueR01*trueR01+trueR11*trueR11+trueR21*trueR21)))
	result[1] = trueR01 * yScaleFactor
	result[4] = trueR11 * yScaleFactor
	result[7] = trueR21 * yScaleFactor

	sYZSum := result[1]*midresult[2] + result[4]*midresult[5] + result[7]*midresult[8]
	sXZSum := result[0]*midresult[2] + result[3]*midresult[5] + result[6]*midresult[8]
	trueR02 := midresult[2] - result[0]*sXZSum - result[1]*sYZSum
	trueR12 := midresult[5] - result[3]*sXZSum - result[4]*sYZSum
	trueR22 := midresult[8] - result[6]*sXZSum - result[7]*sYZSum

	zScaleFactor := float32(1.0 / math.Sqrt(float64(trueR02*trueR02+trueR12*trueR12+trueR22*trueR22)))
	result[2] = trueR02 * zScaleFactor
	result[5] = trueR12 * zScaleFactor
	result[8] = trueR22 * zScaleFactor

	return result
}

var specialRows = [6][3]float32{
	[3]float32{1, 0, 0},
	[3]float32{0, 1, 0},
	[3]float32{0, 0, 1},
	[3]float32{-1, 0, 0},
	[3]float32{0, -1, 0},
	[3]float32{0, 0, -1},
}

func lookupRotMatrix(special uint64) [9]float32 {
	specialRowDiv6 := specialRows[special/6]
	specialRowMod6 := specialRows[special%6]

	ret := [9]float32{
		0, 0, 0,
		specialRowMod6[0], specialRowMod6[1], specialRowMod6[2],
		specialRowDiv6[0], specialRowDiv6[1], specialRowDiv6[2],
	}
	ret[0] = ret[2*3+1]*ret[1*3+2] - ret[2*3+2]*ret[1*3+1]
	ret[1] = ret[1*3+0]*ret[2*3+2] - ret[2*3+0]*ret[1*3+2]
	ret[2] = ret[2*3+0]*ret[1*3+1] - ret[1*3+0]*ret[2*3+1]

	trueRet := [9]float32{
		ret[6], ret[7], ret[8],
		ret[3], ret[4], ret[5],
		ret[0], ret[1], ret[2],
	}

	return trueRet
}

func (b *extendedReader) readCFrame() (rbxfile.ValueCFrame, error) {
	var err error
	val := rbxfile.ValueCFrame{}
	val.Position, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}

	special, err := b.readUint8()
	if err != nil {
		return val, err
	}
	if special > 0 {
		if err != nil {
			return val, err
		}
		if special > 36 {
			println("oob, special", special)
			return val, errors.New("special rotmatrix oob")
		}
		val.Rotation = lookupRotMatrix(uint64(special - 1))
	} else {
		val.Rotation, err = b.readPhysicsMatrix()
	}

	return val, err
}

func (b *JoinSerializeReader) readContent() (rbxfile.ValueContent, error) {
	var result string
	stringLen, err := b.readUint32BE()
	if err != nil {
		return rbxfile.ValueContent(result), err
	}
	result, err = b.readASCII(int(stringLen))
	return rbxfile.ValueContent(result), err
}

func (b *extendedReader) readContent(caches *Caches) (rbxfile.ValueContent, error) {
	val, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(val), err
}

// TODO: Make this function uniform with other cache functions
func (b *extendedReader) readSystemAddress(caches *Caches) (rbxfile.ValueSystemAddress, error) {
	cache := &caches.SystemAddress

	thisAddress := rbxfile.ValueSystemAddress("0.0.0.0:0")
	var err error
	var cacheIndex uint8
	cacheIndex, err = b.readUint8()
	if err != nil {
		return thisAddress, err
	}
	if cacheIndex == 0x00 {
		return thisAddress, err
	}

	if cacheIndex < 0x80 {
		result, ok := cache.Get(cacheIndex)
		if !ok {
			return thisAddress, nil
		}
		return result.(rbxfile.ValueSystemAddress), nil
	}
	thisAddr := net.UDPAddr{}
	thisAddr.IP = make([]byte, 4)
	err = b.bytes(thisAddr.IP, 4)
	if err != nil {
		return thisAddress, err
	}
	for i := 0; i < 4; i++ {
		thisAddr.IP[i] = thisAddr.IP[i] ^ 0xFF // bitwise NOT
	}

	port, err := b.readUint16BE()
	thisAddr.Port = int(port)
	if err != nil {
		return thisAddress, err
	}

	cache.Put(thisAddress, cacheIndex-0x80)

	return rbxfile.ValueSystemAddress(thisAddr.String()), nil
}

func (b *JoinSerializeReader) readSystemAddress() (rbxfile.ValueSystemAddress, error) {
	var err error
	thisAddress := rbxfile.ValueSystemAddress("0.0.0.0:0")
	thisAddr := net.UDPAddr{}
	thisAddr.IP = make([]byte, 4)
	err = b.bytes(thisAddr.IP, 4)
	if err != nil {
		return thisAddress, err
	}
	for i := 0; i < 4; i++ {
		thisAddr.IP[i] = thisAddr.IP[i] ^ 0xFF // bitwise NOT
	}

	port, err := b.readUint16BE()
	thisAddr.Port = int(port)
	if err != nil {
		return thisAddress, err
	}
	return rbxfile.ValueSystemAddress(thisAddr.String()), nil
}

func (b *extendedReader) readUintUTF8() (uint32, error) {
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
func (b *extendedReader) readSintUTF8() (int32, error) {
	res, err := b.readUintUTF8()
	return int32((res >> 1) ^ -(res & 1)), err
}
func (b *extendedReader) readVarint64() (uint64, error) {
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
func (b *extendedReader) readVarsint64() (int64, error) {
	res, err := b.readVarint64()
	return int64((res >> 1) ^ -(res & 1)), err
}

func (b *extendedReader) readVarLengthString() (string, error) {
	stringLen, err := b.readUintUTF8()
	if err != nil {
		return "", err
	}
	return b.readASCII(int(stringLen))
}

func (b *JoinSerializeReader) readNewPString() (rbxfile.ValueString, error) {
	val, err := b.readVarLengthString()
	return rbxfile.ValueString(val), err
}
func (b *extendedReader) readNewPString(caches *Caches) (rbxfile.ValueString, error) {
	val, err := b.readCached(caches)
	return rbxfile.ValueString(val), err
}

func (b *JoinSerializeReader) readNewProtectedString() (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewPString()
	return rbxfile.ValueProtectedString(res), err
}
func (b *extendedReader) readNewProtectedString(caches *Caches) (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewCachedProtectedString(caches)
	return rbxfile.ValueProtectedString(res), err
}

func (b *JoinSerializeReader) readNewContent() (rbxfile.ValueContent, error) {
	res, err := b.readNewPString()
	return rbxfile.ValueContent(res), err
}
func (b *extendedReader) readNewContent(caches *Caches) (rbxfile.ValueContent, error) {
	res, err := b.readCachedContent(caches)
	return rbxfile.ValueContent(res), err
}
func (b *extendedReader) readNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.readVarLengthString()
	return rbxfile.ValueBinaryString(res), err
}

// TODO: Remove context argument dependency
func (b *extendedReader) readNewEnumValue(enumID uint16, context *CommunicationContext) (rbxfile.ValueToken, error) {
	val, err := b.readUintUTF8()
	token := rbxfile.ValueToken{
		Value: val,
		ID:    enumID,
		Name:  getEnumName(context, enumID),
	}
	return token, err
}

func (b *extendedReader) readNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.readSintUTF8()
	return rbxfile.ValueInt(val), err
}

func getEnumName(context *CommunicationContext, id uint16) string {
	return context.StaticSchema.Enums[id].Name
}

// readNewTypeAndValue is never used by join data!
func (b *extendedReader) readNewTypeAndValue(reader PacketReader) (rbxfile.Value, error) {
	var val rbxfile.Value
	thisType, err := b.readUint8()
	if err != nil {
		return val, err
	}

	var enumID uint16
	if thisType == PROP_TYPE_ENUM {
		enumID, err = b.readUint16BE()
		if err != nil {
			return val, err
		}
	}

	val, err = b.ReadSerializedValue(reader, thisType, enumID)
	return val, err
}

func (b *extendedReader) readNewTuple(reader PacketReader) (rbxfile.ValueTuple, error) {
	var tuple rbxfile.ValueTuple
	tupleLen, err := b.readUintUTF8()
	if err != nil {
		return tuple, err
	}
	if tupleLen > 0x10000 {
		return tuple, errors.New("sanity check: exceeded maximum tuple len")
	}
	tuple = make(rbxfile.ValueTuple, tupleLen)
	for i := 0; i < int(tupleLen); i++ {
		val, err := b.readNewTypeAndValue(reader)
		if err != nil {
			return tuple, err
		}
		tuple[i] = val
	}

	return tuple, nil
}

func (b *extendedReader) readNewArray(reader PacketReader) (rbxfile.ValueArray, error) {
	array, err := b.readNewTuple(reader)
	return rbxfile.ValueArray(array), err
}

func (b *extendedReader) readNewDictionary(reader PacketReader) (rbxfile.ValueDictionary, error) {
	var dictionary rbxfile.ValueDictionary
	dictionaryLen, err := b.readUintUTF8()
	if err != nil {
		return dictionary, err
	}
	if dictionaryLen > 0x10000 {
		return dictionary, errors.New("sanity check: exceeded maximum dictionary len")
	}
	dictionary = make(rbxfile.ValueDictionary, dictionaryLen)
	for i := 0; i < int(dictionaryLen); i++ {
		keyLen, err := b.readUintUTF8()
		if err != nil {
			return dictionary, err
		}
		key, err := b.readASCII(int(keyLen))
		if err != nil {
			return dictionary, err
		}
		dictionary[key], err = b.readNewTypeAndValue(reader)
		if err != nil {
			return dictionary, err
		}
	}

	return dictionary, nil
}

func (b *extendedReader) readNewMap(reader PacketReader) (rbxfile.ValueMap, error) {
	thisMap, err := b.readNewDictionary(reader)
	return rbxfile.ValueMap(thisMap), err
}

func (b *extendedReader) readNumberSequenceKeypoint() (rbxfile.ValueNumberSequenceKeypoint, error) {
	var err error
	thisKeypoint := rbxfile.ValueNumberSequenceKeypoint{}
	thisKeypoint.Time, err = b.readFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Value, err = b.readFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Envelope, err = b.readFloat32BE()
	return thisKeypoint, err
}

func (b *extendedReader) readNumberSequence() (rbxfile.ValueNumberSequence, error) {
	var err error
	numKeypoints, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum numberseq len")
	}
	thisSequence := make(rbxfile.ValueNumberSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.readNumberSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *extendedReader) readNumberRange() (rbxfile.ValueNumberRange, error) {
	thisRange := rbxfile.ValueNumberRange{}
	var err error
	thisRange.Min, err = b.readFloat32BE()
	if err != nil {
		return thisRange, err
	}
	thisRange.Max, err = b.readFloat32BE()
	return thisRange, err
}

func (b *extendedReader) readColorSequenceKeypoint() (rbxfile.ValueColorSequenceKeypoint, error) {
	var err error
	thisKeypoint := rbxfile.ValueColorSequenceKeypoint{}
	thisKeypoint.Time, err = b.readFloat32BE()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Value, err = b.readColor3()
	if err != nil {
		return thisKeypoint, err
	}
	thisKeypoint.Envelope, err = b.readFloat32BE()
	return thisKeypoint, err
}

func (b *extendedReader) readColorSequence() (rbxfile.ValueColorSequence, error) {
	var err error
	numKeypoints, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum colorseq len")
	}
	thisSequence := make(rbxfile.ValueColorSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.readColorSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *extendedReader) readRect2D() (rbxfile.ValueRect2D, error) {
	var err error
	thisRect := rbxfile.ValueRect2D{rbxfile.ValueVector2{}, rbxfile.ValueVector2{}}

	thisRect.Min.X, err = b.readFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Min.Y, err = b.readFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Max.X, err = b.readFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Max.Y, err = b.readFloat32BE()
	return thisRect, err
}

func (b *extendedReader) readPhysicalProperties() (rbxfile.ValuePhysicalProperties, error) {
	var err error
	props := rbxfile.ValuePhysicalProperties{}
	props.CustomPhysics, err = b.readBoolByte()
	if props.CustomPhysics {
		if err != nil {
			return props, err
		}
		props.Density, err = b.readFloat32BE()
		if err != nil {
			return props, err
		}
		props.Friction, err = b.readFloat32BE()
		if err != nil {
			return props, err
		}
		props.Elasticity, err = b.readFloat32BE()
		if err != nil {
			return props, err
		}
		props.FrictionWeight, err = b.readFloat32BE()
		if err != nil {
			return props, err
		}
		props.ElasticityWeight, err = b.readFloat32BE()
	}

	return props, err
}

func (b *extendedReader) readCoordsMode0() (rbxfile.ValueVector3, error) {
	return b.readVector3Simple()
}
func (b *extendedReader) readCoordsMode1() (rbxfile.ValueVector3, error) {
	value := rbxfile.ValueVector3{}
	cRange, err := b.readFloat32BE()
	if err != nil {
		return value, err
	}
	if cRange <= 0.0000099999997 { // Has to be precise
		return rbxfile.ValueVector3{0, 0, 0}, nil
	}
	x, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	y, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	z, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	value.X = (float32(x)/32767.0 - 1.0) * cRange
	value.Y = (float32(y)/32767.0 - 1.0) * cRange
	value.Z = (float32(z)/32767.0 - 1.0) * cRange
	return value, nil
}
func (b *extendedReader) readCoordsMode2() (rbxfile.ValueVector3, error) {
	val := rbxfile.ValueVector3{}
	x, err := b.bits(15)
	if err != nil {
		return val, err
	}
	xShort := uint16((x >> 7) | ((x & 0x7F) << 8))
	y, err := b.bits(14)
	if err != nil {
		return val, err
	}
	yShort := uint16((y >> 6) | ((y & 0x3F) >> 8))
	z, err := b.bits(15)
	if err != nil {
		return val, err
	}
	zShort := uint16((z >> 7) | ((z & 0x7F) << 8))

	val.X = float32(xShort)*0.0625 - 2048.0
	val.Y = float32(yShort)*0.0625 - 2048.0
	val.Z = float32(zShort)*0.0625 - 2048.0
	return val, nil
}

func (b *extendedReader) readPhysicsCoords() (rbxfile.ValueVector3, error) {
	var val rbxfile.ValueVector3
	flags, err := b.readUint8()
	if err != nil {
		return val, err
	}
	exp := flags >> 3 // void signs
	scale := float32(math.Exp2(float64(exp)))

	// nice bitpacking!
	if exp > 4 {
		if exp > 10 { // when scale is > 0x400
			// minimum precision is 2^11/2^21 = 2^-10
			// 21 bits per value
			val1, err := b.readUint32BE()
			if err != nil {
				return val, err
			}
			val2, err := b.readUint32BE()
			if err != nil {
				return val, err
			}
			xSign := uint32((flags >> 2) & 1)
			ySign := uint32((flags >> 1) & 1)
			zSign := uint32((flags >> 0) & 1)

			v22 := (xSign << 21) | (val1 >> 11)
			v25 := int32(v22<<10) >> 10
			v26 := int32(((val2>>21)|(((ySign<<11)|val1&0x7FF)<<10))<<10) >> 10
			v27 := val2&0x1FFFFF | (zSign << 31 >> 10)
			val.X = float32(float32(v25)*0.00000047683739) * scale
			val.Y = float32(float32(v26)*0.00000047683739) * scale
			val.Y = float32(float32(v27)*0.00000047683739) * scale
		} else { // when scale is in range 0x10 < scale <= 0x400
			// 16 bits per value (and sign), precision ranges from 1/4096 to 1/128
			// 2^16 = 65536, so values range from -16~16 to -512~512
			x, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			y, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			z, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			xSign := uint32((flags >> 2) & 1)
			ySign := uint32((flags >> 1) & 1)
			zSign := uint32((flags >> 0) & 1)
			rx := uint32(x)
			ry := uint32(y)
			rz := uint32(z)

			v20 := int32((ry|(ySign<<16))<<15) >> 15
			v21 := int32(((rz)|(zSign<<16))<<15) >> 15 // for some reason IDA doesn't show this as signed, despite that fact that it's created by a sar instruction!
			val.X = float32(float32(int32((rx|(xSign<<16))<<15)>>15) * 0.000015259022 * scale)
			val.Y = float32(float32(v20) * 0.000015259022 * scale)
			val.Z = float32(float32(v21) * 0.000015259022 * scale)
		}
	} else { // when scale is in range 0 <= scale <= 0x10
		val1, err := b.readUint32BE()
		if err != nil {
			return val, err
		}
		xSign := uint32((flags >> 2) & 1)
		ySign := uint32((flags >> 1) & 1)
		zSign := uint32((flags >> 0) & 1)

		res := (int32(ySign<<31) >> 21) | int32((val1>>10)&0x3FF)
		v17 := int32(val1&0x3FF) | (int32(zSign<<31) >> 21)
		val.X = float32(float32(int32(((xSign<<10)|(val1>>20))<<21)>>21)*0.00097751711) * scale
		val.Y = float32(float32(res)*0.00097751711) * scale
		val.Z = float32(float32(v17)*0.00097751711) * scale
	}
	return val, err
}

func (b *extendedReader) readMatrixMode0() ([9]float32, error) {
	var val [9]float32
	var err error

	q := [4]float32{}
	q[3], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[0], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[1], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[2], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}

	return quaternionToRotMatrix(q), nil
}

func (b *extendedReader) readMatrixMode1() ([9]float32, error) {
	q := [4]float32{}
	var val [9]float32
	invertW, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertX, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertY, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertZ, err := b.readBool()
	if err != nil {
		return val, err
	}
	x, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	y, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	z, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	xs := float32(x) / 65535.0
	ys := float32(y) / 65535.0
	zs := float32(z) / 65535.0
	if invertX {
		xs = -xs
	}
	if invertY {
		ys = -ys
	}
	if invertZ {
		zs = -zs
	}
	w := float32(math.Sqrt(math.Max(0.0, float64(1.0-xs-ys-zs))))
	if invertW {
		w = -w
	}
	q = [4]float32{xs, ys, zs, w}
	return quaternionToRotMatrix(q), nil
}
func (b *extendedReader) readMatrixMode2() ([9]float32, error) {
	return b.readMatrixMode1()
}

var quaternionIndices = [4][3]int{
	// the index is the number that is omitted
	[3]int{1, 2, 3}, // index 0
	[3]int{0, 2, 3}, // index 1
	[3]int{0, 1, 3}, // index 2
	[3]int{0, 1, 2}, // index 3
}

func (b *extendedReader) readPhysicsMatrix() ([9]float32, error) {
	var quat [4]float32
	var err error

	val1, err := b.readUint16BE()
	if err != nil {
		return [9]float32{}, err
	}
	val2, err := b.readUint32BE()
	if err != nil {
		return [9]float32{}, err
	}
	mode := val2 >> 0x1E
	indexSet := quaternionIndices[mode]

	//		16 bits for val 1
	// +	2 bits for mode from val 2
	// +	17 bits from val 2
	// +	17 bits from val 2
	//		52 bits
	xmm4 := float32(int32(uint32(val1)<<17) >> 17)
	xmm4 = xmm4 / (16383.0 * math.Sqrt2)
	xmm3 := float32(int32(val2<<2) >> 17)
	xmm3 = xmm3 / (16383.0 * math.Sqrt2)
	xmm2 := float32(int32(val2<<17) >> 17)
	xmm2 = xmm2 / (16383.0 * math.Sqrt2)

	quat[indexSet[0]] = xmm4
	quat[indexSet[1]] = xmm3
	quat[indexSet[2]] = xmm2
	quat[mode] = float32(math.Sqrt(math.Max(0.0, float64(1.0-xmm4*xmm4-xmm3*xmm3-xmm2*xmm2))))

	return quaternionToRotMatrix(quat), err
}

func (b *extendedReader) readPhysicsCFrame() (rbxfile.ValueCFrame, error) {
	var val rbxfile.ValueCFrame
	coords, err := b.readPhysicsCoords()
	if err != nil {
		return val, err
	}
	matrix, err := b.readPhysicsMatrix()
	if err != nil {
		return val, err
	}
	return rbxfile.ValueCFrame{coords, matrix}, nil
}

/* code to convert compact cf to real:
fCos := math.Cos(realAngleRadians)
fSin := math.Sin(realAngleRadians)
fOneMinusCos := 1 - fCos

fX2 := unitVector.X ** 2
fY2 := unitVector.Y ** 2
fZ2 := unitVector.Z ** 2
fXYM := unitVector.X * unitVector.Y * fOneMinusCos
fXZM := unitVector.X * unitVector.Z * fOneMinusCos
fYZM := unitVector.Y * unitVector.Z * fOneMinusCos
fXSin := unitVector.X * fSin
fYSin := unitVector.Y * fSin
fZSin := unitVector.Z * fSin

return PhysicsMotor{rbxfile.ValueVector3{}, [9]float32{
	fX2 * fOneMinusCos + fCos,
	fXYM - fZSin,
	fXZM + fYSin,

	fXYM + fZSin,
	fY2 * fOneMinusCos + fCos,
	FYZM + fXSin,

	fXZM - fYSin,
	fYZM + fXSin,
	fZ2 * fOneMinusCos + fCos,
}}
*/
func (b *extendedReader) readPhysicsMotor() (PhysicsMotor, error) {
	var motor PhysicsMotor
	flags, err := b.readUint8()
	if err != nil {
		return motor, err
	}

	if flags&1 == 1 {
		motor.Position, err = b.readPhysicsCoords()
		if err != nil {
			return motor, err
		}
	}

	if flags&2 == 2 {
		var quat [4]float32

		val1, err := b.readUint32BE()
		if err != nil {
			return motor, err
		}
		mode := val1 >> 0x1E
		indexSet := quaternionIndices[mode]

		xmm4 := float32(int32(uint32(val1)<<2) >> 22)
		xmm4 = xmm4 / (511.0 * math.Sqrt2)
		xmm3 := float32(int32(val1<<12) >> 22)
		xmm3 = xmm3 / (511.0 * math.Sqrt2)
		xmm2 := float32(int32(val1<<22) >> 22)
		xmm2 = xmm2 / (511.0 * math.Sqrt2)

		quat[indexSet[0]] = xmm4
		quat[indexSet[1]] = xmm3
		quat[indexSet[2]] = xmm2
		quat[mode] = float32(math.Sqrt(math.Max(0.0, float64(1.0-xmm4*xmm4-xmm3*xmm3-xmm2*xmm2))))

		motor.Rotation = quaternionToRotMatrix(quat)
	}

	return motor, nil
}

func (b *extendedReader) readMotors() ([]PhysicsMotor, error) {
	countMotors, err := b.readVarint64()
	if err != nil {
		return nil, err
	}
	//println("reading", countMotors, "motors")
	if countMotors > 0x10000 {
		return nil, errors.New("numMotors is excessive")
	}

	motors := make([]PhysicsMotor, countMotors)
	var i uint64
	for i = 0; i < countMotors; i++ {
		motors[i], err = b.readPhysicsMotor()
		if err != nil {
			return motors, err
		}
	}
	return motors, nil
}

func (b *extendedReader) readInt64() (rbxfile.ValueInt64, error) {
	val, err := b.readVarsint64()
	return rbxfile.ValueInt64(val), err
}

func (b *extendedReader) readPhysicsVelocity() (rbxfile.ValueVector3, error) {
	var val rbxfile.ValueVector3
	flags, err := b.readUint8()
	if err != nil || flags == 0 {
		return val, err
	}
	exp := flags>>4 - 1
	scale := float32(math.Exp2(float64(exp)))
	val1 := flags & 0xF
	val2, err := b.readUint32BE()
	if err != nil {
		return val, err
	}
	res := int32(val2<<12) >> 20
	v9 := int32(val1) | (int32(val2<<24) >> 20)
	val.X = float32(int32(val2)>>20) * 0.00048851978 * scale
	val.Y = float32(res) * 0.00048851978 * scale
	val.Z = float32(v9) * 0.00048851978 * scale

	return val, err
}

func (b *extendedReader) readSerializedValueGeneric(reader PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	var temp string
	switch valueType {
	case PROP_TYPE_INVALID: // I assume this is how it works, anyway
		result = nil
		err = nil
	case PROP_TYPE_STRING_NO_CACHE:
		temp, err = b.readVarLengthString()
		result = rbxfile.ValueString(temp)
	case PROP_TYPE_ENUM:
		result, err = b.readNewEnumValue(enumId, reader.Context())
	case PROP_TYPE_BINARYSTRING:
		result, err = b.readNewBinaryString()
	case PROP_TYPE_PBOOL:
		result, err = b.readPBool()
	case PROP_TYPE_PSINT:
		result, err = b.readNewPSint()
	case PROP_TYPE_PFLOAT:
		result, err = b.readPFloat()
	case PROP_TYPE_PDOUBLE:
		result, err = b.readPDouble()
	case PROP_TYPE_UDIM:
		result, err = b.readUDim()
	case PROP_TYPE_UDIM2:
		result, err = b.readUDim2()
	case PROP_TYPE_RAY:
		result, err = b.readRay()
	case PROP_TYPE_FACES:
		result, err = b.readFaces()
	case PROP_TYPE_AXES:
		result, err = b.readAxes()
	case PROP_TYPE_BRICKCOLOR:
		result, err = b.readBrickColor()
	case PROP_TYPE_COLOR3:
		result, err = b.readColor3()
	case PROP_TYPE_COLOR3UINT8:
		result, err = b.readColor3uint8()
	case PROP_TYPE_VECTOR2:
		result, err = b.readVector2()
	case PROP_TYPE_VECTOR3_SIMPLE:
		result, err = b.readVector3Simple()
	case PROP_TYPE_VECTOR3_COMPLICATED:
		result, err = b.readVector3()
	case PROP_TYPE_VECTOR2UINT16:
		result, err = b.readVector2int16()
	case PROP_TYPE_VECTOR3UINT16:
		result, err = b.readVector3int16()
	case PROP_TYPE_CFRAME_SIMPLE:
		result, err = b.readCFrameSimple()
	case PROP_TYPE_CFRAME_COMPLICATED:
		result, err = b.readCFrame()
	case PROP_TYPE_NUMBERSEQUENCE:
		result, err = b.readNumberSequence()
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		result, err = b.readNumberSequenceKeypoint()
	case PROP_TYPE_NUMBERRANGE:
		result, err = b.readNumberRange()
	case PROP_TYPE_COLORSEQUENCE:
		result, err = b.readColorSequence()
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		result, err = b.readColorSequenceKeypoint()
	case PROP_TYPE_RECT2D:
		result, err = b.readRect2D()
	case PROP_TYPE_PHYSICALPROPERTIES:
		result, err = b.readPhysicalProperties()
	case PROP_TYPE_REGION3:
		result, err = b.readRegion3()
	case PROP_TYPE_REGION3INT16:
		result, err = b.readRegion3int16()
	case PROP_TYPE_INT64:
		result, err = b.readInt64()
	}
	return result, err
}
