package peer
import "net"
import "errors"
import "github.com/gskartwii/rbxfile"
import "fmt"
import "math"

type Referent string
type PhysicsMotor struct {
	Coords1 rbxfile.ValueVector3
	Coords2 rbxfile.ValueVector3
	Angle uint8
}

func (b *ExtendedReader) ReadUDim() (rbxfile.ValueUDim, error) {
	var err error
	val := rbxfile.ValueUDim{}
	val.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	off, err := b.ReadUint32BE()
	val.Offset = int16(off)
	return val, err
}

func (b *ExtendedReader) ReadUDim2() (rbxfile.ValueUDim2, error) {
	var err error
	val := rbxfile.ValueUDim2{rbxfile.ValueUDim{}, rbxfile.ValueUDim{}}
	val.X.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	offx, err := b.ReadUint32BE()
	val.X.Offset = int16(offx)
	if err != nil {
		return val, err
	}
	val.Y.Scale, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	offy, err := b.ReadUint32BE()
	val.Y.Offset = int16(offy)
	return val, err
}

func (b *ExtendedReader) ReadRay() (rbxfile.ValueRay, error) {
	var err error
	val := rbxfile.ValueRay{}
	val.Origin, err = b.ReadVector3Simple()
	if err != nil {
		return val, err
	}
	val.Direction, err = b.ReadVector3Simple()
	return val, err
}

func (b *ExtendedReader) ReadRegion3() (rbxfile.ValueRegion3, error) {
	var err error
	val := rbxfile.ValueRegion3{}
	val.Start, err = b.ReadVector3Simple()
	if err != nil {
		return val, err
	}
	val.End, err = b.ReadVector3Simple()
	return val, err
}
func (b *ExtendedReader) ReadRegion3int16() (rbxfile.ValueRegion3int16, error) {
	var err error
	val := rbxfile.ValueRegion3int16{}
	val.Start, err = b.ReadVector3int16()
	if err != nil {
		return val, err
	}
	val.End, err = b.ReadVector3int16()
	return val, err
}

