package peer

import (
	"errors"
	"fmt"
	"math"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

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
	val.Offset = int32(off)
	return val, err
}

func (b *extendedReader) readUDim2() (rbxfile.ValueUDim2, error) {
	var err error
	val := rbxfile.ValueUDim2{}
	val.X.Scale, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	offx, err := b.readUint32BE()
	val.X.Offset = int32(offx)
	if err != nil {
		return val, err
	}
	val.Y.Scale, err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	offy, err := b.readUint32BE()
	val.Y.Offset = int32(offy)
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

func (b *extendedReader) readRegion3() (datamodel.ValueRegion3, error) {
	var err error
	val := datamodel.ValueRegion3{}
	val.Start, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}
	val.End, err = b.readVector3Simple()
	return val, err
}
func (b *extendedReader) readRegion3int16() (datamodel.ValueRegion3int16, error) {
	var err error
	val := datamodel.ValueRegion3int16{}
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

// reads a offline Vector3 value (f32 X, f32 Y, f32 Z)
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
	val := rbxfile.ValueVector3{}
	/*isInteger, err := b.readBool()
	if err != nil {
		return rbxfile.ValueVector3{}, err
	}
	if !isInteger {
		return b.readVector3Simple()
	}
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
	val.Z = float32(int16(zShort)) * 0.5*/
	return val, errors.New("v3comp not implemented")
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
	val = rbxfile.ValueVector3int16{X: int16(valX), Y: int16(valY), Z: int16(valZ)}
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

func (b *extendedReader) readObject(context *CommunicationContext, caches *Caches) (datamodel.Reference, error) {
	if context.ServerPeerID != 0 {
		return b.readObjectPeerID(context)
	}
	scope, err := b.readCachedScope(caches)
	if err != nil && err != ErrCacheReadOOB { // TODO: hack! physics packets may have problems with caches
		return datamodel.Reference{}, err
	}
	reference := datamodel.Reference{Scope: scope}
	if scope != "NULL" {
		reference.Id, err = b.readUint32LE()
	} else {
		reference.IsNull = true
	}

	return reference, err
}

func (b *extendedReader) readCFrameSimple() (rbxfile.ValueCFrame, error) {
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
		for i := 0; i < 9; i++ {
			val.Rotation[i], err = b.readFloat32BE()
			if err != nil {
				return val, err
			}
		}
	}

	return val, nil
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

var specialColumns = [6][3]float32{
	[3]float32{1, 0, 0},
	[3]float32{0, 1, 0},
	[3]float32{0, 0, 1},
	[3]float32{-1, 0, 0},
	[3]float32{0, -1, 0},
	[3]float32{0, 0, -1},
}

func lookupRotMatrix(special uint64) [9]float32 {
	column0 := specialColumns[special/6]
	column1 := specialColumns[special%6]

	ret := [9]float32{
		column0[0], column1[0], 0,
		column0[1], column1[1], 0,
		column0[2], column1[2], 0,
	}
	ret[2] = column0[1]*column1[2] - column1[1]*column0[2]
	ret[5] = column1[0]*column0[2] - column0[0]*column1[2]
	ret[8] = column0[0]*column1[1] - column1[0]*column0[1]

	return ret
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

func (b *joinSerializeReader) readContent() (rbxfile.ValueContent, error) {
	var result string
	stringLen, err := b.readUint32BE()
	if err != nil {
		return rbxfile.ValueContent(result), err
	}
	result, err = b.readASCII(int(stringLen))
	return rbxfile.ValueContent(result), err
}

// TODO: Make this function uniform with other cache functions
func (b *extendedReader) readSystemAddress() (datamodel.ValueSystemAddress, error) {
	val, err := b.readVarint64()
	if err != nil {
		return datamodel.ValueSystemAddress(0), err
	}
	thisAddress := datamodel.ValueSystemAddress(val)
	return thisAddress, err
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

func (b *extendedReader) readLuauProtectedStringRaw() (datamodel.ValueSignedProtectedString, error) {
	stringLen, err := b.readUintUTF8()
	if err != nil {
		return datamodel.ValueSignedProtectedString{}, err
	}

	val, err := b.readString(int(stringLen))
	if err != nil {
		return datamodel.ValueSignedProtectedString{}, err
	}

	extraBytesLen, err := b.readUintUTF8()
	if err != nil {
		return datamodel.ValueSignedProtectedString{}, err
	}
	extraBytes, err := b.readString(int(extraBytesLen))
	if err != nil {
		return datamodel.ValueSignedProtectedString{}, err
	}

	return datamodel.ValueSignedProtectedString{
		Value:     val,
		Signature: extraBytes,
	}, nil
}

func (b *joinSerializeReader) readNewPString() (rbxfile.ValueString, error) {
	val, err := b.readVarLengthString()
	return rbxfile.ValueString(val), err
}
func (b *extendedReader) readNewPString(caches *Caches) (rbxfile.ValueString, error) {
	val, err := b.readCached(caches)
	return rbxfile.ValueString(val), err
}

func (b *joinSerializeReader) readNewProtectedString() (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewPString()
	return rbxfile.ValueProtectedString(res), err
}
func (b *extendedReader) readNewProtectedString(caches *Caches) (rbxfile.ValueProtectedString, error) {
	res, err := b.readNewCachedProtectedString(caches)
	return rbxfile.ValueProtectedString(res), err
}

func (b *joinSerializeReader) readLuauProtectedString() (datamodel.ValueSignedProtectedString, error) {
	return b.readLuauProtectedStringRaw()
}
func (b *extendedReader) readLuauProtectedString(caches *Caches) (datamodel.ValueSignedProtectedString, error) {
	res, err := b.readLuauCachedProtectedString(caches)
	return res, err
}

func (b *joinSerializeReader) readNewContent() (rbxfile.ValueContent, error) {
	res, err := b.readNewPString()
	return rbxfile.ValueContent(res), err
}
func (b *extendedReader) readNewContent(context *CommunicationContext) (rbxfile.ValueContent, error) {
	baseId, err := b.readUint8()
	if err != nil {
		return rbxfile.ValueContent(""), err
	}
	base := context.NetworkSchema.ContentPrefixes[baseId>>1]
	if baseId&1 == 0 {
		res, err := b.readVarLengthString()
		if err != nil {
			return rbxfile.ValueContent(""), err
		}
		return rbxfile.ValueContent(base + res), nil
	}
	// numeric id
	number, err := b.readVarsint64()
	if err != nil {
		return rbxfile.ValueContent(""), err
	}
	return rbxfile.ValueContent(fmt.Sprintf("%s%d", base, number)), nil
}
func (b *extendedReader) readNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.readVarLengthString()
	return rbxfile.ValueBinaryString(res), err
}

// TODO: Remove context argument dependency
func (b *extendedReader) readNewEnumValue(enumID uint16, context *CommunicationContext) (datamodel.ValueToken, error) {
	val, err := b.readUintUTF8()
	token := datamodel.ValueToken{Value: val, ID: enumID}
	return token, err
}

func (b *extendedReader) readNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.readSintUTF8()
	return rbxfile.ValueInt(val), err
}

func getEnumName(context *CommunicationContext, id uint16) string {
	return context.NetworkSchema.Enums[id].Name
}

// readNewTypeAndValue is never used by join data!
func (b *extendedReader) readNewTypeAndValue(reader PacketReader, deferred deferredStrings) (rbxfile.Value, error) {
	var val rbxfile.Value
	thisType, err := b.readUint8()
	if err != nil {
		return val, err
	}

	var enumID uint16
	if thisType == PropertyTypeEnum {
		enumID, err = b.readUint16BE()
		if err != nil {
			return val, err
		}
	}

	val, err = b.ReadSerializedValue(reader, thisType, enumID, deferred)
	return val, err
}

func (b *extendedReader) readNewTuple(reader PacketReader, deferred deferredStrings) (datamodel.ValueTuple, error) {
	var tuple datamodel.ValueTuple
	tupleLen, err := b.readUintUTF8()
	if err != nil {
		return tuple, err
	}
	if tupleLen > 0x10000 {
		return tuple, errors.New("sanity check: exceeded maximum tuple len")
	}
	tuple = make(datamodel.ValueTuple, tupleLen)
	for i := 0; i < int(tupleLen); i++ {
		val, err := b.readNewTypeAndValue(reader, deferred)
		if err != nil {
			return tuple, err
		}
		tuple[i] = val
	}

	return tuple, nil
}

func (b *extendedReader) readNewArray(reader PacketReader, deferred deferredStrings) (datamodel.ValueArray, error) {
	array, err := b.readNewTuple(reader, deferred)
	return datamodel.ValueArray(array), err
}

func (b *extendedReader) readNewDictionary(reader PacketReader, deferred deferredStrings) (datamodel.ValueDictionary, error) {
	var dictionary datamodel.ValueDictionary
	dictionaryLen, err := b.readUintUTF8()
	if err != nil {
		return dictionary, err
	}
	if dictionaryLen > 0x10000 {
		return dictionary, errors.New("sanity check: exceeded maximum dictionary len")
	}
	dictionary = make(datamodel.ValueDictionary, dictionaryLen)
	for i := 0; i < int(dictionaryLen); i++ {
		keyLen, err := b.readUintUTF8()
		if err != nil {
			return dictionary, err
		}
		key, err := b.readASCII(int(keyLen))
		if err != nil {
			return dictionary, err
		}
		dictionary[key], err = b.readNewTypeAndValue(reader, deferred)
		if err != nil {
			return dictionary, err
		}
	}

	return dictionary, nil
}

func (b *extendedReader) readNewMap(reader PacketReader, deferred deferredStrings) (datamodel.ValueMap, error) {
	thisMap, err := b.readNewDictionary(reader, deferred)
	return datamodel.ValueMap(thisMap), err
}

func (b *extendedReader) readNumberSequenceKeypoint() (datamodel.ValueNumberSequenceKeypoint, error) {
	var err error
	thisKeypoint := datamodel.ValueNumberSequenceKeypoint{}
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

func (b *extendedReader) readNumberSequence() (datamodel.ValueNumberSequence, error) {
	var err error
	numKeypoints, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum numberseq len")
	}
	thisSequence := make(datamodel.ValueNumberSequence, numKeypoints)

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

func (b *extendedReader) readColorSequenceKeypoint() (datamodel.ValueColorSequenceKeypoint, error) {
	var err error
	thisKeypoint := datamodel.ValueColorSequenceKeypoint{}
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

func (b *extendedReader) readColorSequence() (datamodel.ValueColorSequence, error) {
	var err error
	numKeypoints, err := b.readUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum colorseq len")
	}
	thisSequence := make(datamodel.ValueColorSequence, numKeypoints)

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
	thisRect := rbxfile.ValueRect2D{}

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
		return rbxfile.ValueVector3{}, nil
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
	/*x, err := b.bits(15)
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
	val.Z = float32(zShort)*0.0625 - 2048.0*/
	return val, errors.New("coordmode 2 not implemented")
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
	/*invertW, err := b.readBool()
	var val [9]float32
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
	q = [4]float32{xs, ys, zs, w}*/
	return quaternionToRotMatrix(q), errors.New("matrixmode1 not implemented")
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
	return rbxfile.ValueCFrame{Position: coords, Rotation: matrix}, nil
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

func (b *extendedReader) readPathWaypoint() (datamodel.ValuePathWaypoint, error) {
	var val datamodel.ValuePathWaypoint
	var err error
	val.Position, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}

	val.Action, err = b.readUint32BE()
	return val, err
}

func (b *extendedReader) readSharedString(deferred deferredStrings) (*datamodel.ValueDeferredString, error) {
	md5, err := b.readASCII(0x10)
	if err != nil {
		return nil, err
	}

	return deferred.NewValue(md5), nil
}

func (b *extendedReader) readDateTime() (datamodel.ValueDateTime, error) {
	var out datamodel.ValueDateTime
	val, err := b.readUint64BE()
	if err != nil {
		return out, err
	}
	out.UnixMilliseconds = val
	return out, nil
}

func (b *extendedReader) readOptimizedString(context *CommunicationContext) (rbxfile.ValueString, error) {
	header, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	if header <= 0x7F {
		length := int(header)
		if header == 0x7F {
			newLength, err := b.readUintUTF8()
			if err != nil {
				return nil, err
			}
			length = int(newLength)
		}
		str, err := b.readASCII(length)
		if err != nil {
			return nil, err
		}
		return rbxfile.ValueString(str), nil
	}

	header &= 0x7F
	presharedId := int(header)
	if header == 0x7F {
		newId, err := b.readUintUTF8()
		if err != nil {
			return nil, err
		}
		presharedId = int(newId)
	}
	if presharedId >= len(context.NetworkSchema.OptimizedStrings) {
		return nil, errors.New("preshared string id oob")
	}
	return rbxfile.ValueString(context.NetworkSchema.OptimizedStrings[presharedId]), nil
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

func (b *extendedReader) readSerializedValueGeneric(reader PacketReader, valueType uint8, enumID uint16, deferred deferredStrings) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	var temp string
	switch valueType {
	case PropertyTypeNil: // I assume this is how it works, anyway
		result = nil
		err = nil
	case PropertyTypeStringNoCache:
		temp, err = b.readVarLengthString()
		result = rbxfile.ValueString(temp)
	case PropertyTypeEnum:
		result, err = b.readNewEnumValue(enumID, reader.Context())
	case PropertyTypeBinaryString:
		result, err = b.readNewBinaryString()
	case PropertyTypeBool:
		result, err = b.readPBool()
	case PropertyTypeInt:
		result, err = b.readNewPSint()
	case PropertyTypeFloat:
		result, err = b.readPFloat()
	case PropertyTypeDouble:
		result, err = b.readPDouble()
	case PropertyTypeUDim:
		result, err = b.readUDim()
	case PropertyTypeUDim2:
		result, err = b.readUDim2()
	case PropertyTypeRay:
		result, err = b.readRay()
	case PropertyTypeFaces:
		result, err = b.readFaces()
	case PropertyTypeAxes:
		result, err = b.readAxes()
	case PropertyTypeBrickColor:
		result, err = b.readBrickColor()
	case PropertyTypeColor3:
		result, err = b.readColor3()
	case PropertyTypeColor3uint8:
		result, err = b.readColor3uint8()
	case PropertyTypeVector2:
		result, err = b.readVector2()
	case PropertyTypeSimpleVector3:
		result, err = b.readVector3Simple()
	case PropertyTypeComplicatedVector3:
		result, err = b.readVector3()
	case PropertyTypeVector2int16:
		result, err = b.readVector2int16()
	case PropertyTypeVector3int16:
		result, err = b.readVector3int16()
	case PropertyTypeSimpleCFrame:
		result, err = b.readCFrameSimple()
	case PropertyTypeComplicatedCFrame:
		result, err = b.readCFrame()
	case PropertyTypeSystemAddress:
		result, err = b.readSystemAddress()
	case PropertyTypeNumberSequence:
		result, err = b.readNumberSequence()
	case PropertyTypeNumberSequenceKeypoint:
		result, err = b.readNumberSequenceKeypoint()
	case PropertyTypeNumberRange:
		result, err = b.readNumberRange()
	case PropertyTypeColorSequence:
		result, err = b.readColorSequence()
	case PropertyTypeColorSequenceKeypoint:
		result, err = b.readColorSequenceKeypoint()
	case PropertyTypeRect2D:
		result, err = b.readRect2D()
	case PropertyTypePhysicalProperties:
		result, err = b.readPhysicalProperties()
	case PropertyTypeRegion3:
		result, err = b.readRegion3()
	case PropertyTypeRegion3int16:
		result, err = b.readRegion3int16()
	case PropertyTypeInt64:
		result, err = b.readInt64()
	case PropertyTypePathWaypoint:
		result, err = b.readPathWaypoint()
	case PropertyTypeSharedString:
		result, err = b.readSharedString(deferred)
	case PropertyTypeDateTime:
		result, err = b.readDateTime()
	case PropertyTypeOptimizedString:
		result, err = b.readOptimizedString(reader.Context())
	default:
		err = fmt.Errorf("unsupported value type %d", valueType)
	}
	return result, err
}
