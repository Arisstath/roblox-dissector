package peer

import "net"
import "errors"
import "github.com/gskartwii/rbxfile"

func (b *BitstreamReader) readUDim() (rbxfile.ValueUDim, error) {
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

func (b *BitstreamReader) readUDim2() (rbxfile.ValueUDim2, error) {
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

func (b *BitstreamReader) readRay() (rbxfile.ValueRay, error) {
	var err error
	val := rbxfile.ValueRay{}
	val.Origin, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}
	val.Direction, err = b.readVector3Simple()
	return val, err
}

func (b *BitstreamReader) readRegion3() (rbxfile.ValueRegion3, error) {
	var err error
	val := rbxfile.ValueRegion3{}
	val.Start, err = b.readVector3Simple()
	if err != nil {
		return val, err
	}
	val.End, err = b.readVector3Simple()
	return val, err
}
func (b *BitstreamReader) readRegion3int16() (rbxfile.ValueRegion3int16, error) {
	var err error
	val := rbxfile.ValueRegion3int16{}
	val.Start, err = b.readVector3int16()
	if err != nil {
		return val, err
	}
	val.End, err = b.readVector3int16()
	return val, err
}

func (b *BitstreamReader) readColor3() (rbxfile.ValueColor3, error) {
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

func (b *BitstreamReader) readColor3uint8() (rbxfile.ValueColor3uint8, error) {
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

func (b *BitstreamReader) readVector2() (rbxfile.ValueVector2, error) {
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
func (b *BitstreamReader) readVector3Simple() (rbxfile.ValueVector3, error) {
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
func (b *BitstreamReader) readVector3() (rbxfile.ValueVector3, error) {
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

func (b *BitstreamReader) readVector2int16() (rbxfile.ValueVector2int16, error) {
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

func (b *BitstreamReader) readVector3int16() (rbxfile.ValueVector3int16, error) {
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

func (b *BitstreamReader) readAxes() (rbxfile.ValueAxes, error) {
	val, err := b.readUint32BE()
	axesVal := rbxfile.ValueAxes{
		X: val&4 != 0,
		Y: val&2 != 0,
		Z: val&1 != 0,
	}
	return axesVal, err
}
func (b *BitstreamReader) readFaces() (rbxfile.ValueFaces, error) {
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
func (b *BitstreamReader) readBrickColor() (rbxfile.ValueBrickColor, error) {
	val, err := b.readUint16BE()
	return rbxfile.ValueBrickColor(val), err
}

func ConstructReferent(scope string, id uint32) *Referent {
	if referentInt == 0 {
        return &Referent{IsNull: true}
	}
    return &Referent{Scope: scope, Id: id}
}

func (b *BitstreamReader) readJoinObject(context *CommunicationContext) (Reference, error) {
	referent, referentInt, err := b.readJoinReferent(context)
	serialized := ConstructReferent(referent, referentInt)

	return Referent(serialized), err
}
func (b *BitstreamReader) readObject(caches *Caches) (Reference, error) {
	var referentInt uint32
	referent, err := b.readCachedScope(caches)
	if err != nil && err != CacheReadOOB { // TODO: hack! physics packets may have problems with caches
		return "", err
	}
	if referent != "NULL" {
		referentInt, err = b.readUint32LE()
	}

	serialized := objectToRef(referent, referentInt)

	return Referent(serialized), err
}

// TODO: Make this function uniform with other cache functions
func (b *BitstreamReader) readSystemAddress(caches *Caches) (rbxfile.ValueSystemAddress, error) {
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

func (b *BitstreamReader) readNewEnumValue(enumID uint16) (rbxfile.ValueToken, error) {
	val, err := b.readUintUTF8()
	token := rbxfile.ValueToken{
		Value: val,
		ID:    enumID,
	} // Lazy-load token name!
	return token, err
}

func (b *BitstreamReader) readNewTuple(reader PacketReader) (rbxfile.ValueTuple, error) {
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

func (b *BitstreamReader) readNewArray(reader PacketReader) (rbxfile.ValueArray, error) {
	array, err := b.readNewTuple(reader)
	return rbxfile.ValueArray(array), err
}

func (b *BitstreamReader) readNewDictionary(reader PacketReader) (rbxfile.ValueDictionary, error) {
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

func (b *BitstreamReader) readNewMap(reader PacketReader) (rbxfile.ValueMap, error) {
	thisMap, err := b.readNewDictionary(reader)
	return rbxfile.ValueMap(thisMap), err
}

func (b *BitstreamReader) readNumberSequenceKeypoint() (rbxfile.ValueNumberSequenceKeypoint, error) {
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

func (b *BitstreamReader) readNumberSequence() (rbxfile.ValueNumberSequence, error) {
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

func (b *BitstreamReader) readNumberRange() (rbxfile.ValueNumberRange, error) {
	thisRange := rbxfile.ValueNumberRange{}
	var err error
	thisRange.Min, err = b.readFloat32BE()
	if err != nil {
		return thisRange, err
	}
	thisRange.Max, err = b.readFloat32BE()
	return thisRange, err
}

func (b *BitstreamReader) readColorSequenceKeypoint() (rbxfile.ValueColorSequenceKeypoint, error) {
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

func (b *BitstreamReader) readColorSequence() (rbxfile.ValueColorSequence, error) {
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

func (b *BitstreamReader) readRect2D() (rbxfile.ValueRect2D, error) {
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

func (b *BitstreamReader) readPhysicalProperties() (rbxfile.ValuePhysicalProperties, error) {
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