func (b *ExtendedReader) ReadColor3() (rbxfile.ValueColor3, error) {
	var err error
	val := rbxfile.ValueColor3{}
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

func (b *ExtendedReader) ReadColor3uint8() (rbxfile.ValueColor3uint8, error) {
	var err error
	val := rbxfile.ValueColor3uint8{}
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

func (b *ExtendedReader) ReadVector2() (rbxfile.ValueVector2, error) {
	var err error
	val := rbxfile.ValueVector2{}
	val.X, err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	val.Y, err = b.ReadFloat32BE()
	return val, err
}

func (b *ExtendedReader) ReadVector3Simple() (rbxfile.ValueVector3, error) {
	var err error
	val := rbxfile.ValueVector3{}
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

func (b *ExtendedReader) ReadVector3() (rbxfile.ValueVector3, error) {
	isInteger, err := b.ReadBool()
	if err != nil {
		return rbxfile.ValueVector3{}, err
	}
	if !isInteger {
		return b.ReadVector3Simple()
	} else {
		var err error
		val := rbxfile.ValueVector3{}
		x, err := b.Bits(11)
		if err != nil {
			return val, err
		}
		x_short := uint16((x >> 3) | ((x & 7) << 8))
		if x_short & 0x400 != 0 {
			x_short |= 0xFC00
		}

		y, err := b.Bits(11)
		if err != nil {
			return val, err
		}
		y_short := uint16((y >> 3) | ((y & 7) << 8))

		z, err := b.Bits(11)
		if err != nil {
			return val, err
		}

		z_short := uint16((z >> 3) | ((z & 7) << 8))
		if z_short & 0x400 != 0 {
			z_short |= 0xFC00
		}

		val.X = float32(int16(x_short)) * 0.5
		val.Y = float32(y_short) * 0.1
		val.Z = float32(int16(z_short)) * 0.5
		return val, err
	}
}

func (b *ExtendedReader) ReadVector2int16() (rbxfile.ValueVector2int16, error) {
	var err error
	val := rbxfile.ValueVector2int16{}
    valX, err := b.ReadUint16BE()
	if err != nil {
		return val, err
	}
    valY, err := b.ReadUint16BE()
    val.X = int16(valX)
    val.Y = int16(valY)
	return val, err
}

func (b *ExtendedReader) ReadVector3int16() (rbxfile.ValueVector3int16, error) {
	var err error
    var val rbxfile.ValueVector3int16
    valX, err := b.ReadUint16BE()
	if err != nil {
		return val, err
	}
    valY, err := b.ReadUint16BE()
	if err != nil {
		return val, err
	}
    valZ, err := b.ReadUint16BE()
	val = rbxfile.ValueVector3int16{int16(valX), int16(valY), int16(valZ)}
	return val, err
}

func (b *ExtendedReader) ReadPBool() (rbxfile.ValueBool, error) {
	val, err := b.ReadBool()
	return rbxfile.ValueBool(val), err
}
func (b *ExtendedReader) ReadPSInt() (rbxfile.ValueInt, error) {
	val, err := b.ReadUint32BE()
	return rbxfile.ValueInt(val), err
}
func (b *ExtendedReader) ReadPFloat() (rbxfile.ValueFloat, error) {
	val, err := b.ReadFloat32BE()
	return rbxfile.ValueFloat(val), err
}
func (b *ExtendedReader) ReadPDouble() (rbxfile.ValueDouble, error) {
	val, err := b.ReadFloat64BE()
	return rbxfile.ValueDouble(val), err
}
func (b *ExtendedReader) ReadAxes() (rbxfile.ValueAxes, error) {
	val, err := b.ReadUint32BE()
    axesVal := rbxfile.ValueAxes{
        X: val & 4 != 0,
        Y: val & 2 != 0,
        Z: val & 1 != 0,
    }
	return axesVal, err
}
func (b *ExtendedReader) ReadFaces() (rbxfile.ValueFaces, error) {
	val, err := b.ReadUint32BE()
    facesVal := rbxfile.ValueFaces{
        Right: val & 32 != 0,
        Top: val & 16 != 0,
        Back: val & 8 != 0,
        Left: val & 4 != 0,
        Bottom: val & 2 != 0,
        Front: val & 1 != 0,
    }
	return facesVal, err
}
func (b *ExtendedReader) ReadBrickColor() (rbxfile.ValueBrickColor, error) {
	val, err := b.Bits(7)
	return rbxfile.ValueBrickColor(val), err
}

func formatBindable(obj rbxfile.ValueReference) string {
	return obj.String()
}
func objectToRef(referent string, referentInt uint32) string {
	if referentInt == 0 {
		return "null"
	} else {
		return fmt.Sprintf("%d", referentInt)
	}
}

func (b *ExtendedReader) ReadObject(isJoinData bool, context *CommunicationContext) (Referent, error) {
	var err error
    var referent string
    var referentInt uint32
    var Object Referent
	if isJoinData {
		referent, referentInt, err = b.ReadJoinReferent()
	} else {
		referent, err = b.ReadCachedObject(context)
		if err != nil {
			return Object, err
		}
		if referent != "" {
			referentInt, err = b.ReadUint32LE()
		}
	}

    serialized := objectToRef(referent, referentInt)
    context.RefStringsByReferent[referent] = referent

	return Referent(serialized), err
}
func (b *ExtendedReader) ReadCFrameSimple() (rbxfile.ValueCFrame, error) {
	return rbxfile.ValueCFrame{}, nil // nop for now, since nothing uses this
}

func quaternionToRotMatrix(q [4]float32) [9]float32 {
    X  := q[0]
    Y  := q[1]
    Z  := q[2]
    W  := q[3]
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
        1-(YY+ZZ),
        XY-WZ,
        XZ+WY,
        XY+WZ,
        1-(XX+ZZ),
        YZ-WX,
        XZ-WY,
        YZ+WX,
        1-(XX+YY),
    }
}

func transformQuaternionToMatrix(q [4]float32) [9]float32 {
	midresult := quaternionToRotMatrix(q)

	xScaleFactor := float32(1.0/math.Sqrt(float64(midresult[0]*midresult[0] + midresult[3]*midresult[3] + midresult[6]*midresult[6])))

	result := [9]float32{
		xScaleFactor * midresult[0], 0, 0,
		xScaleFactor * midresult[3], 0, 0,
		xScaleFactor * midresult[6], 0, 0,
	} // X has been normalized

	sx_ySum := result[0]*midresult[1] + result[3]*midresult[4] + result[6]*midresult[7]
	trueR01 := midresult[1] - result[0]*sx_ySum
	trueR11 := midresult[4] - result[3]*sx_ySum
	trueR21 := midresult[7] - result[6]*sx_ySum

	yScaleFactor := float32(1.0/math.Sqrt(float64(trueR01*trueR01 + trueR11*trueR11 + trueR21*trueR21)))
	result[1] = trueR01 * yScaleFactor
	result[4] = trueR11 * yScaleFactor
	result[7] = trueR21 * yScaleFactor

	sy_zSum := result[1]*midresult[2] + result[4]*midresult[5] + result[7]*midresult[8]
	sx_zSum := result[0]*midresult[2] + result[3]*midresult[5] + result[6]*midresult[8]
	trueR02 := midresult[2] - result[0]*sx_zSum - result[1]*sy_zSum
	trueR12 := midresult[5] - result[3]*sx_zSum - result[4]*sy_zSum
	trueR22 := midresult[8] - result[6]*sx_zSum - result[7]*sy_zSum

	zScaleFactor := float32(1.0/math.Sqrt(float64(trueR02*trueR02 + trueR12*trueR12 + trueR22*trueR22)))
	result[2] = trueR02 * zScaleFactor
	result[5] = trueR12 * zScaleFactor
	result[8] = trueR22 * zScaleFactor

	return result
}

var specialRows = [6][3]float32{
    [3]float32{1,0,0},
    [3]float32{0,1,0},
    [3]float32{0,0,1},
    [3]float32{-1,0,0},
    [3]float32{0,-1,0},
    [3]float32{0,0,-1},
}

func lookupRotMatrix(special uint64) [9]float32 {
    specialRowDiv6 := specialRows[special/6]
    specialRowMod6 := specialRows[special%6]

    ret := [9]float32{
        0,0,0,
        specialRowMod6[0],specialRowMod6[1],specialRowMod6[2],
        specialRowDiv6[0],specialRowDiv6[1],specialRowDiv6[2],
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

func (b *ExtendedReader) ReadCFrame() (rbxfile.ValueCFrame, error) {
	var err error
	val := rbxfile.ValueCFrame{}
	val.Position, err = b.ReadVector3()
	if err != nil {
		return val, err
	}

	isLookup, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	if isLookup {
        special, err := b.Bits(6)
        if err != nil {
            return val, err
        }
        val.Rotation = lookupRotMatrix(special)
	} else {
        var matrix [4]float32
		matrix[3], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		matrix[0], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		matrix[1], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
		matrix[2], err = b.ReadFloat16BE(-1.0, 1.0)
		if err != nil {
			return val, err
		}
        val.Rotation = quaternionToRotMatrix(matrix)
	}

	return val, nil
}

func (b *ExtendedReader) ReadContent(isJoinData bool, context *CommunicationContext) (rbxfile.ValueContent, error) {
	if !isJoinData {
		val, err := b.ReadCachedContent(context)
		return rbxfile.ValueContent(val), err
	}
	var result string
	stringLen, err := b.ReadUint32BE()
	if err != nil {
		return rbxfile.ValueContent(result), err
	}
	result, err = b.ReadASCII(int(stringLen))
	return rbxfile.ValueContent(result), err
}

func (b *ExtendedReader) ReadSystemAddress(isJoinData bool, context *CommunicationContext) (rbxfile.ValueSystemAddress, error) {
	thisAddress := rbxfile.ValueSystemAddress("0.0.0.0:0")
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
			result, ok := context.ReplicatorSystemAddressCache.Get(cacheIndex)
			if !ok {
                return thisAddress, nil
			}
			return result.(rbxfile.ValueSystemAddress), nil
		}
	}
    thisAddr := net.UDPAddr{}
	thisAddr.IP = make([]byte, 4)
	err = b.Bytes(thisAddr.IP, 4)
	if err != nil {
		return thisAddress, err
	}
	port, err := b.ReadUint16BE()
	thisAddr.Port = int(port)
	if err != nil {
		return thisAddress, err
	}

	if !isJoinData {
		context.ReplicatorSystemAddressCache.Put(thisAddress, cacheIndex - 0x80)
	}

	return rbxfile.ValueSystemAddress(thisAddr.String()), nil
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

func (b *ExtendedReader) ReadNewPString(isJoinData bool, context *CommunicationContext) (rbxfile.ValueString, error) {
	if !isJoinData {
		val, err := b.ReadCached(context)
		return rbxfile.ValueString(val), err
	}
	stringLen, err := b.ReadUintUTF8()
	if err != nil {
		return rbxfile.ValueString(""), err
	}
	val, err := b.ReadASCII(int(stringLen))
	return rbxfile.ValueString(val), err
}
func (b *ExtendedReader) ReadNewProtectedString(isJoinData bool, context *CommunicationContext) (rbxfile.ValueProtectedString, error) {
	if !isJoinData {
		res, err := b.ReadNewCachedProtectedString(context)
		return rbxfile.ValueProtectedString(res), err
	}
	res, err := b.ReadNewPString(true, context)
	return rbxfile.ValueProtectedString(res), err
}
func (b *ExtendedReader) ReadNewContent(isJoinData bool, context *CommunicationContext) (rbxfile.ValueContent, error) {
	if !isJoinData {
		res, err := b.ReadCachedContent(context)
		return rbxfile.ValueContent(res), err
	}
	res, err := b.ReadNewPString(true, context)
	return rbxfile.ValueContent(res), err
}
func (b *ExtendedReader) ReadNewBinaryString() (rbxfile.ValueBinaryString, error) {
	res, err := b.ReadNewPString(true, nil)
	return rbxfile.ValueBinaryString(res), err
}

func (b *ExtendedReader) ReadNewEnumValue(enumID uint16, context *CommunicationContext) (rbxfile.ValueToken, error) {
	val, err := b.ReadUintUTF8()
	token := rbxfile.ValueToken{
		Value: val,
		ID: enumID,
		Name: getEnumName(context, enumID),
	}
	return token, err
}

func (b *ExtendedReader) ReadNewPSint() (rbxfile.ValueInt, error) {
	val, err := b.ReadSintUTF8()
	return rbxfile.ValueInt(val), err
}

func getEnumName(context *CommunicationContext, id uint16) string {
	return context.StaticSchema.Enums[id].Name
}

func (b *ExtendedReader) ReadNewTypeAndValue(isJoinData bool, context *CommunicationContext) (rbxfile.Value, error) {
    var val rbxfile.Value
	thisType, err := b.ReadUint8()
	if err != nil {
		return val, err
	}

	var enumID uint16
	if thisType == 7 {
		enumID, err = b.ReadUint16BE()
		if err != nil {
			return val, err
		}
	}

	val, err = readSerializedValue(isJoinData, enumID, thisType, b, context)
	return val, err
}

func (b *ExtendedReader) ReadNewTuple(isJoinData bool, context *CommunicationContext) (rbxfile.ValueTuple, error) {
	var tuple rbxfile.ValueTuple
	tupleLen, err := b.ReadUintUTF8()
	if err != nil {
		return tuple, err
	}
	if tupleLen > 0x10000 {
		return tuple, errors.New("sanity check: exceeded maximum tuple len")
	}
	tuple = make(rbxfile.ValueTuple, tupleLen)
	for i := 0; i < int(tupleLen); i++ {
		val, err := b.ReadNewTypeAndValue(isJoinData, context)
		if err != nil {
			return tuple, err
		}
		tuple[i] = val
	}

	return tuple, nil
}

func (b *ExtendedReader) ReadNewArray(isJoinData bool, context *CommunicationContext) (rbxfile.ValueArray, error) {
	array, err := b.ReadNewTuple(isJoinData, context)
	return rbxfile.ValueArray(array), err
}

func (b *ExtendedReader) ReadNewDictionary(isJoinData bool, context *CommunicationContext) (rbxfile.ValueDictionary, error) {
	var dictionary rbxfile.ValueDictionary
	dictionaryLen, err := b.ReadUintUTF8()
	if err != nil {
		return dictionary, err
	}
	if dictionaryLen > 0x10000 {
		return dictionary, errors.New("sanity check: exceeded maximum dictionary len")
	}
	dictionary = make(rbxfile.ValueDictionary, dictionaryLen)
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

func (b *ExtendedReader) ReadNewMap(isJoinData bool, context *CommunicationContext) (rbxfile.ValueMap, error) {
	thisMap, err := b.ReadNewDictionary(isJoinData, context)
	return rbxfile.ValueMap(thisMap), err
}

func (b *ExtendedReader) ReadNumberSequenceKeypoint() (rbxfile.ValueNumberSequenceKeypoint, error) {
	var err error
	thisKeypoint := rbxfile.ValueNumberSequenceKeypoint{}
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

func (b *ExtendedReader) ReadNumberSequence() (rbxfile.ValueNumberSequence, error) {
	var err error
	numKeypoints, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum numberseq len")
	}
	thisSequence := make(rbxfile.ValueNumberSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.ReadNumberSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *ExtendedReader) ReadNumberRange() (rbxfile.ValueNumberRange, error) {
	thisRange := rbxfile.ValueNumberRange{}
	var err error
	thisRange.Min, err = b.ReadFloat32BE()
	if err != nil {
		return thisRange, err
	}
	thisRange.Max, err = b.ReadFloat32BE()
	return thisRange, err
}

func (b *ExtendedReader) ReadColorSequenceKeypoint() (rbxfile.ValueColorSequenceKeypoint, error) {
	var err error
	thisKeypoint := rbxfile.ValueColorSequenceKeypoint{}
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

func (b *ExtendedReader) ReadColorSequence() (rbxfile.ValueColorSequence, error) {
	var err error
	numKeypoints, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}
	if numKeypoints > 0x10000 {
		return nil, errors.New("sanity check: exceeded maximum colorseq len")
	}
	thisSequence := make(rbxfile.ValueColorSequence, numKeypoints)

	for i := 0; i < int(numKeypoints); i++ {
		thisSequence[i], err = b.ReadColorSequenceKeypoint()
		if err != nil {
			return thisSequence, err
		}
	}

	return thisSequence, nil
}

func (b *ExtendedReader) ReadRect2D() (rbxfile.ValueRect2D, error) {
	var err error
	thisRect := rbxfile.ValueRect2D{rbxfile.ValueVector2{}, rbxfile.ValueVector2{}}

	thisRect.Min.X, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Min.Y, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Max.X, err = b.ReadFloat32BE()
	if err != nil {
		return thisRect, err
	}
	thisRect.Max.Y, err = b.ReadFloat32BE()
	return thisRect, err
}

func (b *ExtendedReader) ReadPhysicalProperties() (rbxfile.ValuePhysicalProperties, error) {
	var err error
	props := rbxfile.ValuePhysicalProperties{}
	props.CustomPhysics, err = b.ReadBool()
	if props.CustomPhysics {
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

func (b *ExtendedReader) ReadCoordsMode0() (rbxfile.ValueVector3, error) {
	return b.ReadVector3Simple()
}
func (b *ExtendedReader) ReadCoordsMode1() (rbxfile.ValueVector3, error) {
	value := rbxfile.ValueVector3{}
	cRange, err := b.ReadFloat32BE()
	if err != nil {
		return value, err
	}
	if cRange <= 0.0000099999997 { // Has to be precise
		return rbxfile.ValueVector3{0,0,0}, nil
	}
	x, err := b.ReadUint16BE()
	if err != nil {
		return value, err
	}
	y, err := b.ReadUint16BE()
	if err != nil {
		return value, err
	}
	z, err := b.ReadUint16BE()
	if err != nil {
		return value, err
	}
	value.X = (float32(x) / 32767.0 - 1.0) * cRange
	value.Y = (float32(y) / 32767.0 - 1.0) * cRange
	value.Z = (float32(z) / 32767.0 - 1.0) * cRange
	return value, nil
}
func (b *ExtendedReader) ReadCoordsMode2() (rbxfile.ValueVector3, error) {
	val := rbxfile.ValueVector3{}
	x, err := b.Bits(15)
	if err != nil {
		return val, err
	}
	x_short := uint16((x >> 7) | ((x & 0x7F) << 8))
	y, err := b.Bits(14)
	if err != nil {
		return val, err
	}
	y_short := uint16((y >> 6) | ((y & 0x3F) >> 8))
	z, err := b.Bits(15)
	if err != nil {
		return val, err
	}
	z_short := uint16((z >> 7) | ((z & 0x7F) << 8))
	
	val.X = float32(x_short) * 0.0625 - 1024.0
	val.Y = float32(y_short) * 0.0625 - 512.0
	val.Z = float32(z_short) * 0.0625 - 1024.0
	return val, nil
}

func (b *ExtendedReader) ReadPhysicsCoords() (rbxfile.ValueVector3, error) {
	var val rbxfile.ValueVector3
	mode, err := b.Bits(2)
	if err != nil {
		return val, err
	}

	switch mode {
	case 0:
		return b.ReadCoordsMode0()
	case 1:
		return b.ReadCoordsMode1()
	case 2:
		return b.ReadCoordsMode2()
	}
	return val, err
}

func (b *ExtendedReader) ReadMatrixMode0() ([9]float32, error) {
	var val [9]float32
	var err error

	q := [4]float32{}
	q[3], err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	q[0], err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	q[1], err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}
	q[2], err = b.ReadFloat32BE()
	if err != nil {
		return val, err
	}

	return quaternionToRotMatrix(q), nil
}

func (b *ExtendedReader) ReadMatrixMode1() ([9]float32, error) {
	q := [4]float32{}
	var val [9]float32
	invertW, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	invertX, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	invertY, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	invertZ, err := b.ReadBool()
	if err != nil {
		return val, err
	}
	x, err := b.ReadUint16LE()
	if err != nil {
		return val, err
	}
	y, err := b.ReadUint16LE()
	if err != nil {
		return val, err
	}
	z, err := b.ReadUint16LE()
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
	w := float32(math.Sqrt(math.Max(0.0, float64(1.0 - xs - ys - zs))))
	if invertW {
		w = -w
	}
	q = [4]float32{xs, ys, zs, w}
	return quaternionToRotMatrix(q), nil
}
func (b *ExtendedReader) ReadMatrixMode2() ([9]float32, error) {
	return b.ReadMatrixMode1()
}

func (b *ExtendedReader) ReadPhysicsMatrix() ([9]float32, error) {
	var val [9]float32
	mode, err := b.Bits(2)
	if err != nil {
		return val, err
	}

	switch mode {
	case 0:
		return b.ReadMatrixMode0()
	case 1:
		return b.ReadMatrixMode1()
	case 2:
		return b.ReadMatrixMode2()
	}
	return val, err
}

func (b *ExtendedReader) ReadPhysicsCFrame() (rbxfile.ValueCFrame, error) {
	var val rbxfile.ValueCFrame
	coords, err := b.ReadPhysicsCoords()
	if err != nil {
		return val, err
	}
	matrix, err := b.ReadPhysicsMatrix()
	if err != nil {
		return val, err
	}
	return rbxfile.ValueCFrame{coords, matrix}, nil
}

func (b *ExtendedReader) ReadPhysicsMotor() (PhysicsMotor, error) {
	var motor PhysicsMotor
	hasCoords, err := b.ReadBool()
	if err != nil {
		return motor, err
	}
	if hasCoords {
		hasCoords1, err := b.ReadBool()
		hasCoords2, err := b.ReadBool()
		if hasCoords1 {
			motor.Coords1, err = b.ReadPhysicsCoords()
			if err != nil {
				return motor, err
			}
		}
		if hasCoords2 {
			motor.Coords2, err = b.ReadCoordsMode1()
			if err != nil {
				return motor, err
			}
		} else {
			return motor, nil
		}
	}
	motor.Angle, err = b.ReadByte()
	return motor, err
}

func (b *ExtendedReader) ReadMotors() ([]PhysicsMotor, error) {
	countMotors, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	motors := make([]PhysicsMotor, countMotors)
	var i uint8
	for i = 0; i < countMotors; i++ {
		motors[i], err = b.ReadPhysicsMotor()
		if err != nil {
			return motors, err
		}
	}
	return motors, nil
}
